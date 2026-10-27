# Import by UUID.
terraform import openmetadata_ingestion_pipeline.metadata 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, which is <service>.<pipeline name>.
terraform import openmetadata_ingestion_pipeline.metadata warehouse.warehouse_metadata
