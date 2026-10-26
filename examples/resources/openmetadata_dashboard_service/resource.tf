resource "openmetadata_dashboard_service" "custom" {
  name         = "custom_dashboard"
  display_name = "Custom Dashboard"
  description  = "Dashboards built in-house and served by the analytics platform."
  service_type = "CustomDashboard"
}
