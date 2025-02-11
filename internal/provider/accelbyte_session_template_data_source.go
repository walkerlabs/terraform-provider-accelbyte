// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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
		MarkdownDescription: "This data source represents a [session template](https://docs.accelbyte.io/gaming-services/services/play/session/configure-session-templates/).",

		Attributes: map[string]schema.Attribute{

			// Populated by user

			"namespace": schema.StringAttribute{
				MarkdownDescription: "Game Namespace which contains the session template. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of session template. All lowercase characters, max 48 characters in length.",
				Required:            true,
			},

			// Computed during Read() operation

			"id": schema.StringAttribute{
				MarkdownDescription: "Session template identifier, on the format `{{namespace}}/{{name}}`.",
				Computed:            true,
			},

			// Fetched from AccelByte API during Read() opearation

			// "General" screen - Main configuration
			"min_players": schema.Int32Attribute{
				MarkdownDescription: "The minimum number of players required on each team to start the match.",
				Computed:            true,
			},
			"max_players": schema.Int32Attribute{
				MarkdownDescription: "The maximum number of players in a session.",
				Computed:            true,
			},

			// "General" screen - Connection and Joinability
			"joinability": schema.StringAttribute{
				MarkdownDescription: "Control which players can join a session, and through which methods:\n\n" +
					"`OPEN`: Any player can join, either via the session browser or via matchmaking.\n" +
					"`FRIENDS_OF_LEADER`: Only friends of the leader of the session can join, either via the session browser or via matchmaking.\n" +
					"`FRIENDS_OF_FRIENDS`: Only friends of friends of the leader of the session can join, either via the session browser or via matchmaking.\n" +
					"`FRIENDS_OF_MEMBERS`: Friends of any session member can join, either via the session browser or via matchmaking.\n" +
					"`INVITE_ONLY`: Only players who have received an invitation to join the session through either matchmaking, a player in the session requesting to add another player, or a join code that is automatically generated for the session, can join.\n" +
					"`CLOSED`: Players cannot initiate joining a session. They can still be purposefully added as part of matchmaking, or by the game client that requested the creation of the session.",
				Computed: true,
			},

			// "General" screen - Main configuration
			"max_active_sessions": schema.Int32Attribute{
				MarkdownDescription: "The maximum number of sessions of this session template type that a single player can be part of at the same time. -1 = unlimited, >1 = limit value in effect.",
				Computed:            true,
			},
			"custom_session_function": schema.SingleNestedAttribute{
				MarkdownDescription: "Customization points for the session manager. See [docs](https://docs.accelbyte.io/gaming-services/services/extend/override/getting-started-with-session-manager-customization/).",
				Attributes: map[string]schema.Attribute{
					"on_session_created": schema.BoolAttribute{
						MarkdownDescription: "If set to true, the `OnSessionCreated` callback will be invoked when the game session is created.",
						Computed:            true,
					},
					"on_session_updated": schema.BoolAttribute{
						MarkdownDescription: "If set to true, the `OnSessionUpdated` callback will be invoked whenever there are any modification/updates made to the game session.",
						Computed:            true,
					},
					"on_session_deleted": schema.BoolAttribute{
						MarkdownDescription: "If set to true, the `OnSessionDeleted` callback will be invoked when the game session is marked as deleted.",
						Computed:            true,
					},
					"on_party_created": schema.BoolAttribute{
						MarkdownDescription: "If set to true, the `OnPartyCreated` callback will be invoked when the party session is created.",
						Computed:            true,
					},
					"on_party_updated": schema.BoolAttribute{
						MarkdownDescription: "If set to true, the `OnPartyUpdated` callback will be invoked whenever there are any modification/updates made to the party session.",
						Computed:            true,
					},
					"on_party_deleted": schema.BoolAttribute{
						MarkdownDescription: "If set to true, the `OnPartyDeleted` callback will be invoked when the party session is marked as deleted.",
						Computed:            true,
					},
					"custom_url": schema.StringAttribute{
						MarkdownDescription: "Custom URL to a HTTP server. This HTTP server will be called for the events you have enabled. Cannot be used in conjunction with `extend_app`.",
						Computed:            true,
					},
					"extend_app": schema.StringAttribute{
						MarkdownDescription: "Name of an Extend Override app. This app will be called for the events you have enabled. Cannot be used in conjunction with `custom_url`.",
						Computed:            true,
					},
				},
				Optional: true,
				Computed: true,
			},

			// "General" screen - Connection and Joinability
			"invite_timeout": schema.Int32Attribute{
				MarkdownDescription: "Timeout value when invited players will be removed unless they accept or join the invite.",
				Computed:            true,
			},
			"inactive_timeout": schema.Int32Attribute{
				MarkdownDescription: "Timeout value when inactive players will be removed from the session, based on disconnection event from the lobby.",
				Computed:            true,
			},
			"leader_election_grace_period": schema.Int32Attribute{
				MarkdownDescription: "This is only used for party sessions. When the leader disconnects, it will wait until the value is reached, the leader will be changed.",
				Computed:            true,
			},

			// ServerType = NONE is implied when none of the other server types are specified in the configuration

			// Peer-to-Peer server
			"p2p_server": schema.SingleNestedAttribute{
				MarkdownDescription: "Sessions are peer-hosted. Cannot be used in conjunction with `ams_server` or `custom_server`.",
				Attributes:          map[string]schema.Attribute{},
				Optional:            true,
				Computed:            true,
			},

			// AMS server
			"ams_server": schema.SingleNestedAttribute{
				MarkdownDescription: "Sessions are hosted by dedicated servers on AMS. Cannot be used in conjunction with `p2p_server` or `custom_server`.",
				Attributes: map[string]schema.Attribute{
					"requested_regions": schema.ListAttribute{
						MarkdownDescription: "List of regions that will be considered when finding a suitable DS.",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"preferred_claim_keys": schema.ListAttribute{
						MarkdownDescription: "These claim keys have priority over the client version. If a DS can be found using these claim keys, the client version will be ignored.",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"fallback_claim_keys": schema.ListAttribute{
						MarkdownDescription: "The client version has priority over these these claim keys. If a DS cannot be found using either `preferred_claim_keys` or the client version, a third match attempt will be done using these claim keys.",
						ElementType:         types.StringType,
						Computed:            true,
					},
				},
				Optional: true,
				Computed: true,
			},

			// Custom server
			"custom_server": schema.SingleNestedAttribute{
				MarkdownDescription: "Sessions are hosted by [a custom mechanism of your choosing](https://docs.accelbyte.io/gaming-services/services/extend/override/session-dsm-function/). Cannot be used in conjunction with `p2p_server` or `ams_server`.",
				Attributes: map[string]schema.Attribute{
					"custom_url": schema.StringAttribute{
						MarkdownDescription: "Custom URL to a HTTP server. This HTTP server will be called for [creating and terminating game sessions](https://docs.accelbyte.io/gaming-services/services/extend/override/session-dsm-function/#contract-functions). Cannot be used in conjunction with `extend_app`.",
						Optional:            true,
						Computed:            true,
					},
					"extend_app": schema.StringAttribute{
						MarkdownDescription: "Name of an Extend Override app. This app will be called for [creating and terminating game sessions](https://docs.accelbyte.io/gaming-services/services/extend/override/session-dsm-function/#contract-functions). Cannot be used in conjunction with `custom_url`.",
						Optional:            true,
						Computed:            true,
					},
				},
				Optional: true,
				Computed: true,
			},

			// "Additional" screen settings
			"auto_join_session": schema.BoolAttribute{
				MarkdownDescription: "Make the player auto-accept the session invitation and join the session directly.",
				Computed:            true,
			},
			"chat_room": schema.BoolAttribute{
				MarkdownDescription: "Provide a text-based chatroom for users to chat with each other.",
				Computed:            true,
			},
			"secret_validation": schema.BoolAttribute{
				MarkdownDescription: "Generate a secret key for player validation during connection. The secret key will be used to authenticate and validate the player connect and travel to the Dedicated Server (DS).",
				Computed:            true,
			},
			"generate_code": schema.BoolAttribute{
				MarkdownDescription: "Generate a code that can be used as a multiple-use, ticket for joining the session. The code will remain valid for the duration of the session.",
				Computed:            true,
			},
			"immutable_session_storage": schema.BoolAttribute{
				MarkdownDescription: "Prevent players from modifying the session storage.",
				Computed:            true,
			},
			"manual_set_ready_for_ds": schema.BoolAttribute{
				MarkdownDescription: "Require the DS to call a specific endpoint, signaling that the DS will be ready to accept client connections for the game session, before AGS will send players to the DS.",
				Computed:            true,
			},
			"tied_teams_session_lifetime": schema.BoolAttribute{
				MarkdownDescription: "Link the lifetime of any 'partyId' session within the 'teams' attribute to the game session. This exclusively applies when the 'partyId' within 'teams' is the game session.",
				Computed:            true,
			},
			"auto_leave_session": schema.BoolAttribute{
				MarkdownDescription: "Limit the player's active game session to always one. The player will leave the current game session when joining a new one. The Player Active Session Limit will be ignored when this config is enabled.",
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
		notFoundError := &configuration_template.AdminGetConfigurationTemplateV1NotFound{}
		if errors.As(err, &notFoundError) {
			// The data source does not exist in the AccelByte backend
			// This is an actual error; do not update Terraform state, and signal an error to Terraform
			resp.Diagnostics.AddError("Data source not found", fmt.Sprintf("Session template '%s' does not exist in namespace '%s'", input.Name, input.Namespace))
			return
		} else {
			// Failed to retrieve the data source from the AccelByte backend
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

	diags, err := updateFromApiSessionTemplate(ctx, &data, configTemplate)
	resp.Diagnostics.Append(diags...)
	if err != nil {
		resp.Diagnostics.AddError("Error when updating our internal state to match session template", fmt.Sprintf("Error: %#v", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func computeSessionTemplateId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
