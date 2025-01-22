// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/session"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/configuration_template"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AccelByteConfigurationTemplateResource{}
var _ resource.ResourceWithImportState = &AccelByteConfigurationTemplateResource{}

func NewAccelByteConfigurationTemplateResource() resource.Resource {
	return &AccelByteConfigurationTemplateResource{}
}

// AccelByteConfigurationTemplateResource defines the resource implementation.
type AccelByteConfigurationTemplateResource struct {
	client *session.ConfigurationTemplateService
}

func (r *AccelByteConfigurationTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration_template"
}

func (r *AccelByteConfigurationTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByte Configuration Template resource",

		Attributes: map[string]schema.Attribute{

			// Must be set by user; the ID is derived from these

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the configuration template",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of configuration template",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Configuration template identifier",
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

			// Can be set by user during resource creation; will otherwise get defaults from API
		},
	}
}

func (r *AccelByteConfigurationTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccelByteConfigurationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccelByteConfigurationTemplateModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeConfigurationTemplateId(data.Namespace.ValueString(), data.Name.ValueString()))

	input := &configuration_template.AdminCreateConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Body:      toApiConfigurationTemplate(data),
	}

	configurationTemplate, err := r.client.AdminCreateConfigurationTemplateV1Short(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to create new AccelByte configuration template in namespace '%s', name '%s', got error: %s", input.Namespace, input.Body.Name, err))
		return
	}

	updateFromApiConfigurationTemplate(&data, configurationTemplate)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Created an AccelByteConfigurationTemplateResource", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteConfigurationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccelByteConfigurationTemplateModel

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
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to read info on AccelByte configuration template from namespace '%s' name '%s', got error: %s", input.Namespace, input.Name, err))
		return
	}

	updateFromApiConfigurationTemplate(&data, configTemplate)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Read AccelByteConfigurationTemplateResource from AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteConfigurationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccelByteConfigurationTemplateModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := &configuration_template.AdminUpdateConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
		Body:      toApiConfigurationTemplateConfig(data),
	}

	apiConfigurationTemplate, err := r.client.AdminUpdateConfigurationTemplateV1Short(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to update new AccelByte configuration template in namespace '%s', name '%s', got error: %s", input.Namespace, input.Name, err))
		return
	}

	updateFromApiConfigurationTemplate(&data, apiConfigurationTemplate)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccelByteConfigurationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccelByteConfigurationTemplateModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := &configuration_template.AdminDeleteConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
	}
	err := r.client.AdminDeleteConfigurationTemplateV1Short(input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to delete AccelByte configuration template in namespace '%s', name '%s', got error: %s", input.Namespace, input.Name, err))
		return
	}
}

func (r *AccelByteConfigurationTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
