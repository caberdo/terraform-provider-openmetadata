data "openmetadata_database_schema" "public" {
  fully_qualified_name = "local_postgres.sales.public"
}

output "public_schema_id" {
  value = data.openmetadata_database_schema.public.id
}
