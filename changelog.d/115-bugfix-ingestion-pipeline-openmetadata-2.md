`openmetadata_ingestion_pipeline` and the matching data source now work against
OpenMetadata 2.0, which returns `pipelineStatuses` as an array. The provider
ignores that field when decoding a pipeline, so create, read, update, import
and lookups all succeed on 1.13 and 2.0.
