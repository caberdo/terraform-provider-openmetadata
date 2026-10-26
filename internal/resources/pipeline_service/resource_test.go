package pipeline_service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_pipeline_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_pipeline_service")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatal("name must be a required string")
	}

	serviceType, ok := resp.Schema.Attributes["service_type"].(rschema.StringAttribute)
	if !ok || !serviceType.Required {
		t.Fatal("service_type must be a required string")
	}
	if len(serviceType.Validators) == 0 {
		t.Fatal("service_type must validate the connector type")
	}

	connection, ok := resp.Schema.Attributes["connection_config"].(rschema.StringAttribute)
	if !ok || !connection.Optional || !connection.Computed || !connection.Sensitive {
		t.Fatal("connection_config must be optional, computed and sensitive")
	}
	if len(connection.PlanModifiers) != 2 {
		t.Fatalf("connection plan modifiers = %d, want 2", len(connection.PlanModifiers))
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "orchestrator"
	displayName := "Orchestrator"
	description := "Airflow deployment."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model pipelineServiceModel
	apply(&ometa.PipelineService{
		Id:                 id,
		Name:               "orchestrator",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		ServiceType:        ometa.PipelineServiceServiceType("Airflow"),
		Connection: &ometa.PipelineConnection{Config: map[string]any{
			"hostPort": "http://airflow:8080",
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.ServiceType.ValueString() != "Airflow" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if !jsonEqual(model.Connection.ValueString(), `{"hostPort":"http://airflow:8080"}`) {
		t.Fatalf("Connection = %q", model.Connection.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := pipelineServiceModel{Connection: types.StringUnknown()}
	apply(&ometa.PipelineService{
		Id:          id,
		Name:        "bare",
		ServiceType: ometa.PipelineServiceServiceType("CustomPipeline"),
	}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.Connection.IsNull() {
		t.Fatalf("Connection = %v, want null", model.Connection)
	}
}

func TestApplyKeepsConnectionWhenServerOmitsIt(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	existing := `{"hostPort":"http://airflow:8080"}`

	model := pipelineServiceModel{Connection: types.StringValue(existing)}
	apply(&ometa.PipelineService{
		Id:          id,
		Name:        "orchestrator",
		ServiceType: ometa.PipelineServiceServiceType("Airflow"),
	}, &model)

	if model.Connection.ValueString() != existing {
		t.Fatalf("Connection = %q, want %q", model.Connection.ValueString(), existing)
	}
}

func TestPatchOps(t *testing.T) {
	base := pipelineServiceModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		ServiceType: types.StringValue("Airflow"),
		Connection:  types.StringValue(`{"hostPort":"http://airflow:8080"}`),
	}

	tests := []struct {
		name     string
		plan     pipelineServiceModel
		state    pipelineServiceModel
		wantOps  int
		wantPath string
		wantOp   string
	}{
		{name: "no changes", plan: base, state: base},
		{
			name: "display name changed",
			plan: pipelineServiceModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "add",
		},
		{
			name: "description changed",
			plan: pipelineServiceModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/description",
			wantOp:   "add",
		},
		{
			name: "connection changed",
			plan: pipelineServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"hostPort":"http://airflow-2:8080"}`),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "add",
		},
		{
			name: "service type cannot be patched",
			plan: pipelineServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: types.StringValue("Dagster"),
				Connection:  base.Connection,
			},
			state: base,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
			if len(ops) != tt.wantOps {
				t.Fatalf("ops = %d (%+v), want %d", len(ops), ops, tt.wantOps)
			}
			if tt.wantOps == 0 {
				return
			}
			if ops[0].Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", ops[0].Path, tt.wantPath)
			}
			if ops[0].Op != tt.wantOp {
				t.Fatalf("op = %q, want %q", ops[0].Op, tt.wantOp)
			}
		})
	}
}

func TestSemanticJSONPlanModifier(t *testing.T) {
	state := `{"hostPort":"http://airflow:8080"}`
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("{\n  \"hostPort\": \"http://airflow:8080\"\n}")}
	semanticJSONPlanModifier().PlanModifyString(context.Background(), planmodifier.StringRequest{
		PlanValue:  resp.PlanValue,
		StateValue: types.StringValue(state),
	}, resp)

	if resp.PlanValue.ValueString() != state {
		t.Fatalf("PlanValue = %q, want %q", resp.PlanValue.ValueString(), state)
	}
}
