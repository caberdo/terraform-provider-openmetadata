package topic

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_topic" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_topic")
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

	partitions, ok := resp.Schema.Attributes["partitions"].(rschema.Int64Attribute)
	if !ok || !partitions.Required {
		t.Fatal("partitions must be a required number")
	}

	cleanup, ok := resp.Schema.Attributes["cleanup_policies"].(rschema.ListAttribute)
	if !ok || !cleanup.Optional || !cleanup.Computed {
		t.Fatal("cleanup_policies must be optional and computed")
	}
	if len(cleanup.Validators) == 0 {
		t.Fatal("cleanup_policies must validate the policy names")
	}

	for _, attribute := range []string{"message_schema", "topic_config"} {
		jsonAttribute, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !jsonAttribute.Optional || !jsonAttribute.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if len(jsonAttribute.PlanModifiers) != 2 {
			t.Fatalf("%s plan modifiers = %d, want 2", attribute, len(jsonAttribute.PlanModifiers))
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "events.order_created"
	serviceFQN := "events"
	displayName := "Order Created"
	description := "Emitted when an order is created."
	maximumMessageSize := int32(1048588)
	minimumInSyncReplicas := int32(1)
	replicationFactor := int32(2)
	retentionSize := 1073741824.0
	retentionTime := 604800000.0
	version := 0.5
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	schemaText := `{"type":"object"}`
	schemaType := ometa.MessageSchemaSchemaTypeJSON

	var model topicModel
	diags := apply(context.Background(), &ometa.Topic{
		Id:                 id,
		Name:               "order_created",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Name:               &serviceFQN,
		},
		Partitions:            3,
		MaximumMessageSize:    &maximumMessageSize,
		MinimumInSyncReplicas: &minimumInSyncReplicas,
		ReplicationFactor:     &replicationFactor,
		RetentionSize:         &retentionSize,
		RetentionTime:         &retentionTime,
		CleanupPolicies: &[]ometa.TopicCleanupPolicies{
			ometa.TopicCleanupPoliciesDelete,
		},
		MessageSchema: &ometa.MessageSchema{
			SchemaType: &schemaType,
			SchemaText: &schemaText,
		},
		TopicConfig: &ometa.TopicConfig{"cleanup.policy": "delete"},
		Version:     &version,
		UpdatedAt:   &updatedAt,
		UpdatedBy:   &updatedBy,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.Partitions.ValueInt64() != 3 {
		t.Fatalf("Partitions = %d", model.Partitions.ValueInt64())
	}
	if model.MaximumMessageSize.ValueInt64() != int64(maximumMessageSize) {
		t.Fatalf("MaximumMessageSize = %d", model.MaximumMessageSize.ValueInt64())
	}
	if model.RetentionTime.ValueFloat64() != retentionTime {
		t.Fatalf("RetentionTime = %f", model.RetentionTime.ValueFloat64())
	}

	var policies []string
	if diags := model.CleanupPolicies.ElementsAs(context.Background(), &policies, false); diags.HasError() {
		t.Fatalf("cleanup policy decoding diagnostics = %v", diags.Errors())
	}
	if len(policies) != 1 || policies[0] != "delete" {
		t.Fatalf("cleanup policies = %v", policies)
	}
	if !jsonEqual(model.MessageSchema.ValueString(), `{"schemaText":"{\"type\":\"object\"}","schemaType":"JSON"}`) {
		t.Fatalf("MessageSchema = %q", model.MessageSchema.ValueString())
	}
	if !jsonEqual(model.TopicConfig.ValueString(), `{"cleanup.policy":"delete"}`) {
		t.Fatalf("TopicConfig = %q", model.TopicConfig.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyKeepsJSONAndServiceWhenServerOmitsThem(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	existingSchema := `{"schemaType":"JSON"}`
	existingConfig := `{"cleanup.policy":"delete"}`

	model := topicModel{
		Service:       types.StringValue("events"),
		MessageSchema: types.StringValue(existingSchema),
		TopicConfig:   types.StringValue(existingConfig),
	}
	if diags := apply(context.Background(), &ometa.Topic{
		Id:         id,
		Name:       "order_created",
		Partitions: 3,
	}, &model); diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.Service.ValueString() != "events" {
		t.Fatalf("Service = %q, want the prior selector", model.Service.ValueString())
	}
	if model.MessageSchema.ValueString() != existingSchema {
		t.Fatalf("MessageSchema = %q", model.MessageSchema.ValueString())
	}
	if model.TopicConfig.ValueString() != existingConfig {
		t.Fatalf("TopicConfig = %q", model.TopicConfig.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	policies := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("delete"),
	})
	base := topicModel{
		DisplayName:     types.StringValue("Display"),
		Description:     types.StringValue("before"),
		Partitions:      types.Int64Value(3),
		CleanupPolicies: policies,
		MessageSchema:   types.StringValue(`{"schemaType":"JSON"}`),
		TopicConfig:     types.StringValue(`{"cleanup.policy":"delete"}`),
	}

	tests := []struct {
		name     string
		plan     topicModel
		state    topicModel
		wantOps  int
		wantPath string
		wantOp   string
	}{
		{name: "no changes", plan: base, state: base},
		{
			name: "description changed",
			plan: topicModel{
				DisplayName:     base.DisplayName,
				Description:     types.StringValue("after"),
				Partitions:      base.Partitions,
				CleanupPolicies: base.CleanupPolicies,
				MessageSchema:   base.MessageSchema,
				TopicConfig:     base.TopicConfig,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/description",
			wantOp:   "add",
		},
		{
			name: "partitions changed",
			plan: topicModel{
				DisplayName:     base.DisplayName,
				Description:     base.Description,
				Partitions:      types.Int64Value(6),
				CleanupPolicies: base.CleanupPolicies,
				MessageSchema:   base.MessageSchema,
				TopicConfig:     base.TopicConfig,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/partitions",
			wantOp:   "add",
		},
		{
			name: "cleanup policies changed",
			plan: topicModel{
				DisplayName:     base.DisplayName,
				Description:     base.Description,
				Partitions:      base.Partitions,
				CleanupPolicies: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("compact")}),
				MessageSchema:   base.MessageSchema,
				TopicConfig:     base.TopicConfig,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/cleanupPolicies",
			wantOp:   "add",
		},
		{
			name: "message schema reformatted but equal",
			plan: topicModel{
				DisplayName:     base.DisplayName,
				Description:     base.Description,
				Partitions:      base.Partitions,
				CleanupPolicies: base.CleanupPolicies,
				MessageSchema:   types.StringValue("{\n  \"schemaType\": \"JSON\"\n}"),
				TopicConfig:     base.TopicConfig,
			},
			state: base,
		},
		{
			name: "topic config changed",
			plan: topicModel{
				DisplayName:     base.DisplayName,
				Description:     base.Description,
				Partitions:      base.Partitions,
				CleanupPolicies: base.CleanupPolicies,
				MessageSchema:   base.MessageSchema,
				TopicConfig:     types.StringValue(`{"cleanup.policy":"compact"}`),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/topicConfig",
			wantOp:   "add",
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

func TestJSONEqual(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{name: "key order ignored", left: `{"a":1,"b":{"c":2}}`, right: `{"b":{"c":2},"a":1}`, want: true},
		{name: "whitespace ignored", left: "{\n  \"a\": 1\n}", right: `{"a":1}`, want: true},
		{name: "different values", left: `{"a":1}`, right: `{"a":2}`, want: false},
		{name: "invalid left", left: "not json", right: `{"a":1}`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonEqual(tt.left, tt.right); got != tt.want {
				t.Fatalf("jsonEqual(%q, %q) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestSemanticJSONPlanModifier(t *testing.T) {
	state := `{"schemaType":"JSON"}`
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("{\n  \"schemaType\": \"JSON\"\n}")}
	semanticJSONPlanModifier().PlanModifyString(context.Background(), planmodifier.StringRequest{
		PlanValue:  resp.PlanValue,
		StateValue: types.StringValue(state),
	}, resp)

	if resp.PlanValue.ValueString() != state {
		t.Fatalf("PlanValue = %q, want %q", resp.PlanValue.ValueString(), state)
	}
}

func TestParseJSONRejectsInvalidDocuments(t *testing.T) {
	if _, err := parseMessageSchemaJSON(`{"broken"`); err == nil {
		t.Fatal("parseMessageSchemaJSON accepted invalid JSON")
	}
	if _, err := parseTopicConfigJSON(`{"broken"`); err == nil {
		t.Fatal("parseTopicConfigJSON accepted invalid JSON")
	}
	if jsonConfigured(types.StringValue("  ")) {
		t.Fatal("jsonConfigured accepted an empty value")
	}
}
