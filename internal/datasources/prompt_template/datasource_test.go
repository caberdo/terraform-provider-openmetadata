package prompt_template

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

	if resp.TypeName != "openmetadata_prompt_template" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_prompt_template")
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
		"template_content",
		"template_type",
		"system_prompt",
		"updated_at",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "ticket-classifier"
	displayName := "Ticket classifier"
	description := "Classifies support tickets."
	templateType := ometa.PromptTemplateTemplateTypeClassification
	templateVersion := "1.2"
	systemPrompt := "You are a support agent."

	var model promptTemplateModel
	apply(&ometa.PromptTemplate{
		Id:                 id,
		Name:               "ticket-classifier",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		TemplateContent:    "Classify the ticket: {{ticket}}",
		TemplateType:       &templateType,
		TemplateVersion:    &templateVersion,
		SystemPrompt:       &systemPrompt,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "ticket-classifier" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.TemplateContent.ValueString() != "Classify the ticket: {{ticket}}" {
		t.Fatalf("TemplateContent = %q", model.TemplateContent.ValueString())
	}
	if model.TemplateType.ValueString() != string(templateType) {
		t.Fatalf("TemplateType = %q", model.TemplateType.ValueString())
	}
	if model.SystemPrompt.ValueString() != systemPrompt {
		t.Fatalf("SystemPrompt = %q", model.SystemPrompt.ValueString())
	}
}
