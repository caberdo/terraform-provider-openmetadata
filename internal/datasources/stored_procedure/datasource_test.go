package stored_procedure

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

	if resp.TypeName != "openmetadata_stored_procedure" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_stored_procedure")
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
		"database_schema",
		"display_name",
		"description",
		"code",
		"language",
		"stored_procedure_type",
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
	id := uuid.MustParse("2a7b3c64-9e45-4a5c-8e1f-3a2b1c0d9e8f")
	fqn := "postgres.warehouse.public.customer_report"
	displayName := "Customer report"
	description := "Reports on customers."
	code := "BEGIN SELECT 1; END"
	language := ometa.StoredProcedureCodeLanguageSQL
	procedureType := ometa.StoredProcedureStoredProcedureTypeFunction
	parentFQN := "postgres.warehouse.public"

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
	if model.StoredProcedureType.ValueString() != "Function" {
		t.Fatalf("StoredProcedureType = %q", model.StoredProcedureType.ValueString())
	}
}

func TestApplyUsesTopLevelCodeAndParentName(t *testing.T) {
	id := uuid.MustParse("2a7b3c64-9e45-4a5c-8e1f-3a2b1c0d9e8f")
	code := "BEGIN SELECT 2; END"
	parentName := "public"

	var model storedProcedureModel
	apply(&ometa.StoredProcedure{
		Id:             id,
		Name:           "legacy",
		Code:           &code,
		DatabaseSchema: &ometa.EntityReference{Name: &parentName},
	}, &model)

	if model.Code.ValueString() != code {
		t.Fatalf("Code = %q, want the top-level code fallback", model.Code.ValueString())
	}
	if model.DatabaseSchema.ValueString() != parentName {
		t.Fatalf("DatabaseSchema = %q, want the parent name fallback", model.DatabaseSchema.ValueString())
	}
	if !model.Language.IsNull() {
		t.Fatalf("Language = %v, want null", model.Language)
	}
}
