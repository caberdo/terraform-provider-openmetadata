resource "openmetadata_messaging_service" "events" {
  name         = "events"
  display_name = "Event Bus"
  description  = "Messaging backbone for domain events."
  service_type = "Kafka"
}
