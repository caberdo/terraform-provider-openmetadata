data "openmetadata_notification_template" "asset_created" {
  fully_qualified_name = "asset-created-email"
}

output "asset_created_template_body" {
  value = data.openmetadata_notification_template.asset_created.template_body
}
