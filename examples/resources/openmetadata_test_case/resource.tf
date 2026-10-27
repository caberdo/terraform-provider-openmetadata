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
      description      = "Primary key."
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

# tableRowCountToBeBetween is a built-in definition, so no test_definition
# resource is needed. Use an openmetadata_test_definition resource to register
# a custom SQL test instead.
resource "openmetadata_test_case" "users_row_count" {
  name            = "users_row_count"
  entity_link     = "<#E::table::${openmetadata_table.users.fully_qualified_name}>"
  test_definition = "tableRowCountToBeBetween"
  display_name    = "Users row count"
  description     = "Fails when the users table holds fewer than 100 or more than 1000000 rows."

  parameter_values = [
    {
      name  = "min"
      value = "100"
    },
    {
      name  = "max"
      value = "1000000"
    },
  ]

  compute_passed_failed_row_count = true
}
