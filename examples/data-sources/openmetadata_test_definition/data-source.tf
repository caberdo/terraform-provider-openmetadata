data "openmetadata_test_definition" "table_row_count" {
  fully_qualified_name = "customTableRowCount"
}

output "table_row_count_test_definition_id" {
  value = data.openmetadata_test_definition.table_row_count.id
}
