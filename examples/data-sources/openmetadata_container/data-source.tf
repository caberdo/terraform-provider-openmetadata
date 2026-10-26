data "openmetadata_container" "events" {
  fully_qualified_name = "s3-bucket.raw.events"
}

output "events_container_id" {
  value = data.openmetadata_container.events.id
}
