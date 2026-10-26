data "openmetadata_database" "sales" {
  fully_qualified_name = "warehouse.sales"
}

output "sales_database_id" {
  value = data.openmetadata_database.sales.id
}
