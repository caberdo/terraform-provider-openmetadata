package ingestion_pipeline

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_ingestion_pipeline" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ingestion_pipeline")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "service", "pipeline_type"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	for _, attribute := range []string{"fully_qualified_name", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	for _, attribute := range []string{"deployed", "enabled"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.BoolAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed bool", attribute)
		}
	}

	for _, attribute := range []string{"source_config", "airflow_config"} {
		optional, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !optional.Optional || !optional.Computed {
			t.Fatalf("%s must be an optional computed string", attribute)
		}
		if optional.Default == nil {
			t.Fatalf("%s must default to an empty JSON object", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "warehouse.metadata"
	displayName := "Warehouse metadata"
	serviceFQN := "warehouse"
	description := "Ingests warehouse metadata."
	level := ometa.IngestionPipelineLoggerLevelINFO
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	deployed := true
	enabled := true
	streamable := true
	raiseOnError := true
	schedule := "0 * * * *"

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
		PipelineType:         ometa.IngestionPipelinePipelineTypeMetadata,
		SourceConfig:         ometa.SourceConfig{Config: map[string]any{"type": "DatabaseMetadata"}},
		AirflowConfig:        ometa.AirflowConfig{ScheduleInterval: &schedule},
		LoggerLevel:          &level,
		EnableStreamableLogs: &streamable,
		RaiseOnError:         &raiseOnError,
		Deployed:             &deployed,
		Enabled:              &enabled,
		Version:              &version,
		UpdatedAt:            &updatedAt,
		UpdatedBy:            &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "metadata" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
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
	if model.LoggerLevel.ValueString() != "INFO" {
		t.Fatalf("LoggerLevel = %q", model.LoggerLevel.ValueString())
	}
	if !model.EnableStreamable.ValueBool() {
		t.Fatal("EnableStreamable = false")
	}
	if !model.RaiseOnError.ValueBool() {
		t.Fatal("RaiseOnError = false")
	}
	if !model.Deployed.ValueBool() {
		t.Fatal("Deployed = false")
	}
	if !model.Enabled.ValueBool() {
		t.Fatal("Enabled = false")
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
	if model.UpdatedBy.ValueString() != updatedBy {
		t.Fatalf("UpdatedBy = %q", model.UpdatedBy.ValueString())
	}
}

func TestApplyKeepsPriorJSONWhenAPIOmits(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := ingestionPipelineModel{
		SourceConfig:  types.StringValue(`{"type":"DatabaseMetadata"}`),
		AirflowConfig: types.StringValue(`{"scheduleInterval":"0 * * * *"}`),
	}
	apply(&ometa.IngestionPipeline{Id: &id, Name: "bare"}, &model)

	if model.SourceConfig.ValueString() != `{"type":"DatabaseMetadata"}` {
		t.Fatalf("SourceConfig = %q, want the prior value", model.SourceConfig.ValueString())
	}
	if !jsonEqual(model.AirflowConfig.ValueString(), `{}`) {
		t.Fatalf("AirflowConfig = %q, want an empty object", model.AirflowConfig.ValueString())
	}
	if !model.LoggerLevel.IsNull() {
		t.Fatalf("LoggerLevel = %v, want null", model.LoggerLevel)
	}
	if model.Deployed.ValueBool() {
		t.Fatal("Deployed = true, want the false default")
	}
}

func TestCreateBody(t *testing.T) {
	serviceFQN := "warehouse"
	service := ometa.EntityReference{Type: "databaseService", FullyQualifiedName: &serviceFQN}

	body, err := createBody(ingestionPipelineModel{
		Name:             types.StringValue("metadata"),
		Service:          types.StringValue("warehouse"),
		ServiceType:      types.StringValue("databaseService"),
		PipelineType:     types.StringValue("metadata"),
		SourceConfig:     types.StringValue(`{"type":"DatabaseMetadata"}`),
		AirflowConfig:    types.StringValue(`{"scheduleInterval":"0 * * * *","retries":3}`),
		LoggerLevel:      types.StringValue("INFO"),
		EnableStreamable: types.BoolValue(true),
		RaiseOnError:     types.BoolValue(false),
	}, service)
	if err != nil {
		t.Fatalf("createBody() error = %v", err)
	}

	if body.Name != "metadata" {
		t.Fatalf("Name = %q", body.Name)
	}
	if body.Service.FullyQualifiedName == nil || *body.Service.FullyQualifiedName != "warehouse" {
		t.Fatalf("Service = %+v", body.Service)
	}
	if body.Service.Type != "databaseService" {
		t.Fatalf("Service.Type = %q", body.Service.Type)
	}
	if body.PipelineType != ometa.CreateIngestionPipelinePipelineTypeMetadata {
		t.Fatalf("PipelineType = %q", body.PipelineType)
	}
	if body.LoggerLevel == nil || *body.LoggerLevel != ometa.CreateIngestionPipelineLoggerLevelINFO {
		t.Fatalf("LoggerLevel = %v", body.LoggerLevel)
	}
	if body.EnableStreamableLogs == nil || !*body.EnableStreamableLogs {
		t.Fatal("EnableStreamableLogs not set")
	}
	if body.RaiseOnError == nil || *body.RaiseOnError {
		t.Fatal("RaiseOnError not set to false")
	}
	if body.AirflowConfig.ScheduleInterval == nil || *body.AirflowConfig.ScheduleInterval != "0 * * * *" {
		t.Fatalf("ScheduleInterval = %v", body.AirflowConfig.ScheduleInterval)
	}
	if body.AirflowConfig.Retries == nil || *body.AirflowConfig.Retries != 3 {
		t.Fatalf("Retries = %v", body.AirflowConfig.Retries)
	}

	if _, err := createBody(ingestionPipelineModel{
		Name:          types.StringValue("bad"),
		SourceConfig:  types.StringValue("{not json"),
		AirflowConfig: types.StringValue("{}"),
	}, service); err == nil {
		t.Fatal("createBody() with malformed source_config returned no error")
	}
}

func TestPatchOps(t *testing.T) {
	base := ingestionPipelineModel{
		DisplayName:      types.StringValue("Display"),
		Description:      types.StringValue("before"),
		SourceConfig:     types.StringValue(`{"type":"DatabaseMetadata"}`),
		AirflowConfig:    types.StringValue(`{"scheduleInterval":"0 * * * *"}`),
		LoggerLevel:      types.StringValue("INFO"),
		EnableStreamable: types.BoolValue(false),
		RaiseOnError:     types.BoolValue(false),
	}

	tests := []struct {
		name     string
		plan     ingestionPipelineModel
		state    ingestionPipelineModel
		wantOps  int
		wantPath string
		wantOp   string
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "reformatted source config is not a change",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     types.StringValue("{ \"type\" : \"DatabaseMetadata\" }"),
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: ingestionPipelineModel{
				DisplayName:      types.StringValue("New display"),
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "add",
		},
		{
			name: "display name removed",
			plan: ingestionPipelineModel{
				DisplayName:      types.StringNull(),
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      types.StringValue("after"),
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/description",
			wantOp:   "add",
		},
		{
			name: "source config changed",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     types.StringValue(`{"type":"DatabaseMetadata","includeTables":true}`),
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/sourceConfig",
			wantOp:   "add",
		},
		{
			name: "airflow config changed",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    types.StringValue(`{"scheduleInterval":"0 2 * * *"}`),
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/airflowConfig",
			wantOp:   "add",
		},
		{
			name: "logger level changed",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      types.StringValue("DEBUG"),
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/loggerLevel",
			wantOp:   "add",
		},
		{
			name: "logger level removed",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      types.StringNull(),
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/loggerLevel",
			wantOp:   "remove",
		},
		{
			name: "streamable logs enabled",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: types.BoolValue(true),
				RaiseOnError:     base.RaiseOnError,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/enableStreamableLogs",
			wantOp:   "add",
		},
		{
			name: "raise on error enabled",
			plan: ingestionPipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				SourceConfig:     base.SourceConfig,
				AirflowConfig:    base.AirflowConfig,
				LoggerLevel:      base.LoggerLevel,
				EnableStreamable: base.EnableStreamable,
				RaiseOnError:     types.BoolValue(true),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/raiseOnError",
			wantOp:   "add",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ops := patchOps(test.plan, test.state)
			if len(ops) != test.wantOps {
				t.Fatalf("patchOps() returned %d ops, want %d: %+v", len(ops), test.wantOps, ops)
			}
			if test.wantOps == 0 {
				return
			}
			if ops[0].Path != test.wantPath {
				t.Fatalf("ops[0].Path = %q, want %q", ops[0].Path, test.wantPath)
			}
			if ops[0].Op != test.wantOp {
				t.Fatalf("ops[0].Op = %q, want %q", ops[0].Op, test.wantOp)
			}
		})
	}
}
