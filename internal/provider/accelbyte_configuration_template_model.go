// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/AccelByte/accelbyte-go-sdk/session-sdk/pkg/sessionclientmodels"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccelByteConfigurationTemplateModel is shared between AccelByteConfigurationTemplateDataSource and AccelByteConfigurationTemplateResource
type AccelByteConfigurationTemplateModel struct {
	// Populated by user
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Read() operation
	Id types.String `tfsdk:"id"`

	// Must be set by user during resource creation

	MinPlayers  types.Int32  `tfsdk:"min_players"`
	MaxPlayers  types.Int32  `tfsdk:"max_players"`
	Joinability types.String `tfsdk:"joinability"`

	// Can be set by user during resource creation; will otherwise get defaults from API

}

func updateFromApiConfigurationTemplate(data *AccelByteConfigurationTemplateModel, configurationTemplate *sessionclientmodels.ApimodelsConfigurationTemplateResponse) {
	data.MinPlayers = types.Int32Value(*configurationTemplate.MinPlayers)
	data.MaxPlayers = types.Int32Value(*configurationTemplate.MaxPlayers)
	data.Joinability = types.StringValue(*configurationTemplate.Joinability)
}

func toApiConfigurationTemplate(data AccelByteConfigurationTemplateModel) *sessionclientmodels.ApimodelsCreateConfigurationTemplateRequest {
	return &sessionclientmodels.ApimodelsCreateConfigurationTemplateRequest{
		Name:        data.Name.ValueStringPointer(),
		MinPlayers:  data.MinPlayers.ValueInt32Pointer(),
		MaxPlayers:  data.MaxPlayers.ValueInt32Pointer(),
		Joinability: data.Joinability.ValueStringPointer(),
	}
}

func toApiConfigurationTemplateConfig(data AccelByteConfigurationTemplateModel) *sessionclientmodels.ApimodelsUpdateConfigurationTemplateRequest {
	return &sessionclientmodels.ApimodelsUpdateConfigurationTemplateRequest{
		MinPlayers:  data.MinPlayers.ValueInt32Pointer(),
		MaxPlayers:  data.MaxPlayers.ValueInt32Pointer(),
		Joinability: data.Joinability.ValueStringPointer(),
	}
}
