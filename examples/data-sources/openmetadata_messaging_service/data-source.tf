data "openmetadata_messaging_service" "events" {
  fully_qualified_name = "events"
}

output "events_service_id" {
  value = data.openmetadata_messaging_service.events.id
}
