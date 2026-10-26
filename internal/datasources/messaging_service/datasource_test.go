package messaging_service

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

	if resp.TypeName != "openmetadata_messaging_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_messaging_service")
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

	connection, ok := resp.Schema.Attributes["connection_config"].(dschema.StringAttribute)
	if !ok || !connection.Computed || !connection.Sensitive {
		t.Fatal("connection_config must be a computed sensitive string")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "events"
	description := "Messaging backbone."
	version := 0.7
	updatedAt := int64(1_700_000_000_000)

	var model messagingServiceModel
	apply(&ometa.MessagingService{
		Id:                 id,
		Name:               "events",
		FullyQualifiedName: &fqn,
		Description:        &description,
		ServiceType:        ometa.MessagingServiceServiceType("Kafka"),
		Connection: &ometa.MessagingConnection{Config: map[string]any{
			"bootstrapServers": "kafka:9092",
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.ServiceType.ValueString() != "Kafka" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if model.Connection.ValueString() != `{"bootstrapServers":"kafka:9092"}` {
		t.Fatalf("Connection = %q", model.Connection.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyNullsMissingConnection(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model messagingServiceModel
	apply(&ometa.MessagingService{
		Id:          id,
		Name:        "bare",
		ServiceType: ometa.MessagingServiceServiceType("CustomMessaging"),
	}, &model)

	if !model.Connection.IsNull() {
		t.Fatalf("Connection = %v, want null", model.Connection)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}
