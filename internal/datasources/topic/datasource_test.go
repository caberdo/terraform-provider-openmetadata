package topic

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

	if resp.TypeName != "openmetadata_topic" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_topic")
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

	for _, attribute := range []string{"message_schema", "topic_config"} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "events.order_created"
	serviceFQN := "events"
	schemaType := ometa.MessageSchemaSchemaTypeJSON
	schemaText := `{"type":"object"}`

	var model topicModel
	diags := apply(context.Background(), &ometa.Topic{
		Id:                 id,
		Name:               "order_created",
		FullyQualifiedName: &fqn,
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Name:               &serviceFQN,
		},
		Partitions: 3,
		CleanupPolicies: &[]ometa.TopicCleanupPolicies{
			ometa.TopicCleanupPoliciesDelete,
		},
		MessageSchema: &ometa.MessageSchema{
			SchemaType: &schemaType,
			SchemaText: &schemaText,
		},
		TopicConfig: &ometa.TopicConfig{"cleanup.policy": "delete"},
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
	if model.Partitions.ValueInt64() != 3 {
		t.Fatalf("Partitions = %d", model.Partitions.ValueInt64())
	}

	var policies []string
	if diags := model.CleanupPolicies.ElementsAs(context.Background(), &policies, false); diags.HasError() {
		t.Fatalf("cleanup policy decoding diagnostics = %v", diags.Errors())
	}
	if len(policies) != 1 || policies[0] != "delete" {
		t.Fatalf("cleanup policies = %v", policies)
	}
	if model.MessageSchema.IsNull() {
		t.Fatal("MessageSchema is null")
	}
	if model.TopicConfig.ValueString() != `{"cleanup.policy":"delete"}` {
		t.Fatalf("TopicConfig = %q", model.TopicConfig.ValueString())
	}
}

func TestApplyNullsMissingJSON(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model topicModel
	diags := apply(context.Background(), &ometa.Topic{
		Id:         id,
		Name:       "bare",
		Partitions: 1,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if !model.MessageSchema.IsNull() {
		t.Fatalf("MessageSchema = %v, want null", model.MessageSchema)
	}
	if !model.TopicConfig.IsNull() {
		t.Fatalf("TopicConfig = %v, want null", model.TopicConfig)
	}
	if !model.CleanupPolicies.IsNull() {
		t.Fatalf("CleanupPolicies = %v, want null", model.CleanupPolicies)
	}
}
