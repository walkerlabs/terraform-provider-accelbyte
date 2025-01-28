resource "accelbyte_session_template" "test" {
  namespace = "providertest"
  name      = "test"

  // "General" screen - Main configuration
  min_players = 2
  max_players = 12

  // "General" screen - Main configuration
  joinability = "OPEN"

  // "General" screen - Connection and Joinability
  invite_timeout               = 60
  inactive_timeout             = 300
  leader_election_grace_period = 240

  # You can specify one of either: p2p_server, ams_server, or custom_server

  #   p2p_server = {}

  ams_server = {
    requested_regions = [
      "eu-central-1",
      "us-west-2",
    ],
    preferred_claim_keys = [
      "default-match-server"
    ],
  }

  # custom_server = {
  # #   extend_app = "testapp"
  #   custom_url = "https://example.com"
  # }

  // "Additional" screen settings
  auto_join_session           = false
  chat_room                   = false
  secret_validation           = false
  generate_code               = true
  immutable_session_storage   = false
  manual_set_ready_for_ds     = false
  tied_teams_session_lifetime = false
  auto_leave_session          = false

  // "Custom Attributes" screen
  custom_attributes = jsonencode({})
}
