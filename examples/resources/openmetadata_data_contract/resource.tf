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

resource "openmetadata_table" "users" {
  name            = "users"
  database_schema = openmetadata_database_schema.public.fully_qualified_name
  description     = "One row per registered user."

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
    {
      name             = "email"
      data_type        = "VARCHAR"
      data_length      = 255
      ordinal_position = 2
    },
  ]
}

resource "openmetadata_data_contract" "users" {
  name            = "users-monthly-reporting"
  entity          = openmetadata_table.users.fully_qualified_name
  display_name    = "Users monthly reporting"
  description     = "Contract governing the users table for monthly reporting."
  terms_of_use    = "Only consume this table for internal reporting."
  effective_from  = "2026-01-01T00:00:00Z"
  effective_until = "2026-12-31T23:59:59Z"
}
