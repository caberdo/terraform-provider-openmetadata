data "openmetadata_data_product" "customer_insights" {
  fully_qualified_name = "CustomerInsights"
}

output "customer_insights_domains" {
  value = data.openmetadata_data_product.customer_insights.domains
}
