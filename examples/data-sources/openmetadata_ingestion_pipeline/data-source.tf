data "openmetadata_ingestion_pipeline" "warehouse_metadata" {
  fully_qualified_name = "warehouse.warehouse_metadata"
}

output "warehouse_metadata_pipeline_id" {
  value = data.openmetadata_ingestion_pipeline.warehouse_metadata.id
}
