data "openmetadata_storage_service" "data_lake" {
  fully_qualified_name = "data-lake"
}

output "data_lake_storage_service_id" {
  value = data.openmetadata_storage_service.data_lake.id
}
