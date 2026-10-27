package api_service

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

	if resp.TypeName != "openmetadata_api_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_api_service")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want StringAttribute", resp.Schema.Attributes["name"])
	}
	if !name.Required {
		t.Fatal("name must be required")
	}

	serviceType, ok := resp.Schema.Attributes["service_type"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("service_type attribute is %T, want StringAttribute", resp.Schema.Attributes["service_type"])
	}
	if !serviceType.Required {
		t.Fatal("service_type must be required")
	}
	if len(serviceType.Validators) == 0 {
		t.Fatal("service_type must validate the connector type")
	}

	connection, ok := resp.Schema.Attributes["connection_config"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("connection attribute is %T, want StringAttribute", resp.Schema.Attributes["connection_config"])
	}
	if !connection.Optional || !connection.Computed {
		t.Fatal("connection must be optional and computed")
	}
	if !connection.Sensitive {
		t.Fatal("connection must be sensitive")
	}
	if len(connection.PlanModifiers) != 2 {
		t.Fatalf("connection plan modifiers = %d, want 2", len(connection.PlanModifiers))
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed || description.Default == nil {
		t.Fatal("description must be optional and computed with an empty default")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Required || displayName.Computed {
		t.Fatal("display_name must be optional only")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "payments-api"
	displayName := "Payments API"
	description := "REST API for payments."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model apiServiceModel
	apply(&ometa.ApiService{
		Id:                 id,
		Name:               "payments-api",
		ServiceType:        ometa.ApiServiceServiceTypeRest,
		Connection:         &ometa.ApiConnection{Config: map[string]any{"type": "Rest"}},
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "payments-api" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.ServiceType.ValueString() != "Rest" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if model.Connection.ValueString() != `{"type":"Rest"}` {
		t.Fatalf("Connection = %q", model.Connection.ValueString())
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

func TestApplyKeepsPriorValuesWhenServerOmitsThem(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := apiServiceModel{
		ServiceType: types.StringValue("Rest"),
		Connection:  types.StringValue(`{"type":"Rest"}`),
	}
	apply(&ometa.ApiService{Id: id, Name: "bare"}, &model)

	if model.ServiceType.ValueString() != "Rest" {
		t.Fatalf("ServiceType = %q, want the prior value preserved", model.ServiceType.ValueString())
	}
	if model.Connection.ValueString() != `{"type":"Rest"}` {
		t.Fatalf("Connection = %q, want the prior value preserved", model.Connection.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := apiServiceModel{
		Name:        types.StringValue("payments-api"),
		ServiceType: types.StringValue("Rest"),
		Connection:  types.StringNull(),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      apiServiceModel
		state     apiServiceModel
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
			name: "display name replaced",
			plan: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
				DisplayName: types.StringNull(),
				Description: base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "connection added",
			plan: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"type":"Rest"}`),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "add",
		},
		{
			name: "connection removed",
			plan: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  types.StringNull(),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"type":"Rest"}`),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "remove",
		},
		{
			name: "semantically equal connection emits no op",
			plan: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{ "type" : "Rest" }`),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state: apiServiceModel{
				Name:        base.Name,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"type":"Rest"}`),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			wantOps: 0,
		},
		{
			name: "name and service type cannot be patched",
			plan: apiServiceModel{
				Name:        types.StringValue("renamed-api"),
				ServiceType: types.StringValue("WEBHOOK"),
				Connection:  base.Connection,
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state:   base,
			wantOps: 0,
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
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}

func TestSemanticJSONPlanModifier(t *testing.T) {
	modifier := semanticJSONPlanModifier()

	var resp planmodifier.StringResponse
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		PlanValue:  types.StringValue(`{ "type" : "Rest" }`),
		StateValue: types.StringValue(`{"type":"Rest"}`),
	}, &resp)

	if !resp.PlanValue.Equal(types.StringValue(`{"type":"Rest"}`)) {
		t.Fatalf("PlanValue = %v, want the state value", resp.PlanValue)
	}
}

func TestParseConnectionJSONRejectsInvalidDocument(t *testing.T) {
	if _, err := parseConnectionJSON(`{"broken"`); err == nil {
		t.Fatal("parseConnectionJSON accepted invalid JSON")
	}
	if connectionConfigured(types.StringNull()) || connectionConfigured(types.StringUnknown()) {
		t.Fatal("connectionConfigured accepted a null or unknown value")
	}
	if connectionConfigured(types.StringValue("  ")) {
		t.Fatal("connectionConfigured accepted an empty value")
	}
}
