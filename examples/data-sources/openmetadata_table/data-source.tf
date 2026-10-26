data "openmetadata_table" "users" {
  fully_qualified_name = "warehouse.analytics.public.users"
}

output "users_table_id" {
  value = data.openmetadata_table.users.id
}

output "users_columns" {
  value = [for column in data.openmetadata_table.users.columns : column.name]
}
