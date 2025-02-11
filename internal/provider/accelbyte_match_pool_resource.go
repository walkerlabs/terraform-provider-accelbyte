// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2client/match_pools"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/match2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccelByteMatchPoolResource{}
var _ resource.ResourceWithImportState = &AccelByteMatchPoolResource{}

const (
	// Wait this many seconds after any write operation to the AB API, in the hope that cached results are flushed out by then.
	CACHE_INVALIDATION_DELAY_SECONDS = 20
)

func NewAccelByteMatchPoolResource() resource.Resource {
	return &AccelByteMatchPoolResource{}
}

// AccelByteMatchPoolResource defines the resource implementation.
type AccelByteMatchPoolResource struct {
	client *match2.MatchPoolsService
}

func (r *AccelByteMatchPoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_match_pool"
}

func (r *AccelByteMatchPoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "This resource represents a [match pool](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-pools/).",

		Attributes: map[string]schema.Attribute{

			// Must be set by user; the ID is derived from these

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the match pool. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of match pool. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Match pool identifier, on the format `{{namespace}}/{{name}}`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Must be set by user during resource creation

			// Basic information
			"rule_set": schema.StringAttribute{
				MarkdownDescription: "Match ruleset to use for this match pool. This defines the rules that will be used during matchmaking.",
				Required:            true,
			},
			"session_template": schema.StringAttribute{
				MarkdownDescription: "Session template to usew for this match pool. This defines the characteristics of the session, such as joinability, what game server deploymewnt to use, and which regions it should deploy to.",
				Required:            true,
			},
			"ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "Duration of a matchmaking request, in seconds. If matchmaking has not found a suitable match within this time, the matchmaking attempt will be aborted.",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(300),
			},

			// Best latency calculation method
			"best_latency_calculation_method": schema.StringAttribute{
				MarkdownDescription: "Latency calculation used during matchmaking:\n\n" +
					"`Average` (Default): Matches players based on the average latency across all participants.\n" +
					"`P95`: Matches players based on the 95th percentile latency, aiming to minimize the worst-case latency experienced by the majority of players.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},

			// Backfill
			"auto_accept_backfill_proposal": schema.BoolAttribute{
				MarkdownDescription: "If set, allow AGS Matchmaking to handle backfill requests.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"backfill_proposal_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "Duration of a matchmaking proposal ticket, in seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(30),
			},
			"backfill_ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "Duration of a backfill ticket, in seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(300),
			},

			// Customization
			"match_function": schema.StringAttribute{
				MarkdownDescription: "Name of an Extend Override app. If set to `default`, no app will be called. Otherwise, this app will be invoked for [all overridable calls during matchmaking](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/overridable-matchmakingv2/).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
			},
			"match_function_override": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"backfill_matches": schema.StringAttribute{
						MarkdownDescription: "Name of an Extend Override app. If set, this app will have the `BackfillMatches` RPC called to override implementation of matching tickets from queue that handles backfill tickets. This is called before the `MakeMatches` RPC is called.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
					"enrichment": schema.ListAttribute{
						MarkdownDescription: "Ordered list Extend Override apps. If set, these apps will have the `EnrichTicket` RPC called to add additional values to ticket attributes, e.g., insert values from external sources. This method is called after the ticket is hydrated.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListValueMust(basetypes.StringType{}, []attr.Value{})),
					},
					"make_matches": schema.StringAttribute{
						MarkdownDescription: "Name of an Extend Override app. If set, this app will have the `MakeMatches` RPC called to override the implementation of matching tickets from the queue for new tickets. This method is called periodically and takes tickets in a queue as an input. The interval can be configurable with the default time: minimum of 10 seconds and maximum of 30 seconds.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
					"stat_codes": schema.ListAttribute{
						MarkdownDescription: "Ordered list of Extend Override apps. If set, these apps will have the `GetStatCodes` RPC called to override or extend the list of codes and values added to player attributes in the match ticket. This method is called when the match ticket is hydrated.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListValueMust(basetypes.StringType{}, []attr.Value{})),
					},
					"validation": schema.ListAttribute{
						MarkdownDescription: "Ordered list of Extend Override apps. If set, these apps will have the `ValidateTicket` RPC called to override or extend the logic for validating a ticket, e.g., checking if the ruleset is valid or not. This method is called after the ticket is hydrated and enriched.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListValueMust(basetypes.StringType{}, []attr.Value{})),
					},
				},
				Optional: true,
				Computed: true,
			},

			// Matchmaking Preferences
			"crossplay_enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"platform_group_enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
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
	var data AccelByteMatchPoolModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeMatchPoolId(data.Namespace.ValueString(), data.Name.ValueString()))

	// Create pool

	apiMatchPool, apiMatchPoolDiags := toApiMatchPool(ctx, data)
	resp.Diagnostics.Append(apiMatchPoolDiags...)

	tflog.Trace(ctx, "Creating match pool via AccelByte API", map[string]interface{}{
		"namespace":    data.Namespace,
		"name":         data.Name.ValueString(),
		"apiMatchPool": apiMatchPool,
	})

	createInput := &match_pools.CreateMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Body:      apiMatchPool,
	}

	err := r.client.CreateMatchPoolShort(createInput)
	if err != nil {
		resp.Diagnostics.AddError("Error when creating match pool via AccelByte API", fmt.Sprintf("Unable to create match pool '%s' in namespace '%s', got error: %s", *createInput.Body.Name, createInput.Namespace, err))
		return
	}

	time.Sleep(CACHE_INVALIDATION_DELAY_SECONDS * time.Second)

	// Fetch pool immediately after creating it, so we can get the values for un-set defaults

	readInput := match_pools.MatchPoolDetailsParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
	}
	pool, err := r.client.MatchPoolDetailsShort(&readInput)
	if err != nil {
		resp.Diagnostics.AddError("Error when reading match pool via AccelByte API", fmt.Sprintf("Unable to match pool template '%s' in namespace '%s', got error: %s", readInput.Pool, readInput.Namespace, err))
		return
	}

	tflog.Trace(ctx, "Read match pool from AccelByte API", map[string]interface{}{
		"namespace": readInput.Namespace,
		"name":      readInput.Pool,
		"pool":      pool,
	})

	// Reflect new pool from API into our model

	updateDiags, err := updateFromApiMatchPool(ctx, &data, pool)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match pool model according to AccelByte API response", fmt.Sprintf("Unable to process API response for match pool '%s' in namespace '%s' into model, got error: %s", readInput.Pool, readInput.Namespace, err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccelByteMatchPoolModel

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
		// TODO: once the AccelByte SDK introduces match_pools.MatchPoolDetailsNotFound, we should use the following logic to detect API "not found" errors:
		// notFoundError := &match_pools.MatchPoolDetailsNotFound{}
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
			resp.Diagnostics.AddError("Error when reading match pool via AccelByte API", fmt.Sprintf("Unable to read match template '%s' in namespace '%s', got error: %s", input.Pool, input.Namespace, err))
			return
		}
	}

	updateDiags, err := updateFromApiMatchPool(ctx, &data, pool)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match pool model according to AccelByte API response", fmt.Sprintf("Unable to process API response for match pool '%s' in namespace '%s' into model, got error: %s", input.Pool, input.Namespace, err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Read AccelByteMatchPoolResource from AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccelByteMatchPoolModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiMatchPoolConfig, apiMatchPoolConfigDiags := toApiMatchPoolConfig(ctx, data)
	resp.Diagnostics.Append(apiMatchPoolConfigDiags...)

	tflog.Trace(ctx, "Updating match pool via AccelByte API", map[string]interface{}{
		"namespace":          data.Namespace,
		"name":               data.Name.ValueString(),
		"apiMatchPoolConfig": apiMatchPoolConfig,
	})

	input := &match_pools.UpdateMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
		Body:      apiMatchPoolConfig,
	}

	apiMatchPool, err := r.client.UpdateMatchPoolShort(input)
	if err != nil {
		notFoundError := &match_pools.UpdateMatchPoolNotFound{}
		if errors.As(err, &notFoundError) {
			// The resource does not exist in the AccelByte backend
			// This means that the resource has disappeared since the TF state was refreshed at the start of the apply operation; we should abort
			resp.Diagnostics.AddError("Resource not found", fmt.Sprintf("Match pool '%s' does not exist in namespace '%s'", input.Pool, input.Namespace))
			return
		} else {
			// Failed to update the resource in the AccelByte backend
			// The backend refused our update operation; we should abort
			resp.Diagnostics.AddError("Error when updating match pool via AccelByte API", fmt.Sprintf("Unable to update match pool '%s' in namespace '%s', got error: %s", input.Pool, input.Namespace, err))
			return
		}
	}

	time.Sleep(CACHE_INVALIDATION_DELAY_SECONDS * time.Second)

	updateDiags, err := updateFromApiMatchPool(ctx, &data, apiMatchPool)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating match pool model according to AccelByte API response", fmt.Sprintf("Unable to process API response for match pool '%s' in namespace '%s' into model, got error: %s", input.Pool, input.Namespace, err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteMatchPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccelByteMatchPoolModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleting match pool via AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	input := &match_pools.DeleteMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
	}
	err := r.client.DeleteMatchPoolShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when deleting match pool via AccelByte API", fmt.Sprintf("Unable to delete match pool '%s' in namespace '%s', got error: %s", input.Pool, input.Namespace, err))
		return
	}

	time.Sleep(CACHE_INVALIDATION_DELAY_SECONDS * time.Second)
}

func (r *AccelByteMatchPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
