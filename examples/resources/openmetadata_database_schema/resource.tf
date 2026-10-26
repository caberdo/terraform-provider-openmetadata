resource "openmetadata_database_service" "local_postgres" {
  name         = "local_postgres"
  service_type = "Postgres"
  description  = "Postgres instance used by the analytics team."
}

resource "openmetadata_database" "sales" {
  name    = "sales"
  service = openmetadata_database_service.local_postgres.name
}

resource "openmetadata_database_schema" "public" {
  name         = "public"
  database     = openmetadata_database.sales.fully_qualified_name
  display_name = "Public"
  description  = "Default schema for the sales database."
}
