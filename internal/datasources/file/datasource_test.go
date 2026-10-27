package file

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

	if resp.TypeName != "openmetadata_file" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_file")
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
		"service",
		"directory",
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
	serviceFQN := "google-drive"
	directoryName := "reports"
	directoryFQN := "google-drive.reports"
	fqn := "google-drive.reports.quarterly"
	displayName := "Quarterly report"
	description := "Quarterly report PDF."

	var model fileModel
	apply(&ometa.File{
		Id:                 id,
		Name:               "quarterly",
		Service:            ometa.EntityReference{FullyQualifiedName: &serviceFQN},
		Directory:          &ometa.EntityReference{Name: &directoryName, FullyQualifiedName: &directoryFQN},
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "quarterly" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.Directory.ValueString() != directoryFQN {
		t.Fatalf("Directory = %q", model.Directory.ValueString())
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
}

func TestApplyKeepsPriorDirectoryWhenReferenceMissing(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "google-drive"

	model := fileModel{
		Directory: types.StringValue("google-drive.reports"),
	}
	apply(&ometa.File{
		Id:        id,
		Name:      "bare",
		Service:   ometa.EntityReference{Name: &serviceName},
		Directory: &ometa.EntityReference{Id: uuid.New()},
	}, &model)

	if model.Service.ValueString() != "google-drive" {
		t.Fatalf("Service = %q, want the name fallback", model.Service.ValueString())
	}
	if model.Directory.ValueString() != "google-drive.reports" {
		t.Fatalf("Directory = %q, want the prior value preserved", model.Directory.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty fallback", model.Description.ValueString())
	}
}

func TestApplyWithoutDirectoryLeavesPriorValue(t *testing.T) {
	serviceName := "google-drive"

	model := fileModel{
		Directory: types.StringValue("google-drive.reports"),
	}
	apply(&ometa.File{
		Id:      uuid.New(),
		Name:    "bare",
		Service: ometa.EntityReference{Name: &serviceName},
	}, &model)

	if model.Directory.ValueString() != "google-drive.reports" {
		t.Fatalf("Directory = %q, want the prior value preserved", model.Directory.ValueString())
	}
}
