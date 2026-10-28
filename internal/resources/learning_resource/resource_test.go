package learning_resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_learning_resource" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_learning_resource")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want StringAttribute", resp.Schema.Attributes["name"])
	}
	if !name.Required {
		t.Fatal("name must be required")
	}
	if len(name.PlanModifiers) == 0 {
		t.Fatal("name must carry a plan modifier")
	}

	resourceType, ok := resp.Schema.Attributes["resource_type"].(rschema.StringAttribute)
	if !ok || !resourceType.Required {
		t.Fatalf("resource_type = %#v, want a required string", resp.Schema.Attributes["resource_type"])
	}
	if len(resourceType.Validators) == 0 {
		t.Fatal("resource_type must validate the allowed values")
	}

	categories, ok := resp.Schema.Attributes["categories"].(rschema.ListAttribute)
	if !ok || !categories.Required {
		t.Fatalf("categories = %#v, want a required list", resp.Schema.Attributes["categories"])
	}
	if len(categories.Validators) == 0 {
		t.Fatal("categories must validate size and values")
	}

	contexts, ok := resp.Schema.Attributes["contexts"].(rschema.ListNestedAttribute)
	if !ok || !contexts.Required {
		t.Fatalf("contexts = %#v, want a required nested list", resp.Schema.Attributes["contexts"])
	}
	pageID, ok := contexts.NestedObject.Attributes["page_id"].(rschema.StringAttribute)
	if !ok || !pageID.Required {
		t.Fatal("contexts.page_id must be required")
	}

	source, ok := resp.Schema.Attributes["source"].(rschema.ListNestedAttribute)
	if !ok || !source.Required {
		t.Fatalf("source = %#v, want a required nested list", resp.Schema.Attributes["source"])
	}
	url, ok := source.NestedObject.Attributes["url"].(rschema.StringAttribute)
	if !ok || !url.Required {
		t.Fatal("source.url must be required")
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "data_governance_101"
	displayName := "Data Governance 101"
	provider := "user"
	difficulty := ometa.LearningResourceDifficultyIntro
	status := ometa.LearningResourceStatusActive
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	priority := int32(1)
	estimatedDuration := int32(15)

	model := learningResourceModel{CompletionThreshold: types.Float64Value(0.8)}
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
		Version:           &version,
		UpdatedAt:         &updatedAt,
		UpdatedBy:         &updatedBy,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "data_governance_101" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Introduction to governance." {
		t.Fatalf("Description = %q", model.Description.ValueString())
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
	if model.CompletionThreshold.ValueFloat64() != 0.8 {
		t.Fatalf("CompletionThreshold = %v, want the prior value", model.CompletionThreshold.ValueFloat64())
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

	var categories []string
	if diags := model.Categories.ElementsAs(context.Background(), &categories, false); diags.HasError() {
		t.Fatalf("categories decoding diagnostics = %v", diags.Errors())
	}
	if len(categories) != 2 || categories[0] != "AI" || categories[1] != "DataGovernance" {
		t.Fatalf("categories = %v", categories)
	}

	var contexts []learningResourceContextModel
	if diags := model.Contexts.ElementsAs(context.Background(), &contexts, false); diags.HasError() {
		t.Fatalf("contexts decoding diagnostics = %v", diags.Errors())
	}
	if len(contexts) != 1 || contexts[0].PageID.ValueString() != "home" || contexts[0].Priority.ValueInt64() != 1 {
		t.Fatalf("contexts = %+v", contexts)
	}
	if !contexts[0].ComponentID.IsNull() {
		t.Fatalf("ComponentID = %v, want null", contexts[0].ComponentID)
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

func TestPatchOps(t *testing.T) {
	base := learningResourceModel{
		ResourceType:        types.StringValue("Article"),
		Categories:          stringList(t, "AI"),
		Contexts:            contextsList(t, learningResourceContextModel{PageID: types.StringValue("home")}),
		Source:              sourceList(t, learningResourceSourceModel{URL: types.StringValue("https://example.com/academy")}),
		Difficulty:          types.StringValue("Intro"),
		EstimatedDuration:   types.Int64Value(15),
		CompletionThreshold: types.Float64Value(0.8),
		Status:              types.StringValue("Active"),
		DisplayName:         types.StringValue("Display"),
		Description:         types.StringValue("before"),
	}

	t.Run("no changes", func(t *testing.T) {
		ops, diags := patchOps(context.Background(), base, base)
		if diags.HasError() {
			t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
		}
		if len(ops) != 0 {
			t.Fatalf("ops = %#v, want none", ops)
		}
	})

	t.Run("display name and description changed", func(t *testing.T) {
		plan := base
		plan.DisplayName = types.StringValue("New display")
		plan.Description = types.StringValue("after")

		ops, diags := patchOps(context.Background(), plan, base)
		if diags.HasError() {
			t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
		}
		if len(ops) != 2 {
			t.Fatalf("ops = %#v, want 2", ops)
		}
		if ops[0].Path != "/displayName" || ops[0].Value != "New display" {
			t.Fatalf("first op = %#v", ops[0])
		}
		if ops[1].Path != "/description" || ops[1].Value != "after" {
			t.Fatalf("second op = %#v", ops[1])
		}
	})

	t.Run("categories changed", func(t *testing.T) {
		plan := base
		plan.Categories = stringList(t, "AI", "Discovery")

		ops, diags := patchOps(context.Background(), plan, base)
		if diags.HasError() {
			t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
		}
		if len(ops) != 1 || ops[0].Path != "/categories" {
			t.Fatalf("ops = %#v", ops)
		}
		categories, ok := ops[0].Value.([]ometa.CreateLearningResourceCategories)
		if !ok {
			t.Fatalf("value is %T", ops[0].Value)
		}
		if len(categories) != 2 || categories[1] != ometa.CreateLearningResourceCategoriesDiscovery {
			t.Fatalf("categories = %v", categories)
		}
	})

	t.Run("contexts changed", func(t *testing.T) {
		plan := base
		plan.Contexts = contextsList(t, learningResourceContextModel{
			PageID:   types.StringValue("home"),
			Priority: types.Int64Value(2),
		})

		ops, diags := patchOps(context.Background(), plan, base)
		if diags.HasError() {
			t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
		}
		if len(ops) != 1 || ops[0].Path != "/contexts" {
			t.Fatalf("ops = %#v", ops)
		}
		contexts, ok := ops[0].Value.([]ometa.LearningResourceContext)
		if !ok {
			t.Fatalf("value is %T", ops[0].Value)
		}
		if len(contexts) != 1 || contexts[0].PageId != "home" || contexts[0].Priority == nil || *contexts[0].Priority != 2 {
			t.Fatalf("contexts = %+v", contexts)
		}
	})

	t.Run("source changed", func(t *testing.T) {
		plan := base
		plan.Source = sourceList(t, learningResourceSourceModel{URL: types.StringValue("https://example.com/new")})

		ops, diags := patchOps(context.Background(), plan, base)
		if diags.HasError() {
			t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
		}
		if len(ops) != 1 || ops[0].Path != "/source" {
			t.Fatalf("ops = %#v", ops)
		}
		source, ok := ops[0].Value.(ometa.LearningResourceSource)
		if !ok {
			t.Fatalf("value is %T", ops[0].Value)
		}
		if source.Url != "https://example.com/new" {
			t.Fatalf("source = %+v", source)
		}
	})

	t.Run("optional attributes cleared", func(t *testing.T) {
		plan := base
		plan.Difficulty = types.StringNull()
		plan.EstimatedDuration = types.Int64Null()
		plan.CompletionThreshold = types.Float64Null()

		ops, diags := patchOps(context.Background(), plan, base)
		if diags.HasError() {
			t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
		}
		if len(ops) != 3 {
			t.Fatalf("ops = %#v, want 3", ops)
		}
		for i, want := range []string{"/difficulty", "/estimatedDuration", "/completionThreshold"} {
			if ops[i].Path != want || ops[i].Op != "remove" {
				t.Fatalf("op %d = %#v, want remove %s", i, ops[i], want)
			}
		}
	})
}

func stringList(t *testing.T, values ...string) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), types.StringType, values)
	if diags.HasError() {
		t.Fatalf("string list diagnostics = %v", diags.Errors())
	}
	return list
}

func contextsList(t *testing.T, values ...learningResourceContextModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), contextObjectType, values)
	if diags.HasError() {
		t.Fatalf("contexts diagnostics = %v", diags.Errors())
	}
	return list
}

func sourceList(t *testing.T, values ...learningResourceSourceModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), sourceObjectType, values)
	if diags.HasError() {
		t.Fatalf("source diagnostics = %v", diags.Errors())
	}
	return list
}

func ptr(value string) *string {
	return &value
}
