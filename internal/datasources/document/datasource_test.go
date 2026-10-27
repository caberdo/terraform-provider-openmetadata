package document

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

	if resp.TypeName != "openmetadata_document" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_document")
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
		"entity_type",
		"display_name",
		"description",
		"data",
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
		t.Fatalf("Data = %q, want compact sorted JSON", model.Data.ValueString())
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
}
