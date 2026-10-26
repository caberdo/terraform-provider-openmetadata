data "openmetadata_search_service" "elastic" {
  fully_qualified_name = "elastic"
}

output "elastic_search_service_id" {
  value = data.openmetadata_search_service.elastic.id
}
