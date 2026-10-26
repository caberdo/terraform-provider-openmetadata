resource "openmetadata_domain" "marketing" {
  name        = "Marketing"
  domain_type = "Consumer-aligned"
}

resource "openmetadata_data_product" "customer_insights" {
  name         = "CustomerInsights"
  display_name = "Customer Insights"
  description  = "Curated customer analytics data product."
  domains      = [openmetadata_domain.marketing.fully_qualified_name]
}
