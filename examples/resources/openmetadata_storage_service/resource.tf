resource "openmetadata_storage_service" "data_lake" {
  name         = "data-lake"
  service_type = "S3"
  display_name = "Data Lake"
  description  = "Storage service for the analytics data lake."
}
