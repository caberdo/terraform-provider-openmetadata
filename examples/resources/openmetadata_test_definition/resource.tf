resource "openmetadata_test_definition" "table_row_count" {
  name                   = "customTableRowCount"
  display_name           = "Custom table row count"
  description            = "Fails when the table has no rows."
  entity_type            = "TABLE"
  test_platforms         = ["Other"]
  data_quality_dimension = "Completeness"
  supported_services     = ["Mysql", "Snowflake"]
  sql_expression         = "SELECT COUNT(*) FROM {{ table }}"
  source                 = "user"
}
