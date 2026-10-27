package test_suite

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

	if resp.TypeName != "openmetadata_test_suite" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_test_suite")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "name", "description", "basic", "executable", "updated_at"} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "my_suite"
	displayName := "My suite"
	description := "Groups data quality checks."
	basic := true
	executable := true

	var model testSuiteModel
	apply(&ometa.TestSuite{
		Id:                 &id,
		Name:               "my_suite",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Basic:              &basic,
		Executable:         &executable,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "my_suite" {
		t.Fatalf("Name = %q", model.Name.ValueString())
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
	if !model.Executable.ValueBool() {
		t.Fatal("Executable = false")
	}
}
