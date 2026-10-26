package stored_procedure

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

	if resp.TypeName != "openmetadata_stored_procedure" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_stored_procedure")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "database_schema", "code"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
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

	for _, attribute := range []string{"language", "stored_procedure_type"} {
		optional, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !optional.Optional {
			t.Fatalf("%s must be an optional string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("2a7b3c64-9e45-4a5c-8e1f-3a2b1c0d9e8f")
	fqn := "postgres.warehouse.public.customer_report"
	displayName := "Customer report"
	code := "BEGIN SELECT 1; END"
	language := ometa.StoredProcedureCodeLanguageSQL
	procedureType := ometa.StoredProcedureStoredProcedureTypeStoredProcedure
	parentFQN := "postgres.warehouse.public"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	description := "Reports on customers."

	var model storedProcedureModel
	apply(&ometa.StoredProcedure{
		Id:                 id,
		Name:               "customer_report",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		DatabaseSchema:     &ometa.EntityReference{FullyQualifiedName: &parentFQN},
		StoredProcedureCode: ometa.StoredProcedureCode{
			Code:     &code,
			Language: &language,
		},
		StoredProcedureType: &procedureType,
		Version:             &version,
		UpdatedAt:           &updatedAt,
		UpdatedBy:           &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "customer_report" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DatabaseSchema.ValueString() != parentFQN {
		t.Fatalf("DatabaseSchema = %q", model.DatabaseSchema.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Reports on customers." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Code.ValueString() != code {
		t.Fatalf("Code = %q", model.Code.ValueString())
	}
	if model.Language.ValueString() != "SQL" {
		t.Fatalf("Language = %q", model.Language.ValueString())
	}
	if model.StoredProcedureType.ValueString() != "StoredProcedure" {
		t.Fatalf("StoredProcedureType = %q", model.StoredProcedureType.ValueString())
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

func TestApplyFallsBackToTopLevelCode(t *testing.T) {
	id := uuid.MustParse("2a7b3c64-9e45-4a5c-8e1f-3a2b1c0d9e8f")
	code := "BEGIN SELECT 2; END"

	var model storedProcedureModel
	apply(&ometa.StoredProcedure{Id: id, Name: "legacy", Code: &code}, &model)

	if model.Code.ValueString() != code {
		t.Fatalf("Code = %q, want the top-level code fallback", model.Code.ValueString())
	}
}

func TestApplyKeepsSelectorsWhenServerOmitsThem(t *testing.T) {
	id := uuid.MustParse("2a7b3c64-9e45-4a5c-8e1f-3a2b1c0d9e8f")
	parent := types.StringValue("postgres.warehouse.public")
	language := types.StringValue("Python")
	procedureType := types.StringValue("Function")
	emptyDescription := ""

	model := storedProcedureModel{
		DatabaseSchema:      parent,
		Language:            language,
		StoredProcedureType: procedureType,
	}
	apply(&ometa.StoredProcedure{Id: id, Name: "bare", Description: &emptyDescription}, &model)

	if !model.DatabaseSchema.Equal(parent) {
		t.Fatalf("DatabaseSchema = %v, want prior value", model.DatabaseSchema)
	}
	if !model.Language.Equal(language) {
		t.Fatalf("Language = %v, want prior value", model.Language)
	}
	if !model.StoredProcedureType.Equal(procedureType) {
		t.Fatalf("StoredProcedureType = %v, want prior value", model.StoredProcedureType)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
}

func TestPatchOps(t *testing.T) {
	base := storedProcedureModel{
		DisplayName:         types.StringValue("Display"),
		Description:         types.StringValue("before"),
		Code:                types.StringValue("BEGIN SELECT 1; END"),
		Language:            types.StringValue("SQL"),
		StoredProcedureType: types.StringValue("StoredProcedure"),
	}

	tests := []struct {
		name      string
		plan      storedProcedureModel
		state     storedProcedureModel
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
			plan: storedProcedureModel{
				DisplayName:         types.StringValue("New display"),
				Description:         base.Description,
				Code:                base.Code,
				Language:            base.Language,
				StoredProcedureType: base.StoredProcedureType,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "description changed",
			plan: storedProcedureModel{
				DisplayName:         base.DisplayName,
				Description:         types.StringValue("after"),
				Code:                base.Code,
				Language:            base.Language,
				StoredProcedureType: base.StoredProcedureType,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "code changed",
			plan: storedProcedureModel{
				DisplayName:         base.DisplayName,
				Description:         base.Description,
				Code:                types.StringValue("BEGIN SELECT 2; END"),
				Language:            base.Language,
				StoredProcedureType: base.StoredProcedureType,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/storedProcedureCode",
			wantOp:   "add",
		},
		{
			name: "language changed",
			plan: storedProcedureModel{
				DisplayName:         base.DisplayName,
				Description:         base.Description,
				Code:                base.Code,
				Language:            types.StringValue("Python"),
				StoredProcedureType: base.StoredProcedureType,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/storedProcedureCode",
			wantOp:   "add",
		},
		{
			name: "stored procedure type cannot be patched",
			plan: storedProcedureModel{
				DisplayName:         base.DisplayName,
				Description:         base.Description,
				Code:                base.Code,
				Language:            base.Language,
				StoredProcedureType: types.StringValue("Function"),
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

func TestPatchOpsSendsNestedCode(t *testing.T) {
	state := storedProcedureModel{
		Description: types.StringValue("before"),
		Code:        types.StringValue("BEGIN SELECT 1; END"),
		Language:    types.StringValue("SQL"),
	}
	plan := storedProcedureModel{
		Description: state.Description,
		Code:        types.StringValue("BEGIN SELECT 2; END"),
		Language:    types.StringValue("Python"),
	}

	ops := patchOps(plan, state)
	if len(ops) != 1 {
		t.Fatalf("len(ops) = %d, want 1: %#v", len(ops), ops)
	}

	code, ok := ops[0].Value.(ometa.StoredProcedureCode)
	if !ok {
		t.Fatalf("value is %T, want ometa.StoredProcedureCode", ops[0].Value)
	}
	if code.Code == nil || *code.Code != "BEGIN SELECT 2; END" {
		t.Fatalf("code = %#v", code.Code)
	}
	if code.Language == nil || *code.Language != ometa.StoredProcedureCodeLanguagePython {
		t.Fatalf("language = %#v", code.Language)
	}
}
