data "openmetadata_dashboard_service" "custom" {
  fully_qualified_name = "custom_dashboard"
}

output "custom_dashboard_service_id" {
  value = data.openmetadata_dashboard_service.custom.id
}
