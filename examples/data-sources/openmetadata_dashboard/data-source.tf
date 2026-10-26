data "openmetadata_dashboard" "revenue" {
  fully_qualified_name = "analytics_dashboards.revenue"
}

output "revenue_dashboard_id" {
  value = data.openmetadata_dashboard.revenue.id
}
