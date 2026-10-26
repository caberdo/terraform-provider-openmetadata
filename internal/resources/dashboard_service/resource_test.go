package dashboard_service

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

	if resp.TypeName != "openmetadata_dashboard_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_dashboard_service")
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
	if len(connection.PlanModifiers) != 2 {
		t.Fatalf("connection plan modifiers = %d, want 2", len(connection.PlanModifiers))
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"version",
		"updated_at",
		"updated_by",
	} {
		if attribute == "version" {
			floatAttribute, floatOK := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !floatOK || !floatAttribute.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "custom_dashboard"
	displayName := "Custom Dashboard"
	description := "Dashboards built in-house."
	version := 0.3
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	config := map[string]any{
		"type":     "CustomDashboard",
		"hostPort": "https://dashboards.example.com",
	}

	var model dashboardServiceModel
	apply(&ometa.DashboardService{
		Id:                 id,
		Name:               "custom_dashboard",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		ServiceType:        ometa.DashboardServiceServiceType("CustomDashboard"),
		Connection:         &ometa.DashboardConnection{Config: config},
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "custom_dashboard" {
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
	if model.ServiceType.ValueString() != "CustomDashboard" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if !jsonEqual(model.Connection.ValueString(), `{"type":"CustomDashboard","hostPort":"https://dashboards.example.com"}`) {
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

	model := dashboardServiceModel{Connection: types.StringUnknown()}
	apply(&ometa.DashboardService{
		Id:          id,
		Name:        "bare",
		ServiceType: ometa.DashboardServiceServiceType("CustomDashboard"),
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
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestApplyPreservesEquivalentConnection(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	formatted := "{\n  \"type\": \"CustomDashboard\",\n  \"hostPort\": \"https://dashboards.example.com\"\n}"

	model := dashboardServiceModel{Connection: types.StringValue(formatted)}
	apply(&ometa.DashboardService{
		Id:          id,
		Name:        "custom_dashboard",
		ServiceType: ometa.DashboardServiceServiceType("CustomDashboard"),
		Connection: &ometa.DashboardConnection{Config: map[string]any{
			"type":     "CustomDashboard",
			"hostPort": "https://dashboards.example.com",
		}},
	}, &model)

	if model.Connection.ValueString() != formatted {
		t.Fatalf("Connection = %q, want the formatted state value", model.Connection.ValueString())
	}
}

func TestApplyKeepsConnectionWhenServerOmitsIt(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	existing := `{"type":"CustomDashboard"}`

	model := dashboardServiceModel{Connection: types.StringValue(existing)}
	apply(&ometa.DashboardService{
		Id:          id,
		Name:        "custom_dashboard",
		ServiceType: ometa.DashboardServiceServiceType("CustomDashboard"),
	}, &model)

	if model.Connection.ValueString() != existing {
		t.Fatalf("Connection = %q, want %q", model.Connection.ValueString(), existing)
	}

	model = dashboardServiceModel{Connection: types.StringNull()}
	apply(&ometa.DashboardService{
		Id:          id,
		Name:        "custom_dashboard",
		ServiceType: ometa.DashboardServiceServiceType("CustomDashboard"),
	}, &model)

	if model.Connection.IsUnknown() {
		t.Fatal("Connection stayed unknown")
	}
}

func TestPatchOps(t *testing.T) {
	base := dashboardServiceModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		ServiceType: types.StringValue("CustomDashboard"),
		Connection:  types.StringValue(`{"type":"CustomDashboard","hostPort":"https://dashboards.example.com"}`),
	}

	tests := []struct {
		name      string
		plan      dashboardServiceModel
		state     dashboardServiceModel
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
			plan: dashboardServiceModel{
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
			name: "display name removed",
			plan: dashboardServiceModel{
				DisplayName: types.StringNull(),
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  base.Connection,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: dashboardServiceModel{
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
			name: "service type cannot be patched",
			plan: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: types.StringValue("Grafana"),
				Connection:  base.Connection,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "connection added",
			plan: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"type":"Grafana","hostPort":"http://grafana:3000"}`),
			},
			state: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringNull(),
			},
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "add",
			wantValue: &ometa.DashboardConnection{Config: map[string]any{
				"type":     "Grafana",
				"hostPort": "http://grafana:3000",
			}},
		},
		{
			name: "connection removed",
			plan: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringNull(),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "remove",
		},
		{
			name: "connection reformatted but equal",
			plan: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue("{\n  \"hostPort\": \"https://dashboards.example.com\",\n  \"type\": \"CustomDashboard\"\n}"),
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "connection changed",
			plan: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringValue(`{"type":"CustomDashboard","hostPort":"https://new.example.com"}`),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/connection",
			wantOp:   "add",
			wantValue: &ometa.DashboardConnection{Config: map[string]any{
				"type":     "CustomDashboard",
				"hostPort": "https://new.example.com",
			}},
		},
		{
			name: "unknown connection emits no op",
			plan: dashboardServiceModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ServiceType: base.ServiceType,
				Connection:  types.StringUnknown(),
			},
			state:   base,
			wantOps: 0,
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

func TestJSONEqual(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  bool
	}{
		{
			name:  "key order ignored",
			left:  `{"a":1,"b":{"c":2}}`,
			right: `{"b":{"c":2},"a":1}`,
			want:  true,
		},
		{
			name:  "whitespace ignored",
			left:  "{\n  \"a\": 1\n}",
			right: `{"a":1}`,
			want:  true,
		},
		{
			name:  "different values",
			left:  `{"a":1}`,
			right: `{"a":2}`,
			want:  false,
		},
		{
			name:  "invalid left",
			left:  "not json",
			right: `{"a":1}`,
			want:  false,
		},
		{
			name:  "invalid right",
			left:  `{"a":1}`,
			right: "",
			want:  false,
		},
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
	state := `{"type":"CustomDashboard","hostPort":"https://dashboards.example.com"}`

	tests := []struct {
		name     string
		plan     types.String
		state    types.String
		wantPlan types.String
	}{
		{
			name:     "equivalent formatted json adopts state",
			plan:     types.StringValue("{\n  \"hostPort\": \"https://dashboards.example.com\",\n  \"type\": \"CustomDashboard\"\n}"),
			state:    types.StringValue(state),
			wantPlan: types.StringValue(state),
		},
		{
			name:     "different json keeps plan",
			plan:     types.StringValue(`{"type":"Grafana"}`),
			state:    types.StringValue(state),
			wantPlan: types.StringValue(`{"type":"Grafana"}`),
		},
		{
			name:     "null plan untouched",
			plan:     types.StringNull(),
			state:    types.StringValue(state),
			wantPlan: types.StringNull(),
		},
		{
			name:     "null state untouched",
			plan:     types.StringValue(state),
			state:    types.StringNull(),
			wantPlan: types.StringValue(state),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &planmodifier.StringResponse{PlanValue: tt.plan}
			semanticJSONPlanModifier().PlanModifyString(context.Background(), planmodifier.StringRequest{
				PlanValue:  tt.plan,
				StateValue: tt.state,
			}, resp)

			if !resp.PlanValue.Equal(tt.wantPlan) {
				t.Fatalf("PlanValue = %v, want %v", resp.PlanValue, tt.wantPlan)
			}
		})
	}
}

func TestParseConnectionJSONRejectsInvalidDocument(t *testing.T) {
	if _, err := parseConnectionJSON(`{"broken"`); err == nil {
		t.Fatal("parseConnectionJSON accepted invalid JSON")
	}

	decoded, err := parseConnectionJSON(`{"type":"CustomDashboard"}`)
	if err != nil {
		t.Fatalf("parseConnectionJSON error = %v", err)
	}
	if _, err := json.Marshal(decoded); err != nil {
		t.Fatalf("re-encoding decoded connection failed: %v", err)
	}

	if connectionConfigured(types.StringNull()) || connectionConfigured(types.StringUnknown()) {
		t.Fatal("connectionConfigured accepted a null or unknown value")
	}
	if connectionConfigured(types.StringValue("  ")) {
		t.Fatal("connectionConfigured accepted an empty value")
	}
}
