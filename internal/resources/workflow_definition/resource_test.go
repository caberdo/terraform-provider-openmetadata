package workflow_definition

import (
	"context"
	"reflect"
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

	if resp.TypeName != "openmetadata_workflow_definition" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_workflow_definition")
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
		t.Fatalf("name attribute = %#v, want required StringAttribute", resp.Schema.Attributes["name"])
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Required {
		t.Fatalf("description attribute = %#v, want required StringAttribute", resp.Schema.Attributes["description"])
	}

	storeStageStatus, ok := resp.Schema.Attributes["store_stage_status"].(rschema.BoolAttribute)
	if !ok || !storeStageStatus.Optional || !storeStageStatus.Computed || storeStageStatus.Default == nil {
		t.Fatalf("store_stage_status attribute = %#v, want optional+computed bool with a default", resp.Schema.Attributes["store_stage_status"])
	}

	for _, attribute := range []string{"nodes", "edges", "trigger"} {
		jsonAttribute, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !jsonAttribute.Optional {
			t.Fatalf("%s attribute = %#v, want optional StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	for _, attribute := range []string{"deployed", "suspended"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.BoolAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed bool", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	displayName := "Contract approval"
	fqn := "contract_approval"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	deployed := true
	suspended := false
	nodeName := "start"
	nodeType := "startEvent"
	edgeFrom := "start"
	edgeTo := "end"
	triggerType := "eventBasedEntity"
	identifier := id

	var model workflowDefinitionModel
	apply(&ometa.WorkflowDefinition{
		Id:                 &identifier,
		Name:               "contract_approval",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        "Approval flow for data contracts.",
		Config:             &ometa.WorkflowConfiguration{StoreStageStatus: true},
		Nodes: &[]ometa.WorkflowNodeDefinitionInterface{
			{Name: &nodeName, Type: &nodeType},
		},
		Edges: &[]ometa.EdgeDefinition{
			{From: edgeFrom, To: edgeTo},
		},
		Trigger: &ometa.WorkflowTriggerInterface{
			Type: &triggerType,
		},
		Deployed:  &deployed,
		Suspended: &suspended,
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "contract_approval" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Approval flow for data contracts." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if !model.StoreStageStatus.ValueBool() {
		t.Fatal("StoreStageStatus = false")
	}
	if model.Nodes.ValueString() != `[{"name":"start","type":"startEvent"}]` {
		t.Fatalf("Nodes = %q", model.Nodes.ValueString())
	}
	if model.Edges.ValueString() != `[{"from":"start","to":"end"}]` {
		t.Fatalf("Edges = %q", model.Edges.ValueString())
	}
	if model.Trigger.ValueString() != `{"type":"eventBasedEntity"}` {
		t.Fatalf("Trigger = %q", model.Trigger.ValueString())
	}
	if !model.Deployed.ValueBool() {
		t.Fatal("Deployed = false")
	}
	if model.Suspended.ValueBool() {
		t.Fatal("Suspended = true")
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

func TestApplyKeepsPriorFieldsWhenAPIOmitsThem(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	identifier := id

	model := workflowDefinitionModel{
		Type:             types.StringValue("noOp"),
		StoreStageStatus: types.BoolValue(true),
		Nodes:            types.StringValue(`[{"name":"start","type":"startEvent"}]`),
		Edges:            types.StringValue(`[{"from":"start","to":"end"}]`),
		Trigger:          types.StringValue(`{"type":"eventBasedEntity"}`),
	}
	apply(&ometa.WorkflowDefinition{Id: &identifier, Name: "noop", Description: ""}, &model)

	if model.Type.ValueString() != "noOp" {
		t.Fatalf("Type = %q, want the configured value", model.Type.ValueString())
	}
	if !model.StoreStageStatus.ValueBool() {
		t.Fatal("StoreStageStatus = false, want the prior value")
	}
	if model.Nodes.ValueString() != `[{"name":"start","type":"startEvent"}]` {
		t.Fatalf("Nodes = %q, want the prior value", model.Nodes.ValueString())
	}
	if model.Edges.ValueString() != `[{"from":"start","to":"end"}]` {
		t.Fatalf("Edges = %q, want the prior value", model.Edges.ValueString())
	}
	if model.Trigger.ValueString() != `{"type":"eventBasedEntity"}` {
		t.Fatalf("Trigger = %q, want the prior value", model.Trigger.ValueString())
	}
	if model.Deployed.ValueBool() || model.Suspended.ValueBool() {
		t.Fatal("Deployed and Suspended must default to false")
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestApplyKeepsEquivalentJSONVerbatim(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	identifier := id
	nodeName := "start"
	nodeType := "startEvent"

	model := workflowDefinitionModel{
		Nodes: types.StringValue("[\n  {\"type\": \"startEvent\", \"name\": \"start\"}\n]"),
	}
	apply(&ometa.WorkflowDefinition{
		Id:          &identifier,
		Name:        "noop",
		Description: "",
		Nodes: &[]ometa.WorkflowNodeDefinitionInterface{
			{Name: &nodeName, Type: &nodeType},
		},
	}, &model)

	if model.Nodes.ValueString() != "[\n  {\"type\": \"startEvent\", \"name\": \"start\"}\n]" {
		t.Fatalf("Nodes = %q, want the configured formatting", model.Nodes.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := workflowDefinitionModel{
		DisplayName:      types.StringValue("Display"),
		Description:      types.StringValue("before"),
		StoreStageStatus: types.BoolValue(false),
		Nodes:            types.StringValue(`[{"name":"start","type":"startEvent"}]`),
		Edges:            types.StringValue(`[{"from":"start","to":"end"}]`),
		Trigger:          types.StringNull(),
	}

	tests := []struct {
		name      string
		plan      workflowDefinitionModel
		state     workflowDefinitionModel
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "equivalent JSON is not patched",
			plan: workflowDefinitionModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            types.StringValue("[{\"type\":\"startEvent\",\"name\":\"start\"}]"),
				Edges:            base.Edges,
				Trigger:          base.Trigger,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: workflowDefinitionModel{
				DisplayName:      types.StringValue("New display"),
				Description:      base.Description,
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            base.Nodes,
				Edges:            base.Edges,
				Trigger:          base.Trigger,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: workflowDefinitionModel{
				DisplayName:      types.StringNull(),
				Description:      base.Description,
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            base.Nodes,
				Edges:            base.Edges,
				Trigger:          base.Trigger,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: workflowDefinitionModel{
				DisplayName:      base.DisplayName,
				Description:      types.StringValue("after"),
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            base.Nodes,
				Edges:            base.Edges,
				Trigger:          base.Trigger,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "store stage status changed",
			plan: workflowDefinitionModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				StoreStageStatus: types.BoolValue(true),
				Nodes:            base.Nodes,
				Edges:            base.Edges,
				Trigger:          base.Trigger,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/config",
			wantOp:    "add",
			wantValue: &ometa.WorkflowConfiguration{StoreStageStatus: true},
		},
		{
			name: "nodes changed",
			plan: workflowDefinitionModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            types.StringValue(`[{"name":"start","type":"startEvent"},{"name":"end","type":"endEvent"}]`),
				Edges:            base.Edges,
				Trigger:          base.Trigger,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/nodes",
			wantOp:   "add",
		},
		{
			name: "edges removed",
			plan: workflowDefinitionModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            base.Nodes,
				Edges:            types.StringNull(),
				Trigger:          base.Trigger,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/edges",
			wantOp:   "remove",
		},
		{
			name: "trigger added",
			plan: workflowDefinitionModel{
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				StoreStageStatus: base.StoreStageStatus,
				Nodes:            base.Nodes,
				Edges:            base.Edges,
				Trigger:          types.StringValue(`{"type":"eventBasedEntity"}`),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/trigger",
			wantOp:   "add",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
			if len(ops) != tt.wantOps {
				t.Fatalf("len(ops) = %d, want %d: %#v", len(ops), tt.wantOps, ops)
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
			if tt.wantValue != nil && !reflect.DeepEqual(ops[0].Value, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
