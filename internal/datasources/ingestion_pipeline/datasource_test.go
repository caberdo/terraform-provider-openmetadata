package ingestion_pipeline

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_ingestion_pipeline" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ingestion_pipeline")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"name",
		"display_name",
		"description",
		"service",
		"service_type",
		"pipeline_type",
		"source_config",
		"airflow_config",
		"logger_level",
		"enable_streamable_logs",
		"raise_on_error",
		"deployed",
		"enabled",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "warehouse.metadata"
	serviceFQN := "warehouse"
	displayName := "Warehouse metadata"
	description := "Ingests warehouse metadata."
	level := ometa.IngestionPipelineLoggerLevelWARN
	schedule := "0 * * * *"
	deployed := true
	enabled := true

	var model ingestionPipelineModel
	apply(&ometa.IngestionPipeline{
		Id:                 &id,
		Name:               "metadata",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: &ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Type:               "databaseService",
		},
		PipelineType:  ometa.IngestionPipelinePipelineTypeMetadata,
		SourceConfig:  ometa.SourceConfig{Config: map[string]any{"type": "DatabaseMetadata"}},
		AirflowConfig: ometa.AirflowConfig{ScheduleInterval: &schedule},
		LoggerLevel:   &level,
		Deployed:      &deployed,
		Enabled:       &enabled,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.ServiceType.ValueString() != "databaseService" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if model.PipelineType.ValueString() != "metadata" {
		t.Fatalf("PipelineType = %q", model.PipelineType.ValueString())
	}
	if !jsonEqual(model.SourceConfig.ValueString(), `{"type":"DatabaseMetadata"}`) {
		t.Fatalf("SourceConfig = %q", model.SourceConfig.ValueString())
	}
	if !jsonEqual(model.AirflowConfig.ValueString(), `{"scheduleInterval":"0 * * * *"}`) {
		t.Fatalf("AirflowConfig = %q", model.AirflowConfig.ValueString())
	}
	if model.LoggerLevel.ValueString() != "WARN" {
		t.Fatalf("LoggerLevel = %q", model.LoggerLevel.ValueString())
	}
	if !model.Deployed.ValueBool() || !model.Enabled.ValueBool() {
		t.Fatalf("Deployed = %v, Enabled = %v", model.Deployed, model.Enabled)
	}
}
