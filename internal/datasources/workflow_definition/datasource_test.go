package workflow_definition

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

	if resp.TypeName != "openmetadata_workflow_definition" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_workflow_definition")
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
		"store_stage_status",
		"nodes",
		"edges",
		"trigger",
		"deployed",
		"suspended",
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
	displayName := "Contract approval"
	fqn := "contract_approval"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	deployed := true
	nodeName := "start"
	nodeType := "startEvent"
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
			{From: "start", To: "end"},
		},
		Trigger: &ometa.WorkflowTriggerInterface{
			Type: &triggerType,
		},
		Deployed:  &deployed,
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
		t.Fatal("Suspended = true, want false")
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

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	identifier := id

	var model workflowDefinitionModel
	apply(&ometa.WorkflowDefinition{Id: &identifier, Name: "bare", Description: ""}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.Nodes.IsNull() {
		t.Fatalf("Nodes = %v, want null", model.Nodes)
	}
	if !model.Edges.IsNull() {
		t.Fatalf("Edges = %v, want null", model.Edges)
	}
	if !model.Trigger.IsNull() {
		t.Fatalf("Trigger = %v, want null", model.Trigger)
	}
	if model.StoreStageStatus.ValueBool() {
		t.Fatal("StoreStageStatus = true, want false")
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}
