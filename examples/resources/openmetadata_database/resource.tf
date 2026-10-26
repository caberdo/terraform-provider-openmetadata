resource "openmetadata_database_service" "warehouse" {
  name         = "warehouse"
  service_type = "Postgres"
}

resource "openmetadata_database" "sales" {
  name         = "sales"
  service      = openmetadata_database_service.warehouse.name
  display_name = "Sales"
  description  = "Sales transactional database."
}
