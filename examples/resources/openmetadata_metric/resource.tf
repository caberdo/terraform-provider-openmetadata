resource "openmetadata_metric" "weekly_revenue" {
  name                = "WeeklyRevenue"
  display_name        = "Weekly revenue"
  description         = "Revenue grouped by week."
  metric_type         = "SUM"
  unit_of_measurement = "DOLLARS"
  granularity         = "WEEK"
  expression_code     = "sum(revenue)"
  expression_language = "SQL"
}
