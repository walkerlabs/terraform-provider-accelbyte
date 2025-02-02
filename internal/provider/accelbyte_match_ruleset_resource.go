// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2client/rule_sets"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/match2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccelByteMatchRuleSetResource{}
var _ resource.ResourceWithImportState = &AccelByteMatchRuleSetResource{}

func NewAccelByteMatchRuleSetResource() resource.Resource {
	return &AccelByteMatchRuleSetResource{}
}

// AccelByteMatchRuleSetResource defines the resource implementation.
type AccelByteMatchRuleSetResource struct {
	client *match2.RuleSetsService
}

func (r *AccelByteMatchRuleSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_ruleset"
}

func (r *AccelByteMatchRuleSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByte Match Ruleset resource",

		Attributes: map[string]schema.Attribute{

			// Must be set by user; the ID is derived from these

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the match ruleset",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of match ruleset",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Match ruleset identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Can be set by user during resource creation; will otherwise get defaults from schema

			"enable_custom_match_function": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			// Must be set by user during resource creation

			"configuration": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
		},
	}
}

func (r *AccelByteMatchRuleSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*AccelByteProviderClients)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AccelByteProviderClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = clients.RuleSetsService
}

func (r *AccelByteMatchRuleSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccelByteMatchRuleSetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeMatchRuleSetId(data.Namespace.ValueString(), data.Name.ValueString()))

	apiMatchRuleSet, diags, err := toApiMatchRuleSet(ctx, data)
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when converting our internal state to an AccelByte API match ruleset", fmt.Sprintf("Error: %#v", err))
		return
	}

	tflog.Trace(ctx, "Creating match ruleset via AccelByte API", map[string]interface{}{
		"namespace":       data.Namespace,
		"name":            data.Name.ValueString(),
		"apiMatchRuleSet": apiMatchRuleSet,
	})

	input := &rule_sets.CreateRuleSetParams{
		Namespace: data.Namespace.ValueString(),
		Body:      apiMatchRuleSet,
	}

	err = r.client.CreateRuleSetShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when creating match ruleset via AccelByte API", fmt.Sprintf("Unable to create match ruleset '%s' in namespace '%s', got error: %s", *input.Body.Name, input.Namespace, err))
		return
	}

	readInput := &rule_sets.RuleSetDetailsParams{
		Namespace: data.Namespace.ValueString(),
		Ruleset:   data.Name.ValueString(),
	}

	matchRuleSet, err := r.client.RuleSetDetailsShort(readInput)
	if err != nil {
		resp.Diagnostics.AddError("Error when refreshing match ruleset via AccelByte API", fmt.Sprintf("Unable to read match ruleset '%s' in namespace '%s', got error: %s", readInput.Ruleset, readInput.Namespace, err))
		return
	}

	updateDiags, err := updateFromApiMatchRuleSet(ctx, &data, matchRuleSet)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match ruleset model according to AccelByte API response", fmt.Sprintf("Unable to process API response for ruleset '%s' in namespace '%s' into model, got error: %s", readInput.Ruleset, readInput.Namespace, err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchRuleSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccelByteMatchRuleSetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := &rule_sets.RuleSetDetailsParams{
		Namespace: data.Namespace.ValueString(),
		Ruleset:   data.Name.ValueString(),
	}

	matchRuleSet, err := r.client.RuleSetDetailsShort(input)

	if err != nil {
		// TODO: once the AccelByte SDK introduces rule_sets.RuleSetDetailsNotFound, we should use the following logic to detect API "not found" errors:
		// notFoundError := &rule_sets.RuleSetDetailsNotFound{}
		// if errors.As(err, &notFoundError) {
		if strings.Contains(err.Error(), "error 404:") {
			// The resource does not exist in the AccelByte backend
			// Ensure that it does not exist in the Terraform state either
			// This not an error condition; Terraform will proceed assuming that the resource does not exist in the backend
			resp.State.RemoveResource(ctx)
			return
		} else {
			// Failed to retrieve the resource from the AccelByte backend
			// This is an actual error; do not update Terraform state, and signal an error to Terraform
			resp.Diagnostics.AddError("Error when reading match ruleset via AccelByte API", fmt.Sprintf("Unable to read match ruleset '%s' in namespace '%s', got error: %s", input.Ruleset, input.Namespace, err))
			return
		}
	}

	tflog.Trace(ctx, "Read match ruleset from AccelByte API", map[string]interface{}{
		"namespace":    data.Namespace,
		"name":         data.Name.ValueString(),
		"matchRuleSet": matchRuleSet,
	})

	updateDiags, err := updateFromApiMatchRuleSet(ctx, &data, matchRuleSet)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match ruleset model according to AccelByte API response", fmt.Sprintf("Unable to process API response for ruleset '%s' in namespace '%s' into model, got error: %s", input.Ruleset, input.Namespace, err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchRuleSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccelByteMatchRuleSetModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiMatchRuleSet, diagnostics, err := toApiMatchRuleSet(ctx, data)
	resp.Diagnostics.Append(diagnostics...)
	if err != nil {
		resp.Diagnostics.AddError("Error when converting our internal state to an AccelByte API match ruleset", fmt.Sprintf("Error: %#v", err))
		return
	}

	tflog.Trace(ctx, "Updating match ruleset via AccelByte API", map[string]interface{}{
		"namespace":       data.Namespace,
		"name":            data.Name.ValueString(),
		"apiMatchRuleSet": apiMatchRuleSet,
	})

	input := &rule_sets.UpdateRuleSetParams{
		Namespace: data.Namespace.ValueString(),
		Ruleset:   data.Name.ValueString(),
		Body:      apiMatchRuleSet,
	}

	apiMatchRuleSet2, err := r.client.UpdateRuleSetShort(input)
	if err != nil {
		notFoundError := &rule_sets.UpdateRuleSetNotFound{}
		if errors.As(err, &notFoundError) {
			// The resource does not exist in the AccelByte backend
			// This means that the resource has disappeared since the TF state was refreshed at the start of the apply operation; we should abort
			resp.Diagnostics.AddError("Resource not found", fmt.Sprintf("Match ruleset '%s' does not exist in namespace '%s'", input.Ruleset, input.Namespace))
			return
		} else {
			// Failed to update the resource in the AccelByte backend
			// The backend refused our update operation; we should abort
			resp.Diagnostics.AddError("Error when updating match ruleset via AccelByte API", fmt.Sprintf("Unable to update match ruleset '%s' in namespace '%s', got error: %s", input.Ruleset, input.Namespace, err))
			return
		}
	}

	updateDiags, err := updateFromApiMatchRuleSet(ctx, &data, apiMatchRuleSet2)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match ruleset model according to AccelByte API response", fmt.Sprintf("Unable to process API response for ruleset '%s' in namespace '%s' into model, got error: %s", input.Ruleset, input.Namespace, err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchRuleSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccelByteMatchRuleSetModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleting match ruleset via AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	input := &rule_sets.DeleteRuleSetParams{
		Namespace: data.Namespace.ValueString(),
		Ruleset:   data.Name.ValueString(),
	}
	err := r.client.DeleteRuleSetShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when deleting ruleset via AccelByte API", fmt.Sprintf("Unable to ruleset template '%s' in namespace '%s', got error: %s", input.Ruleset, input.Namespace, err))
		return
	}
}

func (r *AccelByteMatchRuleSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
