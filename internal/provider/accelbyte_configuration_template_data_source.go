// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/session"
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclient/configuration_template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccelByteConfigurationTemplateDataSource{}

func NewAccelByteConfigurationTemplateDataSource() datasource.DataSource {
	return &AccelByteConfigurationTemplateDataSource{}
}

// AccelByteConfigurationTemplateDataSource defines the data source implementation.
type AccelByteConfigurationTemplateDataSource struct {
	client *session.ConfigurationTemplateService
}

func (d *AccelByteConfigurationTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration_template"
}

func (d *AccelByteConfigurationTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByteConfigurationTemplate data source",

		Attributes: map[string]schema.Attribute{

			// Populated by user

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the configuration template",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of configuration template",
				Required:            true,
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Configuration template identifier",
				Computed:            true,
			},

			// Fetched from AccelByte API during Read() opearation

			// "General" screen - Main configuration
			"min_players": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"max_players": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},

			// "General" screen - Connection and Joinability
			"joinability": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},

			// "General" screen - Main configuration
			"max_active_sessions": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			// TODO: support "use Custom Session Function"

			// "General" screen - Connection and Joinability
			"invite_timeout": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"inactive_timeout": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"leader_election_grace_period": schema.Int32Attribute{
				MarkdownDescription: "",
				Computed:            true,
			},

			// "General" screen - Server
			"server_type": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			// Only used when ServerType = AMS
			"requested_regions": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "",
				Computed:            true,
			},
			"preferred_claim_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "",
				Computed:            true,
			},
			"fallback_claim_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "",
				Computed:            true,
			},
			// TODO: support ServerType = CUSTOM

			// "Additional" screen settings
			"auto_join_session": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"chat_room": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"secret_validation": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"generate_code": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"immutable_session_storage": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"manual_set_ready_for_ds": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"tied_teams_session_lifetime": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"auto_leave_session": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},

			// TODO: support "3rd party sync" options

			// "Custom Attributes" screen
			"custom_attributes": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
		},
	}
}

func (d *AccelByteConfigurationTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = clients.SessionConfigurationTemplateService
}

func (d *AccelByteConfigurationTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccelByteConfigurationTemplateModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeConfigurationTemplateId(data.Namespace.ValueString(), data.Name.ValueString()))

	input := configuration_template.AdminGetConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
	}
	configTemplate, err := d.client.AdminGetConfigurationTemplateV1Short(&input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to read info on AccelByte configuration template from namespace '%s' name '%s', got error: %s", input.Namespace, input.Name, err))
		return
	}

	err = updateFromApiConfigurationTemplate(ctx, &data, configTemplate)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating our internal state from the configuration template", fmt.Sprintf("Error: %#v", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Read AccelByteConfigurationTemplateDataSource from AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeConfigurationTemplateId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
