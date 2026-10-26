data "openmetadata_chart" "revenue_by_region" {
  fully_qualified_name = "looker.revenue_by_region"
}

output "revenue_by_region_chart_id" {
  value = data.openmetadata_chart.revenue_by_region.id
}
