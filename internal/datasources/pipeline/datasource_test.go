package pipeline

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_pipeline" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_pipeline")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be an optional computed selector", attribute)
		}
		if len(selector.Validators) == 0 {
			t.Fatalf("%s must validate the exactly-one-of rule", attribute)
		}
	}

	tasks, ok := resp.Schema.Attributes["tasks"].(dschema.ListNestedAttribute)
	if !ok || !tasks.Computed {
		t.Fatal("tasks must be a computed list")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "orchestrator.daily_orders"
	serviceFQN := "orchestrator"
	schedule := "0 2 * * *"
	state := ometa.PipelineStateActive
	taskType := "Extract"

	var model pipelineModel
	diags := apply(context.Background(), &ometa.Pipeline{
		Id:                 id,
		Name:               "daily_orders",
		FullyQualifiedName: &fqn,
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Name:               &serviceFQN,
		},
		ScheduleInterval: &schedule,
		State:            &state,
		Tasks: &[]ometa.Task{{
			Name:     "extract_orders",
			TaskType: &taskType,
		}},
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.ScheduleInterval.ValueString() != schedule {
		t.Fatalf("ScheduleInterval = %q", model.ScheduleInterval.ValueString())
	}
	if model.State.ValueString() != "Active" {
		t.Fatalf("State = %q", model.State.ValueString())
	}

	var tasks []pipelineTaskModel
	if diags := model.Tasks.ElementsAs(context.Background(), &tasks, false); diags.HasError() {
		t.Fatalf("tasks decoding diagnostics = %v", diags.Errors())
	}
	if len(tasks) != 1 || tasks[0].Name.ValueString() != "extract_orders" {
		t.Fatalf("tasks = %+v", tasks)
	}
}

func TestApplyNullsMissingTasks(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model pipelineModel
	diags := apply(context.Background(), &ometa.Pipeline{
		Id:   id,
		Name: "bare",
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if !model.Tasks.IsNull() {
		t.Fatalf("Tasks = %v, want null", model.Tasks)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}
