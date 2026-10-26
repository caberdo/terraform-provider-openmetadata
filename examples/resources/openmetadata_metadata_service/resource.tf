resource "openmetadata_metadata_service" "elastic" {
  name         = "elastic-metadata"
  service_type = "MetadataES"
  display_name = "Elasticsearch metadata"
  description  = "Metadata service for the external metadata source."
}
