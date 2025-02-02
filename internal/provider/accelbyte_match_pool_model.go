// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"reflect"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2clientmodels"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// AccelByteMatchPoolModel is shared between AccelByteMatchPoolDataSource and AccelByteMatchPoolResource
type AccelByteMatchPoolModel struct {
	// Populated by user
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Read() operation
	Id types.String `tfsdk:"id"`

	// Basic information
	RuleSet                 types.String `tfsdk:"rule_set"`
	SessionTemplate         types.String `tfsdk:"session_template"`
	TicketExpirationSeconds types.Int32  `tfsdk:"ticket_expiration_seconds"`

	// Best latency calculation method
	BestLatencyCalculationMethod types.String `tfsdk:"best_latency_calculation_method"`

	// Backfill
	AutoAcceptBackfillProposal        types.Bool  `tfsdk:"auto_accept_backfill_proposal"`
	BackfillProposalExpirationSeconds types.Int32 `tfsdk:"backfill_proposal_expiration_seconds"`
	BackfillTicketExpirationSeconds   types.Int32 `tfsdk:"backfill_ticket_expiration_seconds"`

	// Customization
	MatchFunction         types.String `tfsdk:"match_function"`
	MatchFunctionOverride types.Object `tfsdk:"match_function_override"` // AccelByteMatchPoolMatchFunctionOverrideModel

	// Matchmaking Preferences
	CrossplayEnabled     types.Bool `tfsdk:"crossplay_enabled"`
	PlatformGroupEnabled types.Bool `tfsdk:"platform_group_enabled"`
}

var AccelByteMatchPoolMatchFunctionOverrideModelAttributeTypes = map[string]attr.Type{
	"backfill_matches": types.StringType,
	"enrichment":       types.ListType{}.WithElementType(types.StringType),
	"make_matches":     types.StringType,
	"stat_codes":       types.ListType{}.WithElementType(types.StringType),
	"validation":       types.ListType{}.WithElementType(types.StringType),
}

type AccelByteMatchPoolMatchFunctionOverrideModel struct {
	BackfillMatches types.String `tfsdk:"backfill_matches"`
	Enrichment      types.List   `tfsdk:"enrichment"`
	MakeMatches     types.String `tfsdk:"make_matches"`
	StatCodes       types.List   `tfsdk:"stat_codes"`
	Validation      types.List   `tfsdk:"validation"`
}

// Replacement for types.ListValueFrom(), which guarantees to return a list object even
//   if the input elements is an empty slice or nil
func listValueFromEvenIfNil(ctx context.Context, elementType attr.Type, elements any) (basetypes.ListValue, diag.Diagnostics) {
	// The input elements are potentially an empty slice (which compares as nil) or nil
	// These are wrapped in an interface, so to detect an empty slice / nil we need to inspect the inner type of the interface
	if !reflect.ValueOf(elements).IsNil() {
		// We know it's not an empty slice / nil; construct list from actual elements
		return types.ListValueFrom(ctx, elementType, elements)
	} else {
		// We know it's an empty slice / nil; construct list and pass in an explicit empty slice of an arbitrary type
		// This will create an empty list of the appropriate type
		return types.ListValueFrom(ctx, elementType, []int{})
	}
}

// Used by Create, Read and Update operations on Match Pools
// This copies data from the AccelByte API `pool` to the TF state `data`
func updateFromApiMatchPool(ctx context.Context, data *AccelByteMatchPoolModel, pool *match2clientmodels.APIMatchPool) diag.Diagnostics {

	var diags diag.Diagnostics = nil

	// Basic information
	data.RuleSet = types.StringValue(*pool.RuleSet)
	data.SessionTemplate = types.StringValue(*pool.SessionTemplate)
	data.TicketExpirationSeconds = types.Int32Value(*pool.TicketExpirationSeconds)

	// Best latency calculation method
	data.BestLatencyCalculationMethod = types.StringValue(pool.BestLatencyCalculationMethod)

	// Backfill
	data.AutoAcceptBackfillProposal = types.BoolValue(*pool.AutoAcceptBackfillProposal)
	data.BackfillProposalExpirationSeconds = types.Int32Value(*pool.BackfillProposalExpirationSeconds)
	data.BackfillTicketExpirationSeconds = types.Int32Value(*pool.BackfillTicketExpirationSeconds)

	// Customization
	data.MatchFunction = types.StringValue(*pool.MatchFunction)

	// Special handling of MatchFunctionOverride:
	// This nested attribute can exist (but with all empty sub-attributes) in the TF plan & state, but the API will prefer to return null instead of an empty object
	// This logic is written so that if the user has requested an empty nesteed attribute in the plan, and the API returns null, then this will be translated into
	//   an empty nested attribute for the state
	// Without this translation, we would need to disallow TF plans with empty MatchFunctionOverride nested attributes (since the API cannot represent & persist
	//   the case of MatchFunctionOverride existing but not being configured). That would be less user friendly than this dance of pretending that
	//   empty MatchFunctionOverride objects can exist in the API.
	matchFunctionOverrideExistsInPlan := (!data.MatchFunctionOverride.IsNull() && !data.MatchFunctionOverride.IsUnknown())
	matchFunctionOverrideExistsInApi := pool.MatchFunctionOverride != nil

	if !matchFunctionOverrideExistsInPlan && !matchFunctionOverrideExistsInApi {
		// The previous plan did not include any MatchFunctionOverride, and the API does not have any MatchFunctionOverride; the state should have a null nested attribute
		data.MatchFunctionOverride = basetypes.NewObjectNull(AccelByteMatchPoolMatchFunctionOverrideModelAttributeTypes)
	} else {
		// The previous contained a MatchFunctionOverride, and/or the API returned a MatchFunctionOverride; the state should contain a non-null nested attribute

		var matchFunctionOverrideModel *AccelByteMatchPoolMatchFunctionOverrideModel

		if matchFunctionOverrideExistsInApi {
			enrichment, enrichmentDiags := listValueFromEvenIfNil(ctx, types.StringType, pool.MatchFunctionOverride.Enrichment)
			diags.Append(enrichmentDiags...)
			statCodes, statCodesDiags := listValueFromEvenIfNil(ctx, types.StringType, pool.MatchFunctionOverride.StatCodes)
			diags.Append(statCodesDiags...)
			validation, validationDiags := listValueFromEvenIfNil(ctx, types.StringType, pool.MatchFunctionOverride.Validation)
			diags.Append(validationDiags...)

			matchFunctionOverrideModel = &AccelByteMatchPoolMatchFunctionOverrideModel{
				BackfillMatches: types.StringValue(pool.MatchFunctionOverride.BackfillMatches),
				Enrichment:      enrichment,
				MakeMatches:     types.StringValue(pool.MatchFunctionOverride.MakeMatches),
				StatCodes:       statCodes,
				Validation:      validation,
			}
		} else {
			// There is no source data in the API; create nested attribute's sub-attributes will be empty
			matchFunctionOverrideModel = &AccelByteMatchPoolMatchFunctionOverrideModel{
				BackfillMatches: types.StringValue(""),
				Enrichment:      types.ListNull(types.StringType),
				MakeMatches:     types.StringValue(""),
				StatCodes:       types.ListNull(types.StringType),
				Validation:      types.ListNull(types.StringType),
			}
		}

		matchFunctionOverride, matchFunctionOverrideDiags := basetypes.NewObjectValueFrom(ctx, AccelByteMatchPoolMatchFunctionOverrideModelAttributeTypes, matchFunctionOverrideModel)
		data.MatchFunctionOverride = matchFunctionOverride
		diags.Append(matchFunctionOverrideDiags...)
	}

	// Matchmaking Preferences
	data.CrossplayEnabled = types.BoolValue(!pool.CrossplayDisabled)
	data.PlatformGroupEnabled = types.BoolValue(pool.PlatformGroupEnabled)

	return diags
}

// Used by Create/Update operations on Match Pools
// This reads from the TF state `matchFunctionOverride` and returns an AccelByte API sub-object
func toApiMatchFunctionOverride(ctx context.Context, matchFunctionOverride types.Object) (*match2clientmodels.APIMatchFunctionOverride, diag.Diagnostics) {

	var matchFunctionOverrideModel AccelByteMatchPoolMatchFunctionOverrideModel
	diags := matchFunctionOverride.As(ctx, &matchFunctionOverrideModel, basetypes.ObjectAsOptions{})

	enrichment := make([]string, len(matchFunctionOverrideModel.Enrichment.Elements()))
	statCodes := make([]string, len(matchFunctionOverrideModel.StatCodes.Elements()))
	validation := make([]string, len(matchFunctionOverrideModel.Validation.Elements()))
	diags.Append(matchFunctionOverrideModel.Enrichment.ElementsAs(ctx, &enrichment, false)...)
	diags.Append(matchFunctionOverrideModel.StatCodes.ElementsAs(ctx, &statCodes, false)...)
	diags.Append(matchFunctionOverrideModel.Validation.ElementsAs(ctx, &validation, false)...)

	apiMatchFunctionOverride := &match2clientmodels.APIMatchFunctionOverride{
		BackfillMatches: matchFunctionOverrideModel.BackfillMatches.ValueString(),
		Enrichment:      enrichment,
		MakeMatches:     matchFunctionOverrideModel.MakeMatches.ValueString(),
		StatCodes:       statCodes,
		Validation:      validation,
	}

	return apiMatchFunctionOverride, diags
}

// Used by Create operations on Match Pools
// This reads from the TF state `data` and returns an AccelByte API object
func toApiMatchPool(ctx context.Context, data AccelByteMatchPoolModel) (*match2clientmodels.APIMatchPool, diag.Diagnostics) {

	var diags diag.Diagnostics = nil

	// Handle match function override

	var matchFunctionOverride *match2clientmodels.APIMatchFunctionOverride = nil

	if !data.MatchFunctionOverride.IsNull() && !data.MatchFunctionOverride.IsUnknown() {

		matchFunctionOverride0, matchFunctionOverrideDiags := toApiMatchFunctionOverride(ctx, data.MatchFunctionOverride)
		matchFunctionOverride = matchFunctionOverride0
		diags.Append(matchFunctionOverrideDiags...)
	}

	return &match2clientmodels.APIMatchPool{
		Name: data.Name.ValueStringPointer(),

		// Basic information
		RuleSet:                 data.RuleSet.ValueStringPointer(),
		SessionTemplate:         data.SessionTemplate.ValueStringPointer(),
		TicketExpirationSeconds: data.TicketExpirationSeconds.ValueInt32Pointer(),

		// Best latency calculation method
		BestLatencyCalculationMethod: data.BestLatencyCalculationMethod.ValueString(),

		// Backfill
		AutoAcceptBackfillProposal:        data.AutoAcceptBackfillProposal.ValueBoolPointer(),
		BackfillProposalExpirationSeconds: data.BackfillProposalExpirationSeconds.ValueInt32Pointer(),
		BackfillTicketExpirationSeconds:   data.BackfillTicketExpirationSeconds.ValueInt32Pointer(),

		// Customization
		MatchFunction:         data.MatchFunction.ValueStringPointer(),
		MatchFunctionOverride: matchFunctionOverride,

		// Matchmaking Preferences
		CrossplayDisabled:    !data.CrossplayEnabled.ValueBool(),
		PlatformGroupEnabled: data.PlatformGroupEnabled.ValueBool(),
	}, diags
}

// Used by Update operations on Match Pools
// This reads from the TF state `data` and returns an AccelByte API object
func toApiMatchPoolConfig(ctx context.Context, data AccelByteMatchPoolModel) (*match2clientmodels.APIMatchPoolConfig, diag.Diagnostics) {

	var diags diag.Diagnostics = nil

	// Handle match function override

	var matchFunctionOverride *match2clientmodels.APIMatchFunctionOverride = nil

	if !data.MatchFunctionOverride.IsNull() && !data.MatchFunctionOverride.IsUnknown() {

		matchFunctionOverride0, matchFunctionOverrideDiags := toApiMatchFunctionOverride(ctx, data.MatchFunctionOverride)
		matchFunctionOverride = matchFunctionOverride0
		diags.Append(matchFunctionOverrideDiags...)
	}

	return &match2clientmodels.APIMatchPoolConfig{

		// Basic information
		RuleSet:                 data.RuleSet.ValueStringPointer(),
		SessionTemplate:         data.SessionTemplate.ValueStringPointer(),
		TicketExpirationSeconds: data.TicketExpirationSeconds.ValueInt32Pointer(),

		// Best latency calculation method
		BestLatencyCalculationMethod: data.BestLatencyCalculationMethod.ValueString(),

		// Backfill
		AutoAcceptBackfillProposal:        data.AutoAcceptBackfillProposal.ValueBoolPointer(),
		BackfillProposalExpirationSeconds: data.BackfillProposalExpirationSeconds.ValueInt32Pointer(),
		BackfillTicketExpirationSeconds:   data.BackfillTicketExpirationSeconds.ValueInt32Pointer(),

		// Customization
		MatchFunction:         data.MatchFunction.ValueStringPointer(),
		MatchFunctionOverride: matchFunctionOverride,

		// Matchmaking Preferences
		CrossplayDisabled:    !data.CrossplayEnabled.ValueBool(),
		PlatformGroupEnabled: data.PlatformGroupEnabled.ValueBool(),
	}, diags
}
