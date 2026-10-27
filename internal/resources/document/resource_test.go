package document

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_document" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_document")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "entity_type", "fully_qualified_name", "data"} {
		attr, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !attr.Required {
			t.Fatalf("%s must be required", attribute)
		}
		if attribute == "data" {
			if len(attr.PlanModifiers) == 0 {
				t.Fatal("data must carry a plan modifier")
			}
			if len(attr.Validators) == 0 {
				t.Fatal("data must carry a validator")
			}
			continue
		}
		if len(attr.PlanModifiers) == 0 {
			t.Fatalf("%s must carry RequiresReplace", attribute)
		}
	}

	for _, attribute := range []string{
		"id",
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
	if !ok || !description.Optional || !description.Computed {
		t.Fatal("description must be optional and computed")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Required || displayName.Computed {
		t.Fatal("display_name must be optional only")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "getting-started.acceptance"
	displayName := "Getting started"
	description := "Onboarding notes."
	version := 0.3
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model documentModel
	apply(&ometa.Document{
		Id:                 &id,
		Name:               "getting-started",
		EntityType:         "KnowledgePanel",
		FullyQualifiedName: fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Data:               ometa.Data{"title": "Acceptance", "body": "Initial"},
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "getting-started" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.EntityType.ValueString() != "KnowledgePanel" {
		t.Fatalf("EntityType = %q", model.EntityType.ValueString())
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
	if model.Data.ValueString() != `{"body":"Initial","title":"Acceptance"}` {
		t.Fatalf("Data = %q, want canonical JSON", model.Data.ValueString())
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

func TestApplyKeepsPriorDataWhenPayloadMissing(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := documentModel{
		Data: types.StringValue(`{"body":"Initial","title":"Acceptance"}`),
	}
	apply(&ometa.Document{
		Id:                 &id,
		Name:               "bare",
		EntityType:         "KnowledgePanel",
		FullyQualifiedName: "bare",
	}, &model)

	if model.Data.ValueString() != `{"body":"Initial","title":"Acceptance"}` {
		t.Fatalf("Data = %q, want the prior value preserved", model.Data.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestCanonicalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "key order is sorted",
			input: `{"title":"Acceptance","body":"Initial"}`,
			want:  `{"body":"Initial","title":"Acceptance"}`,
		},
		{
			name:  "whitespace is removed",
			input: "{\n  \"title\": \"Acceptance\",\n  \"body\": \"Initial\"\n}",
			want:  `{"body":"Initial","title":"Acceptance"}`,
		},
		{
			name:  "html characters are not escaped",
			input: `{"body":"<p>Hello & welcome</p>"}`,
			want:  `{"body":"<p>Hello & welcome</p>"}`,
		},
		{
			name:  "nested objects are sorted",
			input: `{"title":"T","meta":{"z":1,"a":2}}`,
			want:  `{"meta":{"a":2,"z":1},"title":"T"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := canonicalData(tt.input)
			if err != nil {
				t.Fatalf("canonicalData(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("canonicalData(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}

	for _, invalid := range []string{`{"broken"`, `[]`, `"scalar"`, `null`} {
		if _, err := canonicalData(invalid); err == nil {
			t.Fatalf("canonicalData(%q) accepted invalid JSON", invalid)
		}
	}
}

func TestCanonicalDataPlanModifier(t *testing.T) {
	modifier := canonicalDataPlanModifier()

	canonical := `{"body":"Initial","title":"Acceptance"}`

	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("{\n  \"title\": \"Acceptance\",\n  \"body\": \"Initial\"\n}")}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{PlanValue: resp.PlanValue}, resp)
	if resp.PlanValue.ValueString() != canonical {
		t.Fatalf("PlanValue = %q, want %q", resp.PlanValue.ValueString(), canonical)
	}

	raw := types.StringValue(canonical)
	resp = &planmodifier.StringResponse{PlanValue: raw}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{PlanValue: raw}, resp)
	if resp.PlanValue.ValueString() != canonical {
		t.Fatalf("PlanValue = %q, want the canonical value unchanged", resp.PlanValue.ValueString())
	}

	nullPlan := types.StringNull()
	resp = &planmodifier.StringResponse{PlanValue: nullPlan}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{PlanValue: nullPlan}, resp)
	if !resp.PlanValue.IsNull() {
		t.Fatalf("PlanValue = %v, want null", resp.PlanValue)
	}

	unknownPlan := types.StringUnknown()
	resp = &planmodifier.StringResponse{PlanValue: unknownPlan}
	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{PlanValue: unknownPlan}, resp)
	if !resp.PlanValue.IsUnknown() {
		t.Fatalf("PlanValue = %v, want unknown", resp.PlanValue)
	}
}

func TestDataJSONValidatorRejectsInvalidJSON(t *testing.T) {
	dataValidator := dataJSONValidator()

	resp := &validator.StringResponse{}
	dataValidator.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("data"),
		ConfigValue: types.StringValue(`{"broken"`),
	}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("validator accepted invalid JSON")
	}

	resp = &validator.StringResponse{}
	dataValidator.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("data"),
		ConfigValue: types.StringValue(`{"title":"Acceptance"}`),
	}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("validator rejected valid JSON: %v", resp.Diagnostics.Errors())
	}
}

func TestPatchOps(t *testing.T) {
	base := documentModel{
		Name:        types.StringValue("getting-started"),
		EntityType:  types.StringValue("KnowledgePanel"),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Data:        types.StringValue(`{"body":"Initial","title":"Acceptance"}`),
	}

	updatedData := `{"body":"Updated","title":"Acceptance"}`

	tests := []struct {
		name      string
		plan      documentModel
		state     documentModel
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
			plan: documentModel{
				Name:        base.Name,
				EntityType:  base.EntityType,
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Data:        base.Data,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: documentModel{
				Name:        base.Name,
				EntityType:  base.EntityType,
				DisplayName: types.StringNull(),
				Description: base.Description,
				Data:        base.Data,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: documentModel{
				Name:        base.Name,
				EntityType:  base.EntityType,
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Data:        base.Data,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "data changed",
			plan: documentModel{
				Name:        base.Name,
				EntityType:  base.EntityType,
				DisplayName: base.DisplayName,
				Description: base.Description,
				Data:        types.StringValue(updatedData),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/data",
			wantOp:    "add",
			wantValue: map[string]any{"body": "Updated", "title": "Acceptance"},
		},
		{
			name: "name entity type and fully qualified name cannot be patched",
			plan: documentModel{
				Name:        types.StringValue("renamed"),
				EntityType:  types.StringValue("OtherPanel"),
				DisplayName: base.DisplayName,
				Description: base.Description,
				Data:        base.Data,
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
			if tt.wantValue != nil && !reflect.DeepEqual(ops[0].Value, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
