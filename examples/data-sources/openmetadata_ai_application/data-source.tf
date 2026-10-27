data "openmetadata_ai_application" "support_copilot" {
  fully_qualified_name = "support-copilot"
}

output "support_copilot_id" {
  value = data.openmetadata_ai_application.support_copilot.id
}
