resource "openmetadata_database_service" "warehouse" {
  name         = "warehouse"
  service_type = "Postgres"
  display_name = "Warehouse"
  description  = "Database service for the analytics warehouse."
}

resource "openmetadata_ingestion_pipeline" "metadata" {
  name           = "warehouse_metadata"
  display_name   = "Warehouse metadata ingestion"
  description    = "Ingests databases, schemas and tables from the warehouse service."
  service        = openmetadata_database_service.warehouse.fully_qualified_name
  service_type   = "databaseService"
  pipeline_type  = "metadata"
  source_config  = jsonencode({ type = "DatabaseMetadata" })
  airflow_config = jsonencode({ scheduleInterval = "0 * * * *" })
  logger_level   = "INFO"
}
