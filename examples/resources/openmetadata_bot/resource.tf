// A bot is bound at creation to an existing user flagged as a bot. Under basic
// auth the server keeps the user name equal to the email local part, so both
// derive from the same prefix.
resource "openmetadata_user" "reporting_bot" {
  name             = "reporting-bot"
  email            = "reporting-bot@example.com"
  is_bot           = true
  jwt_token_expiry = "30"
}

resource "openmetadata_bot" "reporting" {
  name         = "reporting"
  bot_user     = openmetadata_user.reporting_bot.name
  display_name = "Reporting bot"
  description  = "Runs the scheduled reporting workflows."
}
