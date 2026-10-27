# Import by UUID.
terraform import openmetadata_test_definition.table_row_count 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, which for a test definition is its name.
terraform import openmetadata_test_definition.table_row_count customTableRowCount
