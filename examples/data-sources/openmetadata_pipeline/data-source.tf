data "openmetadata_pipeline" "daily_orders" {
  fully_qualified_name = "orchestrator.daily_orders"
}

output "daily_orders_task_names" {
  value = [for task in data.openmetadata_pipeline.daily_orders.tasks : task.name]
}
