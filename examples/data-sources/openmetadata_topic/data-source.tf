data "openmetadata_topic" "order_created" {
  fully_qualified_name = "events.order_created"
}

output "order_created_partitions" {
  value = data.openmetadata_topic.order_created.partitions
}
