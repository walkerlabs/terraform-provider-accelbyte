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
var _ datasource.DataSource = &AccelByteSessionTemplateDataSource{}

func NewAccelByteSessionTemplateDataSource() datasource.DataSource {
	return &AccelByteSessionTemplateDataSource{}
}

// AccelByteSessionTemplateDataSource defines the data source implementation.
type AccelByteSessionTemplateDataSource struct {
	client *session.ConfigurationTemplateService
}

func (d *AccelByteSessionTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_session_template"
}

func (d *AccelByteSessionTemplateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AccelByteSessionTemplate data source",

		Attributes: map[string]schema.Attribute{

			// Populated by user

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the session template",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of session template",
				Required:            true,
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Session template identifier",
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

			// ServerType = NONE is implied when none of the other server types are specified in the configuration

			// Peer-to-Peer server
			"p2p_server": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{},
				Optional:   true,
				Computed:   true,
			},

			// AMS server
			"ams_server": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
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
					},
					"extend_app": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
					},
				},
				Optional: true,
				Computed: true,
			},

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

func (d *AccelByteSessionTemplateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AccelByteSessionTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccelByteSessionTemplateModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(computeSessionTemplateId(data.Namespace.ValueString(), data.Name.ValueString()))

	input := configuration_template.AdminGetConfigurationTemplateV1Params{
		Namespace: data.Namespace.ValueString(),
		Name:      data.Name.ValueString(),
	}
	configTemplate, err := d.client.AdminGetConfigurationTemplateV1Short(&input)
	if err != nil {
		resp.Diagnostics.AddError("Error when accessing AccelByte API", fmt.Sprintf("Unable to read info on AccelByte session template from namespace '%s' name '%s', got error: %s", input.Namespace, input.Name, err))
		return
	}

	diags, err := updateFromApiSessionTemplate(ctx, &data, configTemplate)
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating our internal state from the session template", fmt.Sprintf("Error: %#v", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Read AccelByteSessionTemplateDataSource from AccelByte API", map[string]interface{}{
		"namespace": data.Namespace,
		"name":      data.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeSessionTemplateId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
