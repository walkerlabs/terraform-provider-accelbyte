// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2client/match_pools"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/match2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccelByteMatchPoolDataSource{}

func NewAccelByteMatchPoolDataSource() datasource.DataSource {
	return &AccelByteMatchPoolDataSource{}
}

// AccelByteMatchPoolDataSource defines the data source implementation.
type AccelByteMatchPoolDataSource struct {
	client *match2.MatchPoolsService
}

// AccelByteMatchPoolDataSourceModel describes the data source data model.
type AccelByteMatchPoolDataSourceModel struct {
	// Populated by user
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Read() operation
	Id types.String `tfsdk:"id"`

	// Fetched from AccelByte API during Read() opearation

	AutoAcceptBackfillProposal        types.Bool   `tfsdk:"auto_accept_backfill_proposal"`
	BackfillProposalExpirationSeconds types.Int32  `tfsdk:"backfill_proposal_expiration_seconds"`
	BackfillTicketExpirationSeconds   types.Int32  `tfsdk:"backfill_ticket_expiration_seconds"`
	BestLatencyCalculationMethod      types.String `tfsdk:"best_latency_calculation_method"` // optional
	CrossplayDisabled                 types.Bool   `tfsdk:"crossplay_disabled"`              // optional
	MatchFunction                     types.String `tfsdk:"match_function"`
	// MatchFunctionOverride             types.Object `tfsdk:"match_function_override"` // This is a AccelByteMatchPoolMatchFunctionOverrideDataSourceModel
	PlatformGroupEnabled    types.Bool   `tfsdk:"platform_group_enabled"` // optional
	RuleSet                 types.String `tfsdk:"rule_set"`
	SessionTemplate         types.String `tfsdk:"session_template"`
	TicketExpirationSeconds types.Int32  `tfsdk:"ticket_expiration_seconds"`
}

type AccelByteMatchPoolMatchFunctionOverrideDataSourceModel struct {
	BackfillMatches types.String   `tfsdk:"backfill_matches"` // optional
	Enrichment      []types.String `tfsdk:"enrichment"`       // optional
	MakeMatches     types.String   `tfsdk:"make_matches"`     // optional
	StatCodes       []types.String `tfsdk:"stat_codes"`       // optional
	Validation      []types.String `tfsdk:"validation"`       // optional
}

func (d *AccelByteMatchPoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_pool"
}

func (d *AccelByteMatchPoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByteMatchPool data source",

		Attributes: map[string]schema.Attribute{

			// Populated by user

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
			},

			// Fetched from AccelByte API during Read() opearation

			"auto_accept_backfill_proposal": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"backfill_proposal_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"backfill_ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"best_latency_calculation_method": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            true,
			},
			"crossplay_disabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            true,
			},
			"match_function": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			// "match_function_override": schema.StringAttribute{
			// 	MarkdownDescription: "",
			// 	Computed:            true,
			// },
			"platform_group_enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            true,
			},
			"rule_set": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"session_template": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
		},
	}
}

func (d *AccelByteMatchPoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.Match2PoolsService
}

func (d *AccelByteMatchPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccelByteMatchPoolDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeMatchPoolId(data.Namespace.ValueString(), data.Name.ValueString()))

	input := match_pools.MatchPoolDetailsParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
	}
	pool, err := d.client.MatchPoolDetailsShort(&input)
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeMatchPoolId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
