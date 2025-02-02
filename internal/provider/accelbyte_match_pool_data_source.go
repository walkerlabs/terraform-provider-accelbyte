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

			// Basic information
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

			// Best latency calculation method
			"best_latency_calculation_method": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},

			// Backfill
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

			// Customization
			"match_function": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"match_function_override": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"backfill_matches": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"enrichment": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "",
						Computed:            true,
					},
					"make_matches": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"stat_codes": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "",
						Computed:            true,
					},
					"validation": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "",
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

	updateFromApiMatchPool(ctx, &data, pool)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeMatchPoolId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
