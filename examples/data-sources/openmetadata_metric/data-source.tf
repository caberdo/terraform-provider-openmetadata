data "openmetadata_metric" "weekly_revenue" {
  fully_qualified_name = "WeeklyRevenue"
}

output "weekly_revenue_metric_type" {
  value = data.openmetadata_metric.weekly_revenue.metric_type
}
