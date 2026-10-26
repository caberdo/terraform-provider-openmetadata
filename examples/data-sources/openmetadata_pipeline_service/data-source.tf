data "openmetadata_pipeline_service" "orchestrator" {
  fully_qualified_name = "orchestrator"
}

output "orchestrator_service_id" {
  value = data.openmetadata_pipeline_service.orchestrator.id
}
