resource "accelbyte_match_pool" "pool" {
  namespace = "providertest"
  name      = "test"

  match_function   = "default"
  rule_set         = "transmission_dev"
  session_template = "dev01"

  auto_accept_backfill_proposal        = false
  backfill_proposal_expiration_seconds = 60
  backfill_ticket_expiration_seconds   = 60
  best_latency_calculation_method      = ""
  crossplay_disabled                   = true
  ticket_expiration_seconds            = 999999
  platform_group_enabled               = false
}
