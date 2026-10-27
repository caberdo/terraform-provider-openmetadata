package test_suite

import (
	"context"
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

	if resp.TypeName != "openmetadata_test_suite" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_test_suite")
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

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional {
		t.Fatal("display_name must be an optional string")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed {
		t.Fatal("description must be an optional computed string")
	}

	for _, attribute := range []string{"basic", "executable"} {
		boolean, ok := resp.Schema.Attributes[attribute].(rschema.BoolAttribute)
		if !ok || !boolean.Computed {
			t.Fatalf("%s must be a computed bool", attribute)
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	if version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute); !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "my_suite"
	displayName := "My suite"
	description := "Groups data quality checks."
	basic := true
	executable := false
	version := 0.3
	updatedAt := int64(1700000000000)
	updatedBy := "admin"

	var model testSuiteModel
	apply(&ometa.TestSuite{
		Id:                 &id,
		Name:               "my_suite",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Basic:              &basic,
		Executable:         &executable,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "my_suite" {
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
	if !model.Basic.ValueBool() {
		t.Fatal("Basic = false")
	}
	if model.Executable.ValueBool() {
		t.Fatal("Executable = true")
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

	var model testSuiteModel
	apply(&ometa.TestSuite{Id: &id, Name: "bare"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Basic.ValueBool() || model.Executable.ValueBool() {
		t.Fatal("Basic and Executable must fall back to false when the API omits them")
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := testSuiteModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      testSuiteModel
		state     testSuiteModel
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
			plan: testSuiteModel{
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
			plan: testSuiteModel{
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
			plan: testSuiteModel{
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
			name: "display name and description changed",
			plan: testSuiteModel{
				DisplayName: types.StringValue("New display"),
				Description: types.StringValue("after"),
			},
			state:     base,
			wantOps:   2,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
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
