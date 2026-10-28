resource "openmetadata_prompt_template" "ticket_classifier" {
  name             = "ticket-classifier"
  display_name     = "Ticket classifier"
  description      = "Classifies incoming support tickets into a category."
  template_content = "Classify the ticket: {{ticket}}"
  template_type    = "Classification"
  template_version = "1.0"
  system_prompt    = "You are a support agent. Reply with a single category."
}
