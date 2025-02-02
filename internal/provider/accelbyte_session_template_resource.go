// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/session"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/configuration_template"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccelByteSessionTemplateResource{}
var _ resource.ResourceWithImportState = &AccelByteSessionTemplateResource{}

func NewAccelByteSessionTemplateResource() resource.Resource {
	return &AccelByteSessionTemplateResource{}
}

// AccelByteSessionTemplateResource defines the resource implementation.
type AccelByteSessionTemplateResource struct {
	client *session.ConfigurationTemplateService
}

func (r *AccelByteSessionTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_session_template"
}

func (r *AccelByteSessionTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByte Session Template resource",

		Attributes: map[string]schema.Attribute{

			// Must be set by user; the ID is derived from these

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the session template",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of session template",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Session template identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Must be set by user during resource creation

			"min_players": schema.Int32Attribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"max_players": schema.Int32Attribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"joinability": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},

			// Can be set by user during resource creation; will otherwise get defaults from schema

			// "General" screen - Main configuration
			"max_active_sessions": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(-1),
			},
			"custom_session_function": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"on_session_created": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						Validators: []validator.Bool{
							// At least one of the on_* bools must be set
							boolvalidator.AtLeastOneOf(path.Expressions{
								path.MatchRelative().AtParent().AtName("on_session_updated"),
								path.MatchRelative().AtParent().AtName("on_session_deleted"),
								path.MatchRelative().AtParent().AtName("on_party_created"),
								path.MatchRelative().AtParent().AtName("on_party_updated"),
								path.MatchRelative().AtParent().AtName("on_party_deleted"),
							}...),
						},
					},
					"on_session_updated": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"on_session_deleted": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"on_party_created": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"on_party_updated": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"on_party_deleted": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"custom_url": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						Validators: []validator.String{
							// Custom URL cannot be used at the same time as an Extend App
							stringvalidator.ExactlyOneOf(path.Expressions{
								path.MatchRelative().AtParent().AtName("extend_app"),
							}...),
						},
					},
					"extend_app": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
				},
				Optional: true,
				Computed: true,
			},

			// "General" screen - Connection and Joinability
			"invite_timeout": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(60),
			},
			"inactive_timeout": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(60),
			},
			"leader_election_grace_period": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(0),
			},

			// ServerType = NONE is implied when none of the other server types are specified in the configuration

			// Peer-to-Peer server
			"p2p_server": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{},
				Optional:   true,
				Computed:   true,
				Validators: []validator.Object{
					// P2P server configuration cannot coexist with an AMS or Custom server configuration
					objectvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("ams_server"),
						path.MatchRoot("custom_server"),
					}...),
				},
			},

			// AMS server
			"ams_server": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"requested_regions": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListValueMust(basetypes.StringType{}, []attr.Value{})),
					},
					"preferred_claim_keys": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListValueMust(basetypes.StringType{}, []attr.Value{})),
					},
					"fallback_claim_keys": schema.ListAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListValueMust(basetypes.StringType{}, []attr.Value{})),
					},
				},
				Optional: true,
				Computed: true,
			},

			// Custom server
			"custom_server": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"custom_url": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						Validators: []validator.String{
							// Custom URL cannot be used at the same time as an Extend App
							stringvalidator.ExactlyOneOf(path.Expressions{
								path.MatchRelative().AtParent().AtName("extend_app"),
							}...),
						},
					},
					"extend_app": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
					},
				},
				Optional: true,
				Computed: true,
			},

			// "Additional" screen settings
			"auto_join_session": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"chat_room": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"secret_validation": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"generate_code": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"immutable_session_storage": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"manual_set_ready_for_ds": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tied_teams_session_lifetime": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"auto_leave_session": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			// TODO: support "3rd party sync" options

			// "Custom Attributes" screen
			"custom_attributes": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
		},
	}
}

func (r *AccelByteSessionTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = clients.SessionConfigurationTemplateService
}

func (r *AccelByteSessionTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccelByteSessionTemplateModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeSessionTemplateId(data.Namespace.ValueString(), data.Name.ValueString()))

	apiSessionTemplate, diags, err := toApiSessionTemplate(ctx, data)
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when converting our internal state to an AccelByte API session template", fmt.Sprintf("Error: %#v", err))
		return
	}

	tflog.Trace(ctx, "Creating session template via AccelByte API", map[string]interface{}{
		"namespace":          data.Namespace,
		"name":               data.Name.ValueString(),
		"apiSessionTemplate": apiSessionTemplate,
	})

	input := &configuration_template.AdminCreateConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Body:      apiSessionTemplate,
	}

	configurationTemplate, err := r.client.AdminCreateConfigurationTemplateV1Short(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when creating session template via AccelByte API", fmt.Sprintf("Unable to create session template '%s' in namespace '%s', got error: %s", *input.Body.Name, input.Namespace, err))
		return
	}

	updateDiags, err := updateFromApiSessionTemplate(ctx, &data, configurationTemplate)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating session template model according to AccelByte API response", fmt.Sprintf("Unable to process API response for session template '%s' in namespace '%s' into model, got error: %s", *input.Body.Name, input.Namespace, err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteSessionTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccelByteSessionTemplateModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := configuration_template.AdminGetConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
	}
	configTemplate, err := r.client.AdminGetConfigurationTemplateV1Short(&input)
	if err != nil {
		notFoundError := &configuration_template.AdminGetConfigurationTemplateV1NotFound{}
		if errors.As(err, &notFoundError) {
			// The resource does not exist in the AccelByte backend
			// Ensure that it does not exist in the Terraform state either
			// This not an error condition; Terraform will proceed assuming that the resource does not exist in the backend
			resp.State.RemoveResource(ctx)
			return
		} else {
			// Failed to retrieve the resource from the AccelByte backend
			// This is an actual error; do not update Terraform state, and signal an error to Terraform
			resp.Diagnostics.AddError("Error when reading session template via AccelByte API", fmt.Sprintf("Unable to read session template '%s' in namespace '%s', got error: %s", input.Name, input.Namespace, err))
			return
		}
	}

	tflog.Trace(ctx, "Read session template from AccelByte API", map[string]interface{}{
		"namespace":      data.Namespace,
		"name":           data.Name.ValueString(),
		"configTemplate": configTemplate,
	})

	updateDiags, err := updateFromApiSessionTemplate(ctx, &data, configTemplate)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating session template model according to AccelByte API response", fmt.Sprintf("Unable to process API response for session template '%s' in namespace '%s' into model, got error: %s", input.Name, input.Namespace, err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteSessionTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccelByteSessionTemplateModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiSessionTemplateConfig, diagnostics, err := toApiSessionTemplateConfig(ctx, data)
	resp.Diagnostics.Append(diagnostics...)
	if err != nil {
		resp.Diagnostics.AddError("Error when converting our internal state to an AccelByte API session template config", fmt.Sprintf("Error: %#v", err))
		return
	}

	tflog.Trace(ctx, "Updating session template via AccelByte API", map[string]interface{}{
		"namespace":                data.Namespace,
		"name":                     data.Name.ValueString(),
		"apiSessionTemplateConfig": apiSessionTemplateConfig,
	})

	input := &configuration_template.AdminUpdateConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
		Body:      apiSessionTemplateConfig,
	}

	apiSessionTemplate, err := r.client.AdminUpdateConfigurationTemplateV1Short(input)
	if err != nil {
		notFoundError := &configuration_template.AdminUpdateConfigurationTemplateV1NotFound{}
		if errors.As(err, &notFoundError) {
			// The resource does not exist in the AccelByte backend
			// This means that the resource has disappeared since the TF state was refreshed at the start of the apply operation; we should abort
			resp.Diagnostics.AddError("Resource not found", fmt.Sprintf("Session template '%s' does not exist in namespace '%s'", input.Name, input.Namespace))
			return
		} else {
			// Failed to update the resource in the AccelByte backend
			// The backend refused our update operation; we should abort
			resp.Diagnostics.AddError("Error when updating session template via AccelByte API", fmt.Sprintf("Unable to update session template '%s' in namespace '%s', got error: %s", input.Name, input.Namespace, err))
			return
		}
	}

	updateDiags, err := updateFromApiSessionTemplate(ctx, &data, apiSessionTemplate)
	resp.Diagnostics.Append(updateDiags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating session template model according to AccelByte API response", fmt.Sprintf("Unable to process API response for session template '%s' in namespace '%s' into model, got error: %s", input.Name, input.Namespace, err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteSessionTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccelByteSessionTemplateModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Deleting session template via AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	input := &configuration_template.AdminDeleteConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
	}
	err := r.client.AdminDeleteConfigurationTemplateV1Short(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when deleting session template via AccelByte API", fmt.Sprintf("Unable to delete session template '%s' in namespace '%s', got error: %s", input.Name, input.Namespace, err))
		return
	}
}

func (r *AccelByteSessionTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
