resource "openmetadata_storage_service" "buckets" {
  name         = "s3-bucket"
  service_type = "S3"
}

resource "openmetadata_container" "raw" {
  name        = "raw"
  service     = openmetadata_storage_service.buckets.name
  description = "Raw event archive."
}

resource "openmetadata_container" "events" {
  name         = "events"
  service      = openmetadata_storage_service.buckets.name
  parent       = openmetadata_container.raw.fully_qualified_name
  display_name = "Raw events"
  description  = "Raw events partitioned by day."
}
