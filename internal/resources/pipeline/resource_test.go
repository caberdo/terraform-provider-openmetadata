package pipeline

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_pipeline" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_pipeline")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "service"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	concurrency, ok := resp.Schema.Attributes["concurrency"].(rschema.Int64Attribute)
	if !ok || !concurrency.Optional || !concurrency.Computed {
		t.Fatal("concurrency must be optional and computed")
	}

	state, ok := resp.Schema.Attributes["state"].(rschema.StringAttribute)
	if !ok || !state.Optional || !state.Computed || state.Default == nil {
		t.Fatal("state must be optional, computed and defaulted")
	}

	tasks, ok := resp.Schema.Attributes["tasks"].(rschema.ListNestedAttribute)
	if !ok || !tasks.Optional || !tasks.Computed {
		t.Fatal("tasks must be an optional computed list")
	}
	for _, attribute := range []string{"name", "task_type", "task_sql", "downstream_tasks"} {
		if _, ok := tasks.NestedObject.Attributes[attribute]; !ok {
			t.Fatalf("tasks is missing %s", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "orchestrator.daily_orders"
	serviceFQN := "orchestrator"
	displayName := "Daily Orders"
	description := "Builds the daily orders mart."
	concurrency := int32(2)
	location := "git://airflow/dags/orders.py"
	schedule := "0 2 * * *"
	sourceURL := "http://airflow:8080/dags/orders"
	startDate := int64(1_700_000_000_000)
	state := ometa.PipelineStateActive
	version := 0.4
	updatedAt := int64(1_700_000_100_000)
	updatedBy := "admin"
	taskType := "Extract"
	taskSQL := "select * from raw.orders"
	taskDescription := "Reads orders."

	var model pipelineModel
	diags := apply(context.Background(), &ometa.Pipeline{
		Id:                 id,
		Name:               "daily_orders",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Name:               &serviceFQN,
		},
		Concurrency:      &concurrency,
		PipelineLocation: &location,
		ScheduleInterval: &schedule,
		SourceUrl:        &sourceURL,
		StartDate:        &startDate,
		State:            &state,
		Tasks: &[]ometa.Task{{
			Name:        "extract_orders",
			Description: &taskDescription,
			TaskType:    &taskType,
			TaskSQL:     &taskSQL,
			DownstreamTasks: &[]string{
				"load_orders_mart",
			},
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "daily_orders" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.Concurrency.ValueInt64() != int64(concurrency) {
		t.Fatalf("Concurrency = %d", model.Concurrency.ValueInt64())
	}
	if model.PipelineLocation.ValueString() != location {
		t.Fatalf("PipelineLocation = %q", model.PipelineLocation.ValueString())
	}
	if model.ScheduleInterval.ValueString() != schedule {
		t.Fatalf("ScheduleInterval = %q", model.ScheduleInterval.ValueString())
	}
	if model.SourceURL.ValueString() != sourceURL {
		t.Fatalf("SourceURL = %q", model.SourceURL.ValueString())
	}
	if model.StartDate.ValueInt64() != startDate {
		t.Fatalf("StartDate = %d", model.StartDate.ValueInt64())
	}
	if model.State.ValueString() != "Active" {
		t.Fatalf("State = %q", model.State.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:15:00Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}

	var tasks []pipelineTaskModel
	if diags := model.Tasks.ElementsAs(context.Background(), &tasks, false); diags.HasError() {
		t.Fatalf("tasks decoding diagnostics = %v", diags.Errors())
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(tasks))
	}
	if tasks[0].Name.ValueString() != "extract_orders" {
		t.Fatalf("task name = %q", tasks[0].Name.ValueString())
	}
	if tasks[0].TaskType.ValueString() != "Extract" {
		t.Fatalf("task type = %q", tasks[0].TaskType.ValueString())
	}
	if tasks[0].TaskSQL.ValueString() != taskSQL {
		t.Fatalf("task sql = %q", tasks[0].TaskSQL.ValueString())
	}

	var downstream []string
	if diags := tasks[0].DownstreamTasks.ElementsAs(context.Background(), &downstream, false); diags.HasError() {
		t.Fatalf("downstream decoding diagnostics = %v", diags.Errors())
	}
	if len(downstream) != 1 || downstream[0] != "load_orders_mart" {
		t.Fatalf("downstream tasks = %v", downstream)
	}
}

func TestApplyKeepsServiceAndTasksWhenServerOmitsThem(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	tasks := types.ListValueMust(pipelineTaskObjectType, []attr.Value{
		types.ObjectValueMust(pipelineTaskObjectType.AttrTypes, map[string]attr.Value{
			"name":             types.StringValue("extract_orders"),
			"display_name":     types.StringNull(),
			"description":      types.StringNull(),
			"task_type":        types.StringNull(),
			"task_sql":         types.StringNull(),
			"source_url":       types.StringNull(),
			"start_date":       types.StringNull(),
			"end_date":         types.StringNull(),
			"downstream_tasks": types.ListNull(types.StringType),
		}),
	})

	model := pipelineModel{
		Service: types.StringValue("orchestrator"),
		Tasks:   tasks,
	}
	if diags := apply(context.Background(), &ometa.Pipeline{
		Id:   id,
		Name: "daily_orders",
	}, &model); diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.Service.ValueString() != "orchestrator" {
		t.Fatalf("Service = %q, want the prior selector", model.Service.ValueString())
	}
	if !model.Tasks.Equal(tasks) {
		t.Fatal("tasks were zeroed when the API omitted them")
	}
}

func TestTasksRoundTrip(t *testing.T) {
	displayName := "Extract Orders"
	description := "Reads orders."
	taskType := "Extract"
	taskSQL := "select 1"
	sourceURL := "http://airflow/task"
	startDate := "2026-01-01T00:00:00Z"
	endDate := "2026-01-01T01:00:00Z"
	downstream := []string{"load"}

	apiTasks := []ometa.Task{{
		Name:            "extract",
		DisplayName:     &displayName,
		Description:     &description,
		TaskType:        &taskType,
		TaskSQL:         &taskSQL,
		SourceUrl:       &sourceURL,
		StartDate:       &startDate,
		EndDate:         &endDate,
		DownstreamTasks: &downstream,
	}}

	list, diags := tasksValue(context.Background(), apiTasks)
	if diags.HasError() {
		t.Fatalf("tasksValue() diagnostics = %v", diags.Errors())
	}

	roundTripped, diags := tasksToAPI(context.Background(), list)
	if diags.HasError() {
		t.Fatalf("tasksToAPI() diagnostics = %v", diags.Errors())
	}
	if len(roundTripped) != 1 {
		t.Fatalf("tasks = %d, want 1", len(roundTripped))
	}
	if roundTripped[0].Name != "extract" {
		t.Fatalf("name = %q", roundTripped[0].Name)
	}
	if roundTripped[0].DisplayName == nil || *roundTripped[0].DisplayName != displayName {
		t.Fatalf("display name = %v", roundTripped[0].DisplayName)
	}
	if roundTripped[0].DownstreamTasks == nil || len(*roundTripped[0].DownstreamTasks) != 1 {
		t.Fatalf("downstream tasks = %v", roundTripped[0].DownstreamTasks)
	}
}

func TestPatchOps(t *testing.T) {
	emptyTasks := types.ListNull(pipelineTaskObjectType)
	base := pipelineModel{
		DisplayName:      types.StringValue("Display"),
		Description:      types.StringValue("before"),
		Concurrency:      types.Int64Value(1),
		PipelineLocation: types.StringValue("git://airflow/dags/orders.py"),
		ScheduleInterval: types.StringValue("0 2 * * *"),
		State:            types.StringValue("Active"),
		Tasks:            emptyTasks,
	}

	tests := []struct {
		name     string
		plan     pipelineModel
		state    pipelineModel
		wantOps  int
		wantPath string
		wantOp   string
	}{
		{name: "no changes", plan: base, state: base},
		{
			name: "description changed",
			plan: pipelineModel{
				DisplayName:      base.DisplayName,
				Description:      types.StringValue("after"),
				Concurrency:      base.Concurrency,
				PipelineLocation: base.PipelineLocation,
				ScheduleInterval: base.ScheduleInterval,
				State:            base.State,
				Tasks:            emptyTasks,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/description",
			wantOp:   "add",
		},
		{
			name: "concurrency changed",
			plan: pipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				Concurrency:      types.Int64Value(3),
				PipelineLocation: base.PipelineLocation,
				ScheduleInterval: base.ScheduleInterval,
				State:            base.State,
				Tasks:            emptyTasks,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/concurrency",
			wantOp:   "add",
		},
		{
			name: "location removed",
			plan: pipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				Concurrency:      base.Concurrency,
				PipelineLocation: types.StringNull(),
				ScheduleInterval: base.ScheduleInterval,
				State:            base.State,
				Tasks:            emptyTasks,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/pipelineLocation",
			wantOp:   "remove",
		},
		{
			name: "state changed",
			plan: pipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				Concurrency:      base.Concurrency,
				PipelineLocation: base.PipelineLocation,
				ScheduleInterval: base.ScheduleInterval,
				State:            types.StringValue("Inactive"),
				Tasks:            emptyTasks,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/state",
			wantOp:   "add",
		},
		{
			name: "unknown tasks emit no op",
			plan: pipelineModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				Concurrency:      base.Concurrency,
				PipelineLocation: base.PipelineLocation,
				ScheduleInterval: base.ScheduleInterval,
				State:            base.State,
				Tasks:            types.ListUnknown(pipelineTaskObjectType),
			},
			state: base,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, diags := patchOps(context.Background(), tt.plan, tt.state)
			if diags.HasError() {
				t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
			}
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

func TestStatePointer(t *testing.T) {
	if pointer := statePointer(types.StringNull()); pointer != nil {
		t.Fatalf("statePointer(null) = %v, want nil", pointer)
	}
	pointer := statePointer(types.StringValue("Inactive"))
	if pointer == nil || *pointer != ometa.CreatePipelineStateInactive {
		t.Fatalf("statePointer(Inactive) = %v", pointer)
	}
}
