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
