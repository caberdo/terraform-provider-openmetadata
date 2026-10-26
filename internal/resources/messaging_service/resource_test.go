package messaging_service

import (
	"context"
	"encoding/json"
	"reflect"
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

	if resp.TypeName != "openmetadata_messaging_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_messaging_service")
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

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "events"
	displayName := "Event Bus"
	description := "Messaging backbone."
	version := 0.3
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model messagingServiceModel
	apply(&ometa.MessagingService{
		Id:                 id,
		Name:               "events",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		ServiceType:        ometa.MessagingServiceServiceType("Kafka"),
		Connection: &ometa.MessagingConnection{Config: map[string]any{
			"bootstrapServers": "kafka:9092",
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "events" {
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
	if model.ServiceType.ValueString() != "Kafka" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if !jsonEqual(model.Connection.ValueString(), `{"bootstrapServers":"kafka:9092"}`) {
		t.Fatalf("Connection = %q", model.Connection.ValueString())
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

	model := messagingServiceModel{Connection: types.StringUnknown()}
	apply(&ometa.MessagingService{
		Id:          id,
		Name:        "bare",
		ServiceType: ometa.MessagingServiceServiceType("CustomMessaging"),
	}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.Connection.IsNull() {
		t.Fatalf("Connection = %v, want null", model.Connection)
	}
}

func TestApplyKeepsConnectionWhenServerOmitsIt(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	existing := `{"bootstrapServers":"kafka:9092"}`

	model := messagingServiceModel{Connection: types.StringValue(existing)}
	apply(&ometa.MessagingService{
		Id:          id,
		Name:        "events",
		ServiceType: ometa.MessagingServiceServiceType("Kafka"),
	}, &model)

	if model.Connection.ValueString() != existing {
		t.Fatalf("Connection = %q, want %q", model.Connection.ValueString(), existing)
	}
}

func TestApplyPreservesEquivalentConnection(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	formatted := "{\n  \"bootstrapServers\": \"kafka:9092\"\n}"

	model := messagingServiceModel{Connection: types.StringValue(formatted)}
	apply(&ometa.MessagingService{
		Id:          id,
		Name:        "events",
		ServiceType: ometa.MessagingServiceServiceType("Kafka"),
		Connection: &ometa.MessagingConnection{Config: map[string]any{
			"bootstrapServers": "kafka:9092",
		}},
	}, &model)

	if model.Connection.ValueString() != formatted {
		t.Fatalf("Connection = %q, want the formatted state value", model.Connection.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := messagingServiceModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		ServiceType: types.StringValue("Kafka"),
		Connection:  types.StringValue(`{"bootstrapServers":"kafka:9092"}`),
	}

	tests := []struct {
		name      string
		plan      messagingServiceModel
		state     messagingServiceModel
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{name: "no changes", plan: base, state: base},
		{
			name: "display name changed",
			plan: messagingServiceModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "description changed",
			plan: messagingServiceModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "connection changed",
			plan: messagingServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"bootstrapServers":"kafka-2:9092"}`),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "add",
			wantValue: &ometa.MessagingConnection{Config: map[string]any{
				"bootstrapServers": "kafka-2:9092",
			}},
		},
		{
			name: "connection reformatted but equal",
			plan: messagingServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue("{\n  \"bootstrapServers\": \"kafka:9092\"\n}"),
			},
			state: base,
		},
		{
			name: "unknown connection emits no op",
			plan: messagingServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringUnknown(),
			},
			state: base,
		},
		{
			name: "service type cannot be patched",
			plan: messagingServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: types.StringValue("PubSub"),
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
			if tt.wantValue != nil && !reflect.DeepEqual(ops[0].Value, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}

func TestSemanticJSONPlanModifier(t *testing.T) {
	state := `{"bootstrapServers":"kafka:9092"}`
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("{\n  \"bootstrapServers\": \"kafka:9092\"\n}")}
	semanticJSONPlanModifier().PlanModifyString(context.Background(), planmodifier.StringRequest{
		PlanValue:  resp.PlanValue,
		StateValue: types.StringValue(state),
	}, resp)

	if resp.PlanValue.ValueString() != state {
		t.Fatalf("PlanValue = %q, want %q", resp.PlanValue.ValueString(), state)
	}
}

func TestParseConnectionJSONRejectsInvalidDocument(t *testing.T) {
	if _, err := parseConnectionJSON(`{"broken"`); err == nil {
		t.Fatal("parseConnectionJSON accepted invalid JSON")
	}

	decoded, err := parseConnectionJSON(`{"bootstrapServers":"kafka:9092"}`)
	if err != nil {
		t.Fatalf("parseConnectionJSON error = %v", err)
	}
	if _, err := json.Marshal(decoded); err != nil {
		t.Fatalf("re-encoding decoded connection failed: %v", err)
	}

	if connectionConfigured(types.StringNull()) || connectionConfigured(types.StringUnknown()) {
		t.Fatal("connectionConfigured accepted a null or unknown value")
	}
}
