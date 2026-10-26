resource "openmetadata_database_service" "warehouse" {
  name         = "warehouse"
  service_type = "Postgres"
}

resource "openmetadata_database" "analytics" {
  name    = "analytics"
  service = openmetadata_database_service.warehouse.name
}

resource "openmetadata_database_schema" "public" {
  name     = "public"
  database = openmetadata_database.analytics.fully_qualified_name
}

resource "openmetadata_stored_procedure" "customer_report" {
  name            = "customer_report"
  database_schema = openmetadata_database_schema.public.fully_qualified_name
  display_name    = "Customer report"
  description     = "Aggregates customer activity for the finance dashboard."
  code            = "BEGIN SELECT 1; END"
  language        = "SQL"
}
