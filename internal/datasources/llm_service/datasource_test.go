package llm_service

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

	if resp.TypeName != "openmetadata_llm_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_llm_service")
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
		"service_type",
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
	id := uuid.MustParse("2b7a1e42-9c55-4a7b-8e11-3d2c2c2c2c2c")
	fqn := "openai"
	displayName := "OpenAI service"
	description := "OpenAI service for prompt generation."

	var model llmServiceModel
	apply(&ometa.LLMService{
		Id:                 id,
		Name:               "openai",
		ServiceType:        ometa.LLMServiceServiceTypeOpenAI,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "openai" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.ServiceType.ValueString() != "OpenAI" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
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
