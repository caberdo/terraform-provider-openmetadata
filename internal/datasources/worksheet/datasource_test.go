package worksheet

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_worksheet" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_worksheet")
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
		"spreadsheet",
		"service",
		"display_name",
		"description",
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
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	spreadsheetFQN := "google-drive.budget"
	serviceFQN := "google-drive"
	fqn := "google-drive.budget.q1"
	displayName := "Q1 budget"
	description := "First-quarter budget."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model worksheetModel
	apply(&ometa.Worksheet{
		Id:                 id,
		Name:               "q1",
		Spreadsheet:        ometa.EntityReference{FullyQualifiedName: &spreadsheetFQN},
		Service:            ometa.EntityReference{FullyQualifiedName: &serviceFQN},
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
	if model.Name.ValueString() != "q1" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Spreadsheet.ValueString() != spreadsheetFQN {
		t.Fatalf("Spreadsheet = %q", model.Spreadsheet.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
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
}

func TestApplyKeepsPriorRelationshipsWhenReferenceMissing(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := worksheetModel{
		Spreadsheet: types.StringValue("google-drive.budget"),
		Service:     types.StringValue("google-drive"),
	}
	apply(&ometa.Worksheet{
		Id:   id,
		Name: "bare",
	}, &model)

	if model.Spreadsheet.ValueString() != "google-drive.budget" {
		t.Fatalf("Spreadsheet = %q, want the prior value preserved", model.Spreadsheet.ValueString())
	}
	if model.Service.ValueString() != "google-drive" {
		t.Fatalf("Service = %q, want the prior value preserved", model.Service.ValueString())
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}
