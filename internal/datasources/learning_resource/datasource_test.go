package learning_resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_learning_resource" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_learning_resource")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be an optional computed selector", attribute)
		}
		if len(selector.Validators) == 0 {
			t.Fatalf("%s must validate the exactly-one-of rule", attribute)
		}
	}

	categories, ok := resp.Schema.Attributes["categories"].(dschema.ListAttribute)
	if !ok || !categories.Computed {
		t.Fatal("categories must be a computed list")
	}

	contexts, ok := resp.Schema.Attributes["contexts"].(dschema.ListNestedAttribute)
	if !ok || !contexts.Computed {
		t.Fatal("contexts must be a computed nested list")
	}
	if _, ok := contexts.NestedObject.Attributes["page_id"]; !ok {
		t.Fatal("contexts.page_id is missing")
	}

	source, ok := resp.Schema.Attributes["source"].(dschema.ListNestedAttribute)
	if !ok || !source.Computed {
		t.Fatal("source must be a computed nested list")
	}
	if _, ok := source.NestedObject.Attributes["url"]; !ok {
		t.Fatal("source.url is missing")
	}

	if _, ok := resp.Schema.Attributes["completion_threshold"]; ok {
		t.Fatal("completion_threshold must not be exposed, because the API never returns it")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "data_governance_101"
	displayName := "Data Governance 101"
	provider := "user"
	difficulty := ometa.LearningResourceDifficultyIntro
	status := ometa.LearningResourceStatusActive
	priority := int32(1)
	estimatedDuration := int32(15)

	var model learningResourceModel
	diags := apply(context.Background(), &ometa.LearningResource{
		Id:                 id,
		Name:               "data_governance_101",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        ptr("Introduction to governance."),
		ResourceType:       ometa.Article,
		Categories:         []ometa.LearningResourceCategories{ometa.AI, ometa.DataGovernance},
		Contexts: []ometa.LearningResourceContext{{
			PageId:   "home",
			Priority: &priority,
		}},
		Source: ometa.LearningResourceSource{
			Url:      "https://example.com/academy",
			Provider: &provider,
		},
		Difficulty:        &difficulty,
		Status:            &status,
		EstimatedDuration: &estimatedDuration,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.ResourceType.ValueString() != "Article" {
		t.Fatalf("ResourceType = %q", model.ResourceType.ValueString())
	}
	if model.Difficulty.ValueString() != "Intro" {
		t.Fatalf("Difficulty = %q", model.Difficulty.ValueString())
	}
	if model.Status.ValueString() != "Active" {
		t.Fatalf("Status = %q", model.Status.ValueString())
	}
	if model.EstimatedDuration.ValueInt64() != 15 {
		t.Fatalf("EstimatedDuration = %d", model.EstimatedDuration.ValueInt64())
	}

	var categories []string
	if diags := model.Categories.ElementsAs(context.Background(), &categories, false); diags.HasError() {
		t.Fatalf("categories decoding diagnostics = %v", diags.Errors())
	}
	if len(categories) != 2 || categories[0] != "AI" {
		t.Fatalf("categories = %v", categories)
	}

	var contexts []learningResourceContextModel
	if diags := model.Contexts.ElementsAs(context.Background(), &contexts, false); diags.HasError() {
		t.Fatalf("contexts decoding diagnostics = %v", diags.Errors())
	}
	if len(contexts) != 1 || contexts[0].PageID.ValueString() != "home" || contexts[0].Priority.ValueInt64() != 1 {
		t.Fatalf("contexts = %+v", contexts)
	}

	var sources []learningResourceSourceModel
	if diags := model.Source.ElementsAs(context.Background(), &sources, false); diags.HasError() {
		t.Fatalf("source decoding diagnostics = %v", diags.Errors())
	}
	if len(sources) != 1 || sources[0].URL.ValueString() != "https://example.com/academy" {
		t.Fatalf("source = %+v", sources)
	}
	if sources[0].Provider.ValueString() != provider {
		t.Fatalf("source provider = %q", sources[0].Provider.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model learningResourceModel
	diags := apply(context.Background(), &ometa.LearningResource{
		Id:           id,
		Name:         "bare",
		ResourceType: ometa.Video,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.Difficulty.IsNull() {
		t.Fatalf("Difficulty = %v, want null", model.Difficulty)
	}
	if !model.Status.IsNull() {
		t.Fatalf("Status = %v, want null", model.Status)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
	if model.Categories.IsNull() || len(model.Categories.Elements()) != 0 {
		t.Fatalf("Categories = %v, want an empty list", model.Categories)
	}
	if model.Contexts.IsNull() || len(model.Contexts.Elements()) != 0 {
		t.Fatalf("Contexts = %v, want an empty list", model.Contexts)
	}
	if len(model.Source.Elements()) != 1 {
		t.Fatalf("Source = %v, want a one-element list", model.Source)
	}
}

func ptr(value string) *string {
	return &value
}
