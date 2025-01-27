// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclientmodels"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
)

// AccelByteConfigurationTemplateModel is shared between AccelByteConfigurationTemplateDataSource and AccelByteConfigurationTemplateResource
type AccelByteConfigurationTemplateModel struct {
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

	// "General" screen - Server
	ServerType types.String `tfsdk:"server_type"`
	// Only used when ServerType = AMS
	RequestedRegions   types.List `tfsdk:"requested_regions"`
	PreferredClaimKeys types.List `tfsdk:"preferred_claim_keys"`
	FallbackClaimKeys  types.List `tfsdk:"fallback_claim_keys"`
	// TODO: support ServerType = CUSTOM

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

func updateFromApiConfigurationTemplate(ctx context.Context, data *AccelByteConfigurationTemplateModel, configurationTemplate *sessionclientmodels.ApimodelsConfigurationTemplateResponse) error {
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
	data.ServerType = types.StringValue(*configurationTemplate.Type)
	// Only used when ServerType = AMS

	requestedRegions, _ := types.ListValueFrom(ctx, types.StringType, configurationTemplate.RequestedRegions)
	data.RequestedRegions = requestedRegions
	preferredClaimKeys, _ := types.ListValueFrom(ctx, types.StringType, configurationTemplate.PreferredClaimKeys)
	data.PreferredClaimKeys = preferredClaimKeys
	fallbackClaimKeys, _ := types.ListValueFrom(ctx, types.StringType, configurationTemplate.FallbackClaimKeys)
	data.FallbackClaimKeys = fallbackClaimKeys

	// TODO: support ServerType = CUSTOM

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
		return errors.Wrap(err, "Unable to convert API's Configuration Template's custom attributes to JSON: "+fmt.Sprintf("%#v", configurationTemplate.Attributes))
	}

	data.CustomAttributes = types.StringValue(string(customAttributesJson))
	return nil
}

func toApiConfigurationTemplate(ctx context.Context, data AccelByteConfigurationTemplateModel) (*sessionclientmodels.ApimodelsCreateConfigurationTemplateRequest, error) {

	requestedRegions := make([]string, len(data.RequestedRegions.Elements()))
	_ = data.RequestedRegions.ElementsAs(ctx, &requestedRegions, false)
	preferredClaimKeys := make([]string, len(data.PreferredClaimKeys.Elements()))
	_ = data.PreferredClaimKeys.ElementsAs(ctx, &preferredClaimKeys, false)
	fallbackClaimKeys := make([]string, len(data.FallbackClaimKeys.Elements()))
	_ = data.FallbackClaimKeys.ElementsAs(ctx, &fallbackClaimKeys, false)

	var customAttributesJson interface{}
	err := json.Unmarshal([]byte(data.CustomAttributes.ValueString()), &customAttributesJson)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to convert Session Template's custom attributes to JSON: "+fmt.Sprintf("%#v", data.CustomAttributes))
	}

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
		Type: data.ServerType.ValueStringPointer(),
		// Only used when ServerType = AMS
		RequestedRegions:   requestedRegions,
		PreferredClaimKeys: preferredClaimKeys,
		FallbackClaimKeys:  fallbackClaimKeys,
		// TODO: support ServerType = CUSTOM

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
	}, nil
}

func toApiConfigurationTemplateConfig(ctx context.Context, data AccelByteConfigurationTemplateModel) (*sessionclientmodels.ApimodelsUpdateConfigurationTemplateRequest, error) {

	requestedRegions := make([]string, len(data.RequestedRegions.Elements()))
	_ = data.RequestedRegions.ElementsAs(ctx, &requestedRegions, false)
	preferredClaimKeys := make([]string, len(data.PreferredClaimKeys.Elements()))
	_ = data.PreferredClaimKeys.ElementsAs(ctx, &preferredClaimKeys, false)
	fallbackClaimKeys := make([]string, len(data.FallbackClaimKeys.Elements()))
	_ = data.FallbackClaimKeys.ElementsAs(ctx, &fallbackClaimKeys, false)

	var customAttributesJson interface{}
	err := json.Unmarshal([]byte(data.CustomAttributes.ValueString()), &customAttributesJson)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to convert Session Template's custom attributes to JSON: "+fmt.Sprintf("%#v", data.CustomAttributes))
	}

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
		Type: data.ServerType.ValueStringPointer(),
		// Only used when ServerType = AMS
		RequestedRegions:   requestedRegions,
		PreferredClaimKeys: preferredClaimKeys,
		FallbackClaimKeys:  fallbackClaimKeys,
		// TODO: support ServerType = CUSTOM

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
	}, nil
}
