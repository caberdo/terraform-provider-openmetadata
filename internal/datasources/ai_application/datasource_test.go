package ai_application

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

	if resp.TypeName != "openmetadata_ai_application" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ai_application")
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
		"application_type",
		"deployment_url",
		"development_stage",
		"documentation",
		"source_code",
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
	fqn := "support-copilot"
	displayName := "Support Copilot"
	deploymentURL := "https://copilot.example.com"
	developmentStage := ometa.AIApplicationDevelopmentStageProduction
	description := "Answers support questions."

	var model aiApplicationModel
	apply(&ometa.AIApplication{
		Id:                 id,
		Name:               "support-copilot",
		ApplicationType:    ometa.AIApplicationApplicationTypeCopilot,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		DeploymentUrl:      &deploymentURL,
		DevelopmentStage:   &developmentStage,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "support-copilot" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.ApplicationType.ValueString() != "Copilot" {
		t.Fatalf("ApplicationType = %q", model.ApplicationType.ValueString())
	}
	if model.DeploymentURL.ValueString() != deploymentURL {
		t.Fatalf("DeploymentURL = %q", model.DeploymentURL.ValueString())
	}
	if model.DevelopmentStage.ValueString() != "Production" {
		t.Fatalf("DevelopmentStage = %q", model.DevelopmentStage.ValueString())
	}
	if model.Documentation.ValueString() != "" {
		t.Fatalf("Documentation = %q, want the empty default", model.Documentation.ValueString())
	}
	if model.SourceCode.ValueString() != "" {
		t.Fatalf("SourceCode = %q, want the empty default", model.SourceCode.ValueString())
	}
}
