// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

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

func (d *AccelByteMatchPoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_pool"
}

func (d *AccelByteMatchPoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "This data source represents a [match pool](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-pools/).",

		Attributes: map[string]schema.Attribute{

			// Populated by user

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the match pool. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of match pool. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.",
				Required:            true,
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Match pool identifier, on the format `{{namespace}}/{{name}}`.",
				Computed:            true,
			},

			// Fetched from AccelByte API during Read() opearation

			// Basic information
			"rule_set": schema.StringAttribute{
				MarkdownDescription: "Match ruleset to use for this match pool. This defines the rules that will be used during matchmaking.",
				Computed:            true,
			},
			"session_template": schema.StringAttribute{
				MarkdownDescription: "Session template to usew for this match pool. This defines the characteristics of the session, such as joinability, what game server deploymewnt to use, and which regions it should deploy to.",
				Computed:            true,
			},
			"ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "Duration of a matchmaking request, in seconds. If matchmaking has not found a suitable match within this time, the matchmaking attempt will be aborted.",
				Computed:            true,
			},

			// Best latency calculation method
			"best_latency_calculation_method": schema.StringAttribute{
				MarkdownDescription: "Latency calculation used during matchmaking:\n\n" +
					"`Average` (Default): Matches players based on the average latency across all participants.\n" +
					"`P95`: Matches players based on the 95th percentile latency, aiming to minimize the worst-case latency experienced by the majority of players.",
				Computed: true,
			},

			// Backfill
			"auto_accept_backfill_proposal": schema.BoolAttribute{
				MarkdownDescription: "If set, allow AGS Matchmaking to handle backfill requests.",
				Computed:            true,
			},
			"backfill_proposal_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "Duration of a matchmaking proposal ticket, in seconds.",
				Computed:            true,
			},
			"backfill_ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "Duration of a backfill ticket, in seconds.",
				Computed:            true,
			},

			// Customization
			"match_function": schema.StringAttribute{
				MarkdownDescription: "Name of an Extend Override app. If set to `default`, no app will be called. Otherwise, this app will be invoked for [all overridable calls during matchmaking](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/overridable-matchmakingv2/).",
				Computed:            true,
			},
			"match_function_override": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"backfill_matches": schema.StringAttribute{
						MarkdownDescription: "Name of an Extend Override app. If set, this app will have the `BackfillMatches` RPC called to override implementation of matching tickets from queue that handles backfill tickets. This is called before the `MakeMatches` RPC is called.",
						Computed:            true,
					},
					"enrichment": schema.ListAttribute{
						MarkdownDescription: "Ordered list Extend Override apps. If set, these apps will have the `EnrichTicket` RPC called to add additional values to ticket attributes, e.g., insert values from external sources. This method is called after the ticket is hydrated.",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"make_matches": schema.StringAttribute{
						MarkdownDescription: "Name of an Extend Override app. If set, this app will have the `MakeMatches` RPC called to override the implementation of matching tickets from the queue for new tickets. This method is called periodically and takes tickets in a queue as an input. The interval can be configurable with the default time: minimum of 10 seconds and maximum of 30 seconds.",
						Computed:            true,
					},
					"stat_codes": schema.ListAttribute{
						MarkdownDescription: "Ordered list of Extend Override apps. If set, these apps will have the `GetStatCodes` RPC called to override or extend the list of codes and values added to player attributes in the match ticket. This method is called when the match ticket is hydrated.",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"validation": schema.ListAttribute{
						MarkdownDescription: "Ordered list of Extend Override apps. If set, these apps will have the `ValidateTicket` RPC called to override or extend the logic for validating a ticket, e.g., checking if the ruleset is valid or not. This method is called after the ticket is hydrated and enriched.",
						ElementType:         types.StringType,
						Computed:            true,
					},
				},
				Optional: true,
				Computed: true,
			},

			// Matchmaking Preferences
			"crossplay_enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"platform_group_enabled": schema.BoolAttribute{
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
	var data AccelByteMatchPoolModel

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
		// TODO: once the AccelByte SDK introduces match_pools.MatchPoolDetailsNotFound, we should use the following logic to detect API "not found" errors:
		// notFoundError := &match_pools.MatchPoolDetailsNotFound{}
		// if errors.As(err, &notFoundError) {
		if strings.Contains(err.Error(), "error 404:") {
			// The data source does not exist in the AccelByte backend
			// This is an actual error; do not update Terraform state, and signal an error to Terraform
			resp.Diagnostics.AddError("Data source not found", fmt.Sprintf("Match pool '%s' does not exist in namespace '%s'", input.Pool, input.Namespace))
			return
		} else {
			// Failed to retrieve the data source from the AccelByte backend
			// This is an actual error; do not update Terraform state, and signal an error to Terraform
			resp.Diagnostics.AddError("Error when reading match pool via AccelByte API", fmt.Sprintf("Unable to read match pool '%s' in namespace '%s', got error: %s", input.Pool, input.Namespace, err))
			return
		}
	}

	tflog.Trace(ctx, "Read match pool from AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
		"pool":      pool,
	})

	updateDiags, err := updateFromApiMatchPool(ctx, &data, pool)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match pool model according to AccelByte API response", fmt.Sprintf("Unable to process API response for match pool '%s' in namespace '%s' into model, got error: %s", input.Pool, input.Namespace, err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeMatchPoolId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
