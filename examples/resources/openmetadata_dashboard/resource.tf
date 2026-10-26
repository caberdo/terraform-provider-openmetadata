resource "openmetadata_dashboard_service" "analytics" {
  name         = "analytics_dashboards"
  service_type = "CustomDashboard"
  description  = "Dashboard service owned by the analytics team."
}

resource "openmetadata_dashboard" "revenue" {
  name           = "revenue"
  service        = openmetadata_dashboard_service.analytics.fully_qualified_name
  display_name   = "Revenue"
  description    = "Monthly revenue dashboard."
  dashboard_type = "Dashboard"
  source_url     = "https://bi.example.com/dashboards/revenue"
}
