---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "accelbyte_match_ruleset Data Source - accelbyte"
subcategory: ""
description: |-
  This data source represents a match ruleset https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-rulesets/.
---

# accelbyte_match_ruleset (Data Source)

This data source represents a [match ruleset](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-rulesets/).

## Example Usage

```terraform
data "accelbyte_match_ruleset" "match_ruleset" {
  namespace = "providertest"
  name      = "test"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of match ruleset. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.
- `namespace` (String) Game Namespace which contains the match ruleset. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.

### Read-Only

- `configuration` (String) Matchmaking ruleset configuration in JSON format. See [ruleset docs](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-rulesets/#overview).
- `enable_custom_match_function` (Boolean) If set to `false`, then the `configuration` block's JSON content will be validated to contain only settings relevant to the default AGS Matchmaking logic. If set to `true`, the `configuration` block will only be validated to be valid JSON. Setting to `true` allows you to pass your own settings to your own custom matchmaking functions.
- `id` (String) Match ruleset identifier, on the format `{{namespace}}/{{name}}`.
