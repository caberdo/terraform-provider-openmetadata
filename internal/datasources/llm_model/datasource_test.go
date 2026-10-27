package llm_model

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

	if resp.TypeName != "openmetadata_llm_model" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_llm_model")
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
		"base_model",
		"model_provider",
		"model_version",
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
	fqn := "llm-service.my-model"
	displayName := "Support model"
	modelProvider := "openai"
	modelVersion := "2024-08-06"
	serviceFQN := "llm-service"
	description := "Answers support questions."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model llmModelModel
	apply(&ometa.LLMModel{
		Id:                 id,
		Name:               "my-model",
		BaseModel:          "gpt-4o",
		ModelProvider:      &modelProvider,
		ModelVersion:       &modelVersion,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service:            &ometa.EntityReference{FullyQualifiedName: &serviceFQN},
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "my-model" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.BaseModel.ValueString() != "gpt-4o" {
		t.Fatalf("BaseModel = %q", model.BaseModel.ValueString())
	}
	if model.ModelProvider.ValueString() != modelProvider {
		t.Fatalf("ModelProvider = %q", model.ModelProvider.ValueString())
	}
	if model.ModelVersion.ValueString() != modelVersion {
		t.Fatalf("ModelVersion = %q", model.ModelVersion.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Answers support questions." {
		t.Fatalf("Description = %q", model.Description.ValueString())
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
