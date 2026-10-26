resource "openmetadata_dashboard_service" "looker" {
  name         = "looker"
  service_type = "Looker"
}

resource "openmetadata_chart" "revenue_by_region" {
  name         = "revenue_by_region"
  service      = openmetadata_dashboard_service.looker.fully_qualified_name
  display_name = "Revenue by region"
  description  = "Monthly revenue split by region."
  chart_type   = "Bar"
  source_url   = "https://looker.example.com/charts/42"
}
