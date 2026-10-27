package notification_template

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

	if resp.TypeName != "openmetadata_notification_template" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_notification_template")
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
		"description",
		"template_subject",
		"template_body",
		"is_modified_from_default",
		"updated_at",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "asset-created"
	displayName := "Asset created"
	description := "Sent when an asset is created."
	modified := true

	var model notificationTemplateModel
	apply(&ometa.NotificationTemplate{
		Id:                    id,
		Name:                  "asset-created",
		FullyQualifiedName:    &fqn,
		DisplayName:           &displayName,
		Description:           &description,
		TemplateSubject:       "Asset {{ entity.name }} created",
		TemplateBody:          "{{ entity.name }} was created.",
		IsModifiedFromDefault: &modified,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "asset-created" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.TemplateSubject.ValueString() != "Asset {{ entity.name }} created" {
		t.Fatalf("TemplateSubject = %q", model.TemplateSubject.ValueString())
	}
	if model.TemplateBody.ValueString() != "{{ entity.name }} was created." {
		t.Fatalf("TemplateBody = %q", model.TemplateBody.ValueString())
	}
	if !model.IsModifiedFromDefault.ValueBool() {
		t.Fatal("IsModifiedFromDefault = false")
	}
}
