resource "openmetadata_pipeline_service" "orchestrator" {
  name         = "orchestrator"
  service_type = "Airflow"
}

resource "openmetadata_pipeline" "daily_orders" {
  name              = "daily_orders"
  service           = openmetadata_pipeline_service.orchestrator.fully_qualified_name
  display_name      = "Daily Orders"
  description       = "Builds the daily orders mart."
  schedule_interval = "0 2 * * *"
  state             = "Active"

  tasks = [
    {
      name        = "extract_orders"
      task_type   = "Extract"
      task_sql    = "select * from raw.orders"
      description = "Reads new orders from the landing zone."
    },
    {
      name             = "load_orders_mart"
      task_type        = "Load"
      task_sql         = "insert into mart.orders select * from staging.orders"
      downstream_tasks = ["publish_metrics"]
    },
  ]
}
