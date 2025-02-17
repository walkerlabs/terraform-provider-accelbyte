---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "accelbyte_match_ruleset Resource - accelbyte"
subcategory: ""
description: |-
  This resource represents a match ruleset https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-rulesets/.
---

# accelbyte_match_ruleset (Resource)

This resource represents a [match ruleset](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-rulesets/).

## Example Usage

```terraform
resource "accelbyte_match_ruleset" "match_ruleset" {
  namespace = "providertest"
  name      = "test"

  enable_custom_match_function = false

  configuration = jsonencode({
    "alliance" : {
      "min_number" : 1,
      "max_number" : 3,
      "player_min_number" : 1,
      "player_max_number" : 4
    }
  })
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `configuration` (String) Matchmaking ruleset configuration in JSON format. See [ruleset docs](https://docs.accelbyte.io/gaming-services/services/play/matchmaking/configuring-match-rulesets/#overview).
- `name` (String) Name of match ruleset. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.
- `namespace` (String) Game Namespace which contains the match ruleset. Uppercase characters, lowercase characters, or digits. Max 64 characters in length.

### Optional

- `enable_custom_match_function` (Boolean) If set to `false`, then the `configuration` block's JSON content will be validated to contain only settings relevant to the default AGS Matchmaking logic. If set to `true`, the `configuration` block will only be validated to be valid JSON. Setting to `true` allows you to pass your own settings to your own custom matchmaking functions.

### Read-Only

- `id` (String) Match ruleset identifier, on the format `{{namespace}}/{{name}}`.
