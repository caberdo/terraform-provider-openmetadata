resource "openmetadata_event_subscription" "finance_updates" {
  name         = "finance-updates"
  display_name = "Finance updates"
  description  = "Notifies the finance team when owned assets change."
  alert_type   = "Notification"
  enabled      = true
  resources    = ["table"]

  destinations = [
    {
      category = "Owners"
      type     = "ActivityFeed"
    },
    {
      category = "Teams"
      type     = "Slack"
      config = jsonencode({
        endpoint = "https://example.com/hooks/slack"
      })
    },
  ]
}
