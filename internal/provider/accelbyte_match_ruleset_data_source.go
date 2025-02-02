// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2client/rule_sets"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/match2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccelByteMatchRuleSetDataSource{}

func NewAccelByteMatchRuleSetDataSource() datasource.DataSource {
	return &AccelByteMatchRuleSetDataSource{}
}

// AccelByteMatchRuleSetDataSource defines the data source implementation.
type AccelByteMatchRuleSetDataSource struct {
	client *match2.RuleSetsService
}

func (d *AccelByteMatchRuleSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_ruleset"
}

func (d *AccelByteMatchRuleSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByteMatchRuleSet data source",

		Attributes: map[string]schema.Attribute{

			// Populated by user

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the match ruleset",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of match ruleset",
				Required:            true,
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Match ruleset identifier",
				Computed:            true,
			},

			// Fetched from AccelByte API during Read() opearation

			"enable_custom_match_function": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},

			"configuration": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
		},
	}
}

func (d *AccelByteMatchRuleSetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.RuleSetsService
}

func (d *AccelByteMatchRuleSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccelByteMatchRuleSetModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeMatchRuleSetId(data.Namespace.ValueString(), data.Name.ValueString()))

	input := &rule_sets.RuleSetDetailsParams{
		Namespace: data.Namespace.ValueString(),
		Ruleset:   data.Name.ValueString(),
	}

	matchRuleSet, err := d.client.RuleSetDetailsShort(input)

	if err != nil {
		// TODO: once the AccelByte SDK introduces rule_sets.RuleSetDetailsNotFound, we should use the following logic to detect API "not found" errors:
		// notFoundError := &rule_sets.RuleSetDetailsNotFound{}
		// if errors.As(err, &notFoundError) {
		if strings.Contains(err.Error(), "error 404:") {
			// The data source does not exist in the AccelByte backend
			// This is an actual error; do not update Terraform state, and signal an error to Terraform
			resp.Diagnostics.AddError("Data source not found", fmt.Sprintf("Match ruleset '%s' does not exist in namespace '%s'", input.Ruleset, input.Namespace))
			return
		} else {
			// Failed to retrieve the data source from the AccelByte backend
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

	diags, err := updateFromApiMatchRuleSet(ctx, &data, matchRuleSet)
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating our internal state to match match ruleset", fmt.Sprintf("Error: %#v", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeMatchRuleSetId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
