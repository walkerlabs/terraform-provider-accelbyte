// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/match2-sdk/pkg/match2clientmodels"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/errors"
)

// AccelByteMatchRuleSetModel is shared between AccelByteMatchRuleSetDataSource and AccelByteMatchRuleSetResource
type AccelByteMatchRuleSetModel struct {
	// Populated by user
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`

	// Computed during Read() operation
	Id types.String `tfsdk:"id"`

	// Must be set by user during resource creation

	EnableCustomMatchFunction types.Bool `tfsdk:"enable_custom_match_function"`

	Configuration types.String `tfsdk:"configuration"`
}

// Used by Create, Read and Update operations on Match Rulesets
// This copies data from the AccelByte API `matchRuleSet` to the TF state `data`
func updateFromApiMatchRuleSet(ctx context.Context, data *AccelByteMatchRuleSetModel, matchRuleSet *match2clientmodels.APIRuleSetPayload) (diag.Diagnostics, error) {

	var diags diag.Diagnostics = nil

	data.EnableCustomMatchFunction = types.BoolValue(*matchRuleSet.EnableCustomMatchFunction)

	// "Custom Attributes" screen
	configurationJson, err := json.Marshal(matchRuleSet.Data)
	if err != nil {
		return diags, errors.Wrap(err, "Unable to convert API's Match RuleSet's data to JSON: "+fmt.Sprintf("%#v", matchRuleSet.Data))
	}

	data.Configuration = types.StringValue(string(configurationJson))
	return diags, nil
}

// Used by Create/Update operations on Match Rulesets
// This reads from the TF state `data` and returns an AccelByte API object
func toApiMatchRuleSet(ctx context.Context, data AccelByteMatchRuleSetModel) (*match2clientmodels.APIRuleSetPayload, diag.Diagnostics, error) {

	var diags diag.Diagnostics = nil

	var configurationJson interface{}
	err := json.Unmarshal([]byte(data.Configuration.ValueString()), &configurationJson)
	if err != nil {
		return nil, diags, errors.Wrap(err, "Unable to convert Match Ruleset's configuration to JSON: "+fmt.Sprintf("%#v", data.Configuration))
	}

	return &match2clientmodels.APIRuleSetPayload{
		Name: data.Name.ValueStringPointer(),

		EnableCustomMatchFunction: data.EnableCustomMatchFunction.ValueBoolPointer(),
		Data:                      configurationJson,
	}, diags, nil
}
