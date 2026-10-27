package directory

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

	if resp.TypeName != "openmetadata_directory" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_directory")
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
		"parent",
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
	serviceFQN := "gdrive"
	parentName := "root"
	parentFQN := "gdrive.root"
	fqn := "gdrive.root.child"
	displayName := "Child folder"
	description := "Child of the shared root."

	var model directoryModel
	apply(&ometa.Directory{
		Id:                 id,
		Name:               "child",
		Service:            ometa.EntityReference{FullyQualifiedName: &serviceFQN},
		Parent:             &ometa.EntityReference{Name: &parentName, FullyQualifiedName: &parentFQN},
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "child" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.Parent.ValueString() != parentFQN {
		t.Fatalf("Parent = %q", model.Parent.ValueString())
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

func TestApplyMissingReferenceKeepsPriorValue(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := directoryModel{
		Service: types.StringValue("gdrive"),
		Parent:  types.StringValue("gdrive.root"),
	}
	apply(&ometa.Directory{Id: id, Name: "root"}, &model)

	if model.Service.ValueString() != "gdrive" {
		t.Fatalf("Service = %q, want the prior value preserved", model.Service.ValueString())
	}
	if model.Parent.ValueString() != "gdrive.root" {
		t.Fatalf("Parent = %q, want the prior value preserved", model.Parent.ValueString())
	}
}
