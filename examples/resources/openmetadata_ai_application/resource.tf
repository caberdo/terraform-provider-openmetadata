resource "openmetadata_ai_application" "support_copilot" {
  name              = "support-copilot"
  application_type  = "Copilot"
  development_stage = "Production"
  deployment_url    = "https://copilot.example.com"
  display_name      = "Support Copilot"
  description       = "Copilot that drafts answers to customer support tickets."
  documentation     = "https://docs.example.com/support-copilot"
  source_code       = "https://github.com/example/support-copilot"
}
