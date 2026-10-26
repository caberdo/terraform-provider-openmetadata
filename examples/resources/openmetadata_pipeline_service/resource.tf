resource "openmetadata_pipeline_service" "orchestrator" {
  name         = "orchestrator"
  display_name = "Orchestrator"
  description  = "Airflow deployment that runs the analytics pipelines."
  service_type = "Airflow"
}
