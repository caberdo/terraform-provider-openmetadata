resource "openmetadata_search_service" "elastic" {
  name         = "elastic"
  service_type = "ElasticSearch"
  display_name = "Elasticsearch"
  description  = "Search service backing the platform's index store."
}
