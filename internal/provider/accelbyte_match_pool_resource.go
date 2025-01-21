// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2client/match_pools"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/match2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of match pool",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"backfill_proposal_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(300),
			},
			"backfill_ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(30),
			},
			"best_latency_calculation_method": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"crossplay_disabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"match_function": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
			},
			// "match_function_override": schema.StringAttribute{
			// 	MarkdownDescription: "",
			// 	Computed:            true,
			// },
			"platform_group_enabled": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"rule_set": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"session_template": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"ticket_expiration_seconds": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(300),
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

	input := &match_pools.CreateMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Body:      toApiMatchPool(data),
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
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to read info on AccelByte match pool from namespace '%s' name '%s', got error: %s", input.Namespace, input.Pool, err))
		return
	}

	updateFromApiMatchPool(data, pool)

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
	var data AccelByteMatchPoolModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := &match_pools.UpdateMatchPoolParams{
		Namespace: data.Namespace.ValueString(),
		Pool:      data.Name.ValueString(),
		Body:      toApiMatchPoolConfig(data),
	}

	apiMatchPool, err := r.client.UpdateMatchPoolShort(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to update new AccelByte match pool in namespace '%s', name '%s', got error: %s", input.Namespace, input.Pool, err))
		return
	}

	updateFromApiMatchPool(data, apiMatchPool)

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
