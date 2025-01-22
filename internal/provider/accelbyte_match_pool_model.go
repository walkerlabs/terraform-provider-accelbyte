// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2clientmodels"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccelByteMatchPoolModel is shared between AccelByteMatchPoolDataSource and AccelByteMatchPoolResource
type AccelByteMatchPoolModel struct {
	// Populated by user
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Read() operation
	Id types.String `tfsdk:"id"`

	// Must be set by user during resource creation

	RuleSet         types.String `tfsdk:"rule_set"`
	SessionTemplate types.String `tfsdk:"session_template"`

	// Can be set by user during resource creation; will otherwise get defaults from API

	AutoAcceptBackfillProposal        types.Bool   `tfsdk:"auto_accept_backfill_proposal"`
	BackfillProposalExpirationSeconds types.Int32  `tfsdk:"backfill_proposal_expiration_seconds"`
	BackfillTicketExpirationSeconds   types.Int32  `tfsdk:"backfill_ticket_expiration_seconds"`
	BestLatencyCalculationMethod      types.String `tfsdk:"best_latency_calculation_method"` // optional
	CrossplayDisabled                 types.Bool   `tfsdk:"crossplay_disabled"`              // optional
	MatchFunction                     types.String `tfsdk:"match_function"`
	// MatchFunctionOverride             types.Object `tfsdk:"match_function_override"` // This is a AccelByteMatchPoolMatchFunctionOverrideDataSourceModel
	PlatformGroupEnabled    types.Bool  `tfsdk:"platform_group_enabled"` // optional
	TicketExpirationSeconds types.Int32 `tfsdk:"ticket_expiration_seconds"`
}

// type AccelByteMatchPoolMatchFunctionOverrideDataSourceModel struct {
// 	BackfillMatches types.String   `tfsdk:"backfill_matches"` // optional
// 	Enrichment      []types.String `tfsdk:"enrichment"`       // optional
// 	MakeMatches     types.String   `tfsdk:"make_matches"`     // optional
// 	StatCodes       []types.String `tfsdk:"stat_codes"`       // optional
// 	Validation      []types.String `tfsdk:"validation"`       // optional
// }

func updateFromApiMatchPool(data *AccelByteMatchPoolModel, pool *match2clientmodels.APIMatchPool) {
	data.AutoAcceptBackfillProposal = types.BoolValue(*pool.AutoAcceptBackfillProposal)
	data.BackfillProposalExpirationSeconds = types.Int32Value(*pool.BackfillProposalExpirationSeconds)
	data.BackfillTicketExpirationSeconds = types.Int32Value(*pool.BackfillTicketExpirationSeconds)
	data.BestLatencyCalculationMethod = types.StringValue(pool.BestLatencyCalculationMethod)
	data.CrossplayDisabled = types.BoolValue(pool.CrossplayDisabled)
	data.MatchFunction = types.StringValue(*pool.MatchFunction)
	// data.MatchFunctionOverride = types.ObjectValue(AccelByteMatchPoolMatchFunctionOverrideDataSourceModel{
	// 	BackfillMatches: pool.MatchFunctionOverride.BackfillMatches,
	// 	Enrichment:      []types.String{pool.MatchFunctionOverride.Enrichment},
	// 	MakeMatches:     pool.MatchFunctionOverride.MakeMatches,
	// 	StatCodes:       []types.String{pool.MatchFunctionOverride.StatCodes},
	// 	Validation:      []types.String{pool.MatchFunctionOverride.Validation},
	// })
	data.PlatformGroupEnabled = types.BoolValue(pool.PlatformGroupEnabled)
	data.RuleSet = types.StringValue(*pool.RuleSet)
	data.SessionTemplate = types.StringValue(*pool.SessionTemplate)
	data.TicketExpirationSeconds = types.Int32Value(*pool.TicketExpirationSeconds)
}

func toApiMatchPool(data AccelByteMatchPoolModel) *match2clientmodels.APIMatchPool {
	return &match2clientmodels.APIMatchPool{
		AutoAcceptBackfillProposal:        data.AutoAcceptBackfillProposal.ValueBoolPointer(),
		BackfillProposalExpirationSeconds: data.BackfillProposalExpirationSeconds.ValueInt32Pointer(),
		BackfillTicketExpirationSeconds:   data.BackfillTicketExpirationSeconds.ValueInt32Pointer(),
		BestLatencyCalculationMethod:      data.BestLatencyCalculationMethod.ValueString(),
		CrossplayDisabled:                 data.CrossplayDisabled.ValueBool(),
		MatchFunction:                     data.MatchFunction.ValueStringPointer(),
		//MatchFunctionOverride: data.MatchFunctionOverride.ValueInt32Pointer(),
		Name:                    data.Name.ValueStringPointer(),
		PlatformGroupEnabled:    data.PlatformGroupEnabled.ValueBool(),
		RuleSet:                 data.RuleSet.ValueStringPointer(),
		SessionTemplate:         data.SessionTemplate.ValueStringPointer(),
		TicketExpirationSeconds: data.TicketExpirationSeconds.ValueInt32Pointer(),
	}
}

func toApiMatchPoolConfig(data AccelByteMatchPoolModel) *match2clientmodels.APIMatchPoolConfig {
	return &match2clientmodels.APIMatchPoolConfig{
		AutoAcceptBackfillProposal:        data.AutoAcceptBackfillProposal.ValueBoolPointer(),
		BackfillProposalExpirationSeconds: data.BackfillProposalExpirationSeconds.ValueInt32Pointer(),
		BackfillTicketExpirationSeconds:   data.BackfillTicketExpirationSeconds.ValueInt32Pointer(),
		BestLatencyCalculationMethod:      data.BestLatencyCalculationMethod.ValueString(),
		CrossplayDisabled:                 data.CrossplayDisabled.ValueBool(),
		MatchFunction:                     data.MatchFunction.ValueStringPointer(),
		//MatchFunctionOverride: data.MatchFunctionOverride.ValueInt32Pointer(),
		PlatformGroupEnabled:    data.PlatformGroupEnabled.ValueBool(),
		RuleSet:                 data.RuleSet.ValueStringPointer(),
		SessionTemplate:         data.SessionTemplate.ValueStringPointer(),
		TicketExpirationSeconds: data.TicketExpirationSeconds.ValueInt32Pointer(),
	}
}
