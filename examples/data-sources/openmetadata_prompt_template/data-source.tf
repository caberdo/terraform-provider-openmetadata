data "openmetadata_prompt_template" "ticket_classifier" {
  fully_qualified_name = "ticket-classifier"
}

output "ticket_classifier_template_content" {
  value = data.openmetadata_prompt_template.ticket_classifier.template_content
}
