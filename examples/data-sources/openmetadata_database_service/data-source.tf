data "openmetadata_database_service" "warehouse" {
  fully_qualified_name = "warehouse"
}

output "warehouse_database_service_id" {
  value = data.openmetadata_database_service.warehouse.id
}
