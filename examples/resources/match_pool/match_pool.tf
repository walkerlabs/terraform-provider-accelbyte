resource "accelbyte_match_pool" "pool" {
  namespace = "providertest"
  name      = "test"

  // Basic information
  rule_set                  = "transmission_dev"
  session_template          = "dev01"
  ticket_expiration_seconds = 300

  // Best latency calculation method
  best_latency_calculation_method = ""

  // Backfill
  auto_accept_backfill_proposal        = false
  backfill_proposal_expiration_seconds = 60
  backfill_ticket_expiration_seconds   = 60

  // Customization
  match_function = "default"

  // Matchmaking Preferences
  crossplay_enabled      = false
  platform_group_enabled = false
}
