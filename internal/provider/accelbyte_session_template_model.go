// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclientmodels"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pkg/errors"
)

// AccelByteSessionTemplateModel is shared between AccelByteSessionTemplateDataSource and AccelByteSessionTemplateResource
type AccelByteSessionTemplateModel struct {
	// Populated by user
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Read() operation
	Id types.String `tfsdk:"id"`

	// Must be set by user during resource creation

	// "General" screen - Main configuration
	MinPlayers types.Int32 `tfsdk:"min_players"`
	MaxPlayers types.Int32 `tfsdk:"max_players"`

	// "General" screen - Connection and Joinability
	Joinability types.String `tfsdk:"joinability"`

	// Can be set by user during resource creation; will otherwise get defaults from the schema

	// "General" screen - Main configuration
	MaxActiveSessions types.Int32 `tfsdk:"max_active_sessions"`
	// TODO: support "use Custom Session Function"

	// "General" screen - Connection and Joinability
	InviteTimeout             types.Int32 `tfsdk:"invite_timeout"`
	InactiveTimeout           types.Int32 `tfsdk:"inactive_timeout"`
	LeaderElectionGracePeriod types.Int32 `tfsdk:"leader_election_grace_period"`

	// Only one of these should exist at a time
	P2pServer    types.Object `tfsdk:"p2p_server"`    // AccelByteSessionTemplateP2pServerModel
	AmsServer    types.Object `tfsdk:"ams_server"`    // AccelByteSessionTemplateAmsServerModel
	CustomServer types.Object `tfsdk:"custom_server"` // AccelByteSessionTemplateCustomServerModel

	// "Additional" screen settings
	AutoJoinSession          types.Bool `tfsdk:"auto_join_session"`
	ChatRoom                 types.Bool `tfsdk:"chat_room"`
	SecretValidation         types.Bool `tfsdk:"secret_validation"`
	GenerateCode             types.Bool `tfsdk:"generate_code"`
	ImmutableSessionStorage  types.Bool `tfsdk:"immutable_session_storage"`
	ManualSetReadyForDS      types.Bool `tfsdk:"manual_set_ready_for_ds"`
	TiedTeamsSessionLifetime types.Bool `tfsdk:"tied_teams_session_lifetime"`
	AutoLeaveSession         types.Bool `tfsdk:"auto_leave_session"`

	// TODO: support "3rd party sync" options

	// "Custom Attributes" screen
	CustomAttributes types.String `tfsdk:"custom_attributes"`
}

type AccelByteSessionTemplateP2pServerModel struct {
}

var AccelByteSessionTemplateP2pServerModelAttributeTypes = map[string]attr.Type{}

type AccelByteSessionTemplateAmsServerModel struct {
	RequestedRegions   types.List `tfsdk:"requested_regions"`
	PreferredClaimKeys types.List `tfsdk:"preferred_claim_keys"`
	FallbackClaimKeys  types.List `tfsdk:"fallback_claim_keys"`
}

var AccelByteSessionTemplateAmsServerModelAttributeTypes = map[string]attr.Type{
	"requested_regions":    types.ListType{}.WithElementType(types.StringType),
	"preferred_claim_keys": types.ListType{}.WithElementType(types.StringType),
	"fallback_claim_keys":  types.ListType{}.WithElementType(types.StringType),
}

type AccelByteSessionTemplateCustomServerModel struct {
	CustomUrl types.String `tfsdk:"custom_url"`
	ExtendApp types.String `tfsdk:"extend_app"`
}

var AccelByteSessionTemplateCustomServerModelAttributeTypes = map[string]attr.Type{
	"custom_url": types.StringType,
	"extend_app": types.StringType,
}

type AccelByteSessionTemplateServerType string

const (
	AccelByteSessionTemplateServerTypeNone AccelByteSessionTemplateServerType = "NONE"
	AccelByteSessionTemplateServerTypeP2P  AccelByteSessionTemplateServerType = "P2P"
	AccelByteSessionTemplateServerTypeDS   AccelByteSessionTemplateServerType = "DS"
)

type AccelByteSessionTemplateDsSourceType string

const (
	AccelByteSessionTemplateDsSourceNone   AccelByteSessionTemplateDsSourceType = ""
	AccelByteSessionTemplateDsSourceAms    AccelByteSessionTemplateDsSourceType = "AMS"
	AccelByteSessionTemplateDsSourceCustom AccelByteSessionTemplateDsSourceType = "custom"
)

func updateFromApiSessionTemplate(ctx context.Context, data *AccelByteSessionTemplateModel, configurationTemplate *sessionclientmodels.ApimodelsConfigurationTemplateResponse) (diag.Diagnostics, error) {

	var diags diag.Diagnostics = nil

	data.MinPlayers = types.Int32Value(*configurationTemplate.MinPlayers)
	data.MaxPlayers = types.Int32Value(*configurationTemplate.MaxPlayers)
	data.Joinability = types.StringValue(*configurationTemplate.Joinability)

	// "General" screen - Main configuration
	data.MaxActiveSessions = types.Int32Value(configurationTemplate.MaxActiveSessions)
	// TODO: support "use Custom Session Function"

	// "General" screen - Connection and Joinability
	data.InviteTimeout = types.Int32Value(*configurationTemplate.InviteTimeout)
	data.InactiveTimeout = types.Int32Value(*configurationTemplate.InactiveTimeout)
	data.LeaderElectionGracePeriod = types.Int32Value(configurationTemplate.LeaderElectionGracePeriod)

	// "General" screen - Server
	serverType := types.StringValue(*configurationTemplate.Type).ValueString()
	dsSource := types.StringValue(configurationTemplate.DsSource).ValueString()
	data.P2pServer = basetypes.NewObjectNull(AccelByteSessionTemplateP2pServerModelAttributeTypes)
	data.AmsServer = basetypes.NewObjectNull(AccelByteSessionTemplateAmsServerModelAttributeTypes)
	data.CustomServer = basetypes.NewObjectNull(AccelByteSessionTemplateCustomServerModelAttributeTypes)

	if serverType == string(AccelByteSessionTemplateServerTypeP2P) {
		p2pServer, p2pServerDiags := basetypes.NewObjectValueFrom(ctx, AccelByteSessionTemplateP2pServerModelAttributeTypes, AccelByteSessionTemplateP2pServerModel{})
		data.P2pServer = p2pServer
		diags.Append(p2pServerDiags...)
	} else if serverType == string(AccelByteSessionTemplateServerTypeDS) && dsSource == string(AccelByteSessionTemplateDsSourceAms) {
		requestedRegions, requestedRegionsDiags := types.ListValueFrom(ctx, types.StringType, configurationTemplate.RequestedRegions)
		diags.Append(requestedRegionsDiags...)
		preferredClaimKeys, preferredClaimKeysDiags := types.ListValueFrom(ctx, types.StringType, configurationTemplate.PreferredClaimKeys)
		diags.Append(preferredClaimKeysDiags...)
		fallbackClaimKeys, fallbackClaimKeysDiags := types.ListValueFrom(ctx, types.StringType, configurationTemplate.FallbackClaimKeys)
		diags.Append(fallbackClaimKeysDiags...)

		amsServerModel := &AccelByteSessionTemplateAmsServerModel{
			RequestedRegions:   requestedRegions,
			PreferredClaimKeys: preferredClaimKeys,
			FallbackClaimKeys:  fallbackClaimKeys,
		}

		amsServer, amsServerDiags := basetypes.NewObjectValueFrom(ctx, AccelByteSessionTemplateAmsServerModelAttributeTypes, amsServerModel)
		data.AmsServer = amsServer
		diags.Append(amsServerDiags...)
	} else if serverType == string(AccelByteSessionTemplateServerTypeDS) && dsSource == string(AccelByteSessionTemplateDsSourceCustom) {

		customServerModel := &AccelByteSessionTemplateCustomServerModel{
			CustomUrl: types.StringValue(configurationTemplate.CustomURLGRPC),
			ExtendApp: types.StringValue(configurationTemplate.AppName),
		}

		customServer, customServerDiags := basetypes.NewObjectValueFrom(ctx, AccelByteSessionTemplateCustomServerModelAttributeTypes, customServerModel)
		data.CustomServer = customServer
		diags.Append(customServerDiags...)
	}

	// "Additional" screen settings
	data.AutoJoinSession = types.BoolValue(configurationTemplate.AutoJoin)
	data.ChatRoom = types.BoolValue(*configurationTemplate.TextChat)
	data.SecretValidation = types.BoolValue(configurationTemplate.EnableSecret)
	data.GenerateCode = types.BoolValue(!configurationTemplate.DisableCodeGeneration)
	data.ImmutableSessionStorage = types.BoolValue(configurationTemplate.ImmutableStorage)
	data.ManualSetReadyForDS = types.BoolValue(configurationTemplate.DsManualSetReady)
	data.TiedTeamsSessionLifetime = types.BoolValue(configurationTemplate.TieTeamsSessionLifetime)
	data.AutoLeaveSession = types.BoolValue(configurationTemplate.AutoLeaveSession)

	// "Custom Attributes" screen
	customAttributesJson, err := json.Marshal(configurationTemplate.Attributes)
	if err != nil {
		return diags, errors.Wrap(err, "Unable to convert API's Session Template's custom attributes to JSON: "+fmt.Sprintf("%#v", configurationTemplate.Attributes))
	}

	data.CustomAttributes = types.StringValue(string(customAttributesJson))
	return diags, nil
}

func toApiSessionTemplate(ctx context.Context, data AccelByteSessionTemplateModel) (*sessionclientmodels.ApimodelsCreateConfigurationTemplateRequest, diag.Diagnostics, error) {

	var diags diag.Diagnostics = nil

	serverType := AccelByteSessionTemplateServerTypeNone
	dsSource := AccelByteSessionTemplateDsSourceNone

	// Handle P2P server

	if !data.P2pServer.IsNull() && !data.P2pServer.IsUnknown() {
		serverType = AccelByteSessionTemplateServerTypeP2P
	}

	// Handle AMS server

	var requestedRegions []string = nil
	var preferredClaimKeys []string = nil
	var fallbackClaimKeys []string = nil

	if !data.AmsServer.IsNull() && !data.AmsServer.IsUnknown() {
		serverType = AccelByteSessionTemplateServerTypeDS
		dsSource = AccelByteSessionTemplateDsSourceAms

		var amsServer AccelByteSessionTemplateAmsServerModel
		diags.Append(data.AmsServer.As(ctx, &amsServer, basetypes.ObjectAsOptions{})...)

		requestedRegions = make([]string, len(amsServer.RequestedRegions.Elements()))
		preferredClaimKeys = make([]string, len(amsServer.PreferredClaimKeys.Elements()))
		fallbackClaimKeys = make([]string, len(amsServer.FallbackClaimKeys.Elements()))
		diags.Append(amsServer.RequestedRegions.ElementsAs(ctx, &requestedRegions, false)...)
		diags.Append(amsServer.PreferredClaimKeys.ElementsAs(ctx, &preferredClaimKeys, false)...)
		diags.Append(amsServer.FallbackClaimKeys.ElementsAs(ctx, &fallbackClaimKeys, false)...)
	}

	// Handle Custom server

	customUrlGrpc := ""
	appName := ""

	if !data.CustomServer.IsNull() && !data.CustomServer.IsUnknown() {
		serverType = AccelByteSessionTemplateServerTypeDS
		dsSource = AccelByteSessionTemplateDsSourceCustom

		var customServer AccelByteSessionTemplateCustomServerModel
		diags.Append(data.CustomServer.As(ctx, &customServer, basetypes.ObjectAsOptions{})...)

		customUrlGrpc = customServer.CustomUrl.ValueString()
		appName = customServer.ExtendApp.ValueString()
	}

	var customAttributesJson interface{}
	err := json.Unmarshal([]byte(data.CustomAttributes.ValueString()), &customAttributesJson)
	if err != nil {
		return nil, diags, errors.Wrap(err, "Unable to convert Session Template's custom attributes to JSON: "+fmt.Sprintf("%#v", data.CustomAttributes))
	}

	serverTypeString := string(serverType)

	return &sessionclientmodels.ApimodelsCreateConfigurationTemplateRequest{
		Name: data.Name.ValueStringPointer(),

		MinPlayers:  data.MinPlayers.ValueInt32Pointer(),
		MaxPlayers:  data.MaxPlayers.ValueInt32Pointer(),
		Joinability: data.Joinability.ValueStringPointer(),

		// "General" screen - Main configuration
		MaxActiveSessions: data.MaxActiveSessions.ValueInt32(),
		// TODO: support "use Custom Session Function"

		// "General" screen - Connection and Joinability
		InviteTimeout:             data.InviteTimeout.ValueInt32Pointer(),
		InactiveTimeout:           data.InactiveTimeout.ValueInt32Pointer(),
		LeaderElectionGracePeriod: data.LeaderElectionGracePeriod.ValueInt32(),

		// "General" screen - Server
		Type:     &serverTypeString,
		DsSource: string(dsSource),
		// Only used when ServerType = DS, DsSource = AMS
		RequestedRegions:   requestedRegions,
		PreferredClaimKeys: preferredClaimKeys,
		FallbackClaimKeys:  fallbackClaimKeys,
		// Only used when ServerType = DS, DsSource = Custom
		CustomURLGRPC: customUrlGrpc,
		AppName:       appName,

		// "Additional" screen settings
		AutoJoin:                data.AutoJoinSession.ValueBool(),
		TextChat:                data.ChatRoom.ValueBoolPointer(),
		EnableSecret:            data.SecretValidation.ValueBool(),
		DisableCodeGeneration:   !data.GenerateCode.ValueBool(),
		ImmutableStorage:        data.ImmutableSessionStorage.ValueBool(),
		DsManualSetReady:        data.ManualSetReadyForDS.ValueBool(),
		TieTeamsSessionLifetime: data.TiedTeamsSessionLifetime.ValueBool(),
		AutoLeaveSession:        data.AutoLeaveSession.ValueBool(),

		// "Custom Attributes" screen
		Attributes: customAttributesJson,
	}, diags, nil
}

func toApiSessionTemplateConfig(ctx context.Context, data AccelByteSessionTemplateModel) (*sessionclientmodels.ApimodelsUpdateConfigurationTemplateRequest, diag.Diagnostics, error) {

	var diags diag.Diagnostics = nil

	serverType := AccelByteSessionTemplateServerTypeNone
	dsSource := AccelByteSessionTemplateDsSourceNone

	// Handle P2P server

	if !data.P2pServer.IsNull() && !data.P2pServer.IsUnknown() {
		serverType = AccelByteSessionTemplateServerTypeP2P
	}

	// Handle AMS server

	var requestedRegions []string = nil
	var preferredClaimKeys []string = nil
	var fallbackClaimKeys []string = nil

	if !data.AmsServer.IsNull() && !data.AmsServer.IsUnknown() {
		serverType = AccelByteSessionTemplateServerTypeDS
		dsSource = AccelByteSessionTemplateDsSourceAms

		var amsServer AccelByteSessionTemplateAmsServerModel
		diags.Append(data.AmsServer.As(ctx, &amsServer, basetypes.ObjectAsOptions{})...)

		requestedRegions = make([]string, len(amsServer.RequestedRegions.Elements()))
		preferredClaimKeys = make([]string, len(amsServer.PreferredClaimKeys.Elements()))
		fallbackClaimKeys = make([]string, len(amsServer.FallbackClaimKeys.Elements()))
		diags.Append(amsServer.RequestedRegions.ElementsAs(ctx, &requestedRegions, false)...)
		diags.Append(amsServer.PreferredClaimKeys.ElementsAs(ctx, &preferredClaimKeys, false)...)
		diags.Append(amsServer.FallbackClaimKeys.ElementsAs(ctx, &fallbackClaimKeys, false)...)
	}

	// Handle Custom server

	customUrlGrpc := ""
	appName := ""

	if !data.CustomServer.IsNull() && !data.CustomServer.IsUnknown() {
		serverType = AccelByteSessionTemplateServerTypeDS
		dsSource = AccelByteSessionTemplateDsSourceCustom

		var customServer AccelByteSessionTemplateCustomServerModel
		diags.Append(data.CustomServer.As(ctx, &customServer, basetypes.ObjectAsOptions{})...)

		customUrlGrpc = customServer.CustomUrl.ValueString()
		appName = customServer.ExtendApp.ValueString()
	}

	var customAttributesJson interface{}
	err := json.Unmarshal([]byte(data.CustomAttributes.ValueString()), &customAttributesJson)
	if err != nil {
		return nil, diags, errors.Wrap(err, "Unable to convert Session Template's custom attributes to JSON: "+fmt.Sprintf("%#v", data.CustomAttributes))
	}

	serverTypeString := string(serverType)

	return &sessionclientmodels.ApimodelsUpdateConfigurationTemplateRequest{
		Name: data.Name.ValueStringPointer(),

		MinPlayers:  data.MinPlayers.ValueInt32Pointer(),
		MaxPlayers:  data.MaxPlayers.ValueInt32Pointer(),
		Joinability: data.Joinability.ValueStringPointer(),

		// "General" screen - Main configuration
		MaxActiveSessions: data.MaxActiveSessions.ValueInt32(),
		// TODO: support "use Custom Session Function"

		// "General" screen - Connection and Joinability
		InviteTimeout:             data.InviteTimeout.ValueInt32Pointer(),
		InactiveTimeout:           data.InactiveTimeout.ValueInt32Pointer(),
		LeaderElectionGracePeriod: data.LeaderElectionGracePeriod.ValueInt32(),

		// "General" screen - Server
		Type:     &serverTypeString,
		DsSource: string(dsSource),
		// Only used when ServerType = DS, DsSource = AMS
		RequestedRegions:   requestedRegions,
		PreferredClaimKeys: preferredClaimKeys,
		FallbackClaimKeys:  fallbackClaimKeys,
		// Only used when ServerType = DS, DsSource = Custom
		CustomURLGRPC: customUrlGrpc,
		AppName:       appName,

		// "Additional" screen settings
		AutoJoin:                data.AutoJoinSession.ValueBool(),
		TextChat:                data.ChatRoom.ValueBoolPointer(),
		EnableSecret:            data.SecretValidation.ValueBool(),
		DisableCodeGeneration:   !data.GenerateCode.ValueBool(),
		ImmutableStorage:        data.ImmutableSessionStorage.ValueBool(),
		DsManualSetReady:        data.ManualSetReadyForDS.ValueBool(),
		TieTeamsSessionLifetime: data.TiedTeamsSessionLifetime.ValueBool(),
		AutoLeaveSession:        data.AutoLeaveSession.ValueBool(),

		// "Custom Attributes" screen
		Attributes: customAttributesJson,
	}, diags, nil
}
