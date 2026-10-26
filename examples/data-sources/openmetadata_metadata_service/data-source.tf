data "openmetadata_metadata_service" "elastic" {
  fully_qualified_name = "elastic-metadata"
}

output "elastic_metadata_service_id" {
  value = data.openmetadata_metadata_service.elastic.id
}
