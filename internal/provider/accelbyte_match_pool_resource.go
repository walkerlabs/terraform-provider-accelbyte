// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2client/match_pools"
	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2clientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/match2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccelByteMatchPoolResource{}
var _ resource.ResourceWithImportState = &AccelByteMatchPoolResource{}

func NewAccelByteMatchPoolResource() resource.Resource {
	return &AccelByteMatchPoolResource{}
}

// AccelByteMatchPoolResource defines the resource implementation.
type AccelByteMatchPoolResource struct {
	client *match2.MatchPoolsService
}

// AccelByteMatchPoolResourceModel describes the resource data model.
type AccelByteMatchPoolResourceModel struct {
	// Must be set by user; the ID is derived from these
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Create() operation
	Id types.String `tfsdk:"id"`

	// Can be set or left as default by user

	AutoAcceptBackfillProposal        types.Bool   `tfsdk:"auto_accept_backfill_proposal"`
	BackfillProposalExpirationSeconds types.Int32  `tfsdk:"backfill_proposal_expiration_seconds"`
	BackfillTicketExpirationSeconds   types.Int32  `tfsdk:"backfill_ticket_expiration_seconds"`
	BestLatencyCalculationMethod      types.String `tfsdk:"best_latency_calculation_method"` // optional in AccelByte SDK
	CrossplayDisabled                 types.Bool   `tfsdk:"crossplay_disabled"`              // optional in AccelByte SDK
	MatchFunction                     types.String `tfsdk:"match_function"`
	// MatchFunctionOverride             types.Object `tfsdk:"match_function_override"` // This is a AccelByteMatchPoolMatchFunctionOverrideDataSourceModel
	PlatformGroupEnabled    types.Bool   `tfsdk:"platform_group_enabled"` // optional in AccelByte SDK
	RuleSet                 types.String `tfsdk:"rule_set"`
	SessionTemplate         types.String `tfsdk:"session_template"`
	TicketExpirationSeconds types.Int32  `tfsdk:"ticket_expiration_seconds"`
}

func (r *AccelByteMatchPoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_pool"
}

func (r *AccelByteMatchPoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByte Match Pool resource",

		Attributes: map[string]schema.Attribute{

			// Must be set by user; the ID is derived from these

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the match pool",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of match pool",
				Required:            true,
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Match pool identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Fetched from AccelByte API during Read() opearation

			"auto_accept_backfill_proposal": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"backfill_proposal_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"backfill_ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"best_latency_calculation_method": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"crossplay_disabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"match_function": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			// "match_function_override": schema.StringAttribute{
			// 	MarkdownDescription: "",
			// 	Computed:            true,
			// },
			"platform_group_enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"rule_set": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"session_template": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
			},
		},
	}
}

func (r *AccelByteMatchPoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = clients.Match2PoolsService
}

func (r *AccelByteMatchPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccelByteMatchPoolResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeMatchPoolId(data.Namespace.ValueString(), data.Name.ValueString()))

	input := &match_pools.CreateMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Body: &match2clientmodels.APIMatchPool{
			AutoAcceptBackfillProposal:        data.AutoAcceptBackfillProposal.ValueBoolPointer(),
			BackfillProposalExpirationSeconds: data.BackfillProposalExpirationSeconds.ValueInt32Pointer(),
			BackfillTicketExpirationSeconds:   data.BackfillTicketExpirationSeconds.ValueInt32Pointer(),
			BestLatencyCalculationMethod:      data.BestLatencyCalculationMethod.ValueString(),
			CrossplayDisabled:                 data.CrossplayDisabled.ValueBool(),
			MatchFunction:                     data.MatchFunction.ValueStringPointer(),
			//MatchFunctionOverride: data.MatchFunctionOverride.ValueInt32Pointer(),
			Name:                    data.Name.ValueStringPointer(),
			PlatformGroupEnabled:    data.PlatformGroupEnabled.ValueBool(),
			RuleSet:                 data.RuleSet.ValueStringPointer(),
			SessionTemplate:         data.SessionTemplate.ValueStringPointer(),
			TicketExpirationSeconds: data.TicketExpirationSeconds.ValueInt32Pointer(),
		},
	}

	err := r.client.CreateMatchPoolShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to create new AccelByte match pool in namespace '%s', name '%s', got error: %s", input.Namespace, input.Body.Name, err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Created an AccelByteMatchPoolDataSource", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccelByteMatchPoolResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := match_pools.MatchPoolDetailsParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
	}
	pool, err := r.client.MatchPoolDetailsShort(&input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to read info on AccelByte match pool from namespace '%s' name '%s', got error: %s", input.Namespace, input.Pool, err))
		return
	}

	data.AutoAcceptBackfillProposal = types.BoolValue(*pool.AutoAcceptBackfillProposal)
	data.BackfillProposalExpirationSeconds = types.Int32Value(*pool.BackfillProposalExpirationSeconds)
	data.BackfillTicketExpirationSeconds = types.Int32Value(*pool.BackfillTicketExpirationSeconds)
	data.BestLatencyCalculationMethod = types.StringValue(pool.BestLatencyCalculationMethod)
	data.CrossplayDisabled = types.BoolValue(pool.CrossplayDisabled)
	data.MatchFunction = types.StringValue(*pool.MatchFunction)
	// data.MatchFunctionOverride = types.ObjectValue(AccelByteMatchPoolMatchFunctionOverrideDataSourceModel{
	// 	BackfillMatches: pool.MatchFunctionOverride.BackfillMatches,
	// 	Enrichment:      []types.String{pool.MatchFunctionOverride.Enrichment},
	// 	MakeMatches:     pool.MatchFunctionOverride.MakeMatches,
	// 	StatCodes:       []types.String{pool.MatchFunctionOverride.StatCodes},
	// 	Validation:      []types.String{pool.MatchFunctionOverride.Validation},
	// })
	data.PlatformGroupEnabled = types.BoolValue(pool.PlatformGroupEnabled)
	data.RuleSet = types.StringValue(*pool.RuleSet)
	data.SessionTemplate = types.StringValue(*pool.SessionTemplate)
	data.TicketExpirationSeconds = types.Int32Value(*pool.TicketExpirationSeconds)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Read AccelByteMatchPoolDataSource from AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccelByteMatchPoolResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := &match_pools.UpdateMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
		Body: &match2clientmodels.APIMatchPoolConfig{
			AutoAcceptBackfillProposal:        data.AutoAcceptBackfillProposal.ValueBoolPointer(),
			BackfillProposalExpirationSeconds: data.BackfillProposalExpirationSeconds.ValueInt32Pointer(),
			BackfillTicketExpirationSeconds:   data.BackfillTicketExpirationSeconds.ValueInt32Pointer(),
			BestLatencyCalculationMethod:      data.BestLatencyCalculationMethod.ValueString(),
			CrossplayDisabled:                 data.CrossplayDisabled.ValueBool(),
			MatchFunction:                     data.MatchFunction.ValueStringPointer(),
			//MatchFunctionOverride: data.MatchFunctionOverride.ValueInt32Pointer(),
			PlatformGroupEnabled:    data.PlatformGroupEnabled.ValueBool(),
			RuleSet:                 data.RuleSet.ValueStringPointer(),
			SessionTemplate:         data.SessionTemplate.ValueStringPointer(),
			TicketExpirationSeconds: data.TicketExpirationSeconds.ValueInt32Pointer(),
		},
	}

	_, err := r.client.UpdateMatchPoolShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to update new AccelByte match pool in namespace '%s', name '%s', got error: %s", input.Namespace, input.Pool, err))
		return
	}

	// TODO: perhaps we should catch the response and update our state accordingly?

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccelByteMatchPoolResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := &match_pools.DeleteMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
	}
	err := r.client.DeleteMatchPoolShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to delete AccelByte match pool in namespace '%s', name '%s', got error: %s", input.Namespace, input.Pool, err))
		return
	}
}

func (r *AccelByteMatchPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
