resource "openmetadata_messaging_service" "events" {
  name         = "events"
  service_type = "Kafka"
}

resource "openmetadata_topic" "order_created" {
  name        = "order_created"
  service     = openmetadata_messaging_service.events.fully_qualified_name
  description = "Emitted whenever an order is created."
  partitions  = 3

  retention_time   = 604800000
  retention_size   = 1073741824
  cleanup_policies = ["delete"]

  message_schema = jsonencode({
    schemaType = "JSON"
    schemaText = "{\"type\":\"object\",\"properties\":{\"orderId\":{\"type\":\"string\"}}}"
  })
}
