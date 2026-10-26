resource "openmetadata_database_service" "warehouse" {
  name         = "warehouse"
  service_type = "Postgres"
  display_name = "Warehouse"
  description  = "Database service for the analytics warehouse."
}
