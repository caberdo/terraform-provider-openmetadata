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
  display_name    = "Users"
  description     = "One row per registered user."

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      description      = "Primary key."
      ordinal_position = 1
    },
    {
      name             = "email"
      data_type        = "VARCHAR"
      data_length      = 255
      ordinal_position = 2
    },
    {
      name             = "created_at"
      data_type        = "TIMESTAMP"
      ordinal_position = 3
    },
  ]
}
