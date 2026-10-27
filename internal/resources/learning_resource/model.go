package learning_resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// contextObjectType describes one learning resource context. The resource and
// the data source share this shape so a context read from the API can be
// written straight back into state.
var contextObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"page_id":      types.StringType,
	"component_id": types.StringType,
	"priority":     types.Int64Type,
}}

// sourceObjectType describes the single source of a learning resource.
var sourceObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"url":      types.StringType,
	"provider": types.StringType,
}}

// learningResourceContextModel is the Terraform representation of one
// OpenMetadata learning resource context.
type learningResourceContextModel struct {
	PageID      types.String `tfsdk:"page_id"`
	ComponentID types.String `tfsdk:"component_id"`
	Priority    types.Int64  `tfsdk:"priority"`
}

// learningResourceSourceModel is the Terraform representation of the source a
// learning resource points at.
type learningResourceSourceModel struct {
	URL      types.String `tfsdk:"url"`
	Provider types.String `tfsdk:"provider"`
}

// learningResourceModel is the Terraform representation of an OpenMetadata
// learning resource.
type learningResourceModel struct {
	ID                  types.String  `tfsdk:"id"`
	Name                types.String  `tfsdk:"name"`
	ResourceType        types.String  `tfsdk:"resource_type"`
	Categories          types.List    `tfsdk:"categories"`
	Contexts            types.List    `tfsdk:"contexts"`
	Source              types.List    `tfsdk:"source"`
	Difficulty          types.String  `tfsdk:"difficulty"`
	EstimatedDuration   types.Int64   `tfsdk:"estimated_duration"`
	CompletionThreshold types.Float64 `tfsdk:"completion_threshold"`
	Status              types.String  `tfsdk:"status"`
	FullyQualifiedName  types.String  `tfsdk:"fully_qualified_name"`
	DisplayName         types.String  `tfsdk:"display_name"`
	Description         types.String  `tfsdk:"description"`
	Version             types.Float64 `tfsdk:"version"`
	UpdatedAt           types.String  `tfsdk:"updated_at"`
	UpdatedBy           types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Difficulty and
// status keep the prior model value when the server omits them, and the nested
// lists mirror the server's order.
func apply(ctx context.Context, learningResource *ometa.LearningResource, model *learningResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(learningResource.Id.String())
	model.Name = types.StringValue(learningResource.Name)
	model.FullyQualifiedName = convert.StringValue(learningResource.FullyQualifiedName)
	model.DisplayName = convert.StringValue(learningResource.DisplayName)
	model.Description = convert.StringValueOr(learningResource.Description, "")
	model.ResourceType = types.StringValue(string(learningResource.ResourceType))
	if learningResource.Difficulty != nil {
		model.Difficulty = types.StringValue(string(*learningResource.Difficulty))
	} else if model.Difficulty.IsUnknown() {
		model.Difficulty = types.StringNull()
	}
	if learningResource.Status != nil {
		model.Status = types.StringValue(string(*learningResource.Status))
	} else if model.Status.IsUnknown() {
		model.Status = types.StringNull()
	}
	model.EstimatedDuration = convert.Int32Value(learningResource.EstimatedDuration)
	// The read model does not expose completionThreshold. A planned value is
	// kept; an omitted one resolves to null so no unknown survives apply.
	if model.CompletionThreshold.IsUnknown() {
		model.CompletionThreshold = types.Float64Null()
	}
	model.Version = convert.Float64Value(learningResource.Version)
	model.UpdatedAt = convert.TimestampValue(learningResource.UpdatedAt)
	model.UpdatedBy = convert.StringValue(learningResource.UpdatedBy)

	var diags diag.Diagnostics

	categories := make([]string, 0, len(learningResource.Categories))
	for _, category := range learningResource.Categories {
		categories = append(categories, string(category))
	}
	categoryList, categoryDiags := types.ListValueFrom(ctx, types.StringType, categories)
	diags.Append(categoryDiags...)
	model.Categories = categoryList

	contexts := make([]learningResourceContextModel, 0, len(learningResource.Contexts))
	for _, contextItem := range learningResource.Contexts {
		contexts = append(contexts, learningResourceContextModel{
			PageID:      types.StringValue(contextItem.PageId),
			ComponentID: convert.StringValue(contextItem.ComponentId),
			Priority:    convert.Int32Value(contextItem.Priority),
		})
	}
	contextList, contextDiags := types.ListValueFrom(ctx, contextObjectType, contexts)
	diags.Append(contextDiags...)
	model.Contexts = contextList

	sources := []learningResourceSourceModel{{
		URL:      types.StringValue(learningResource.Source.Url),
		Provider: convert.StringValue(learningResource.Source.Provider),
	}}
	sourceList, sourceDiags := types.ListValueFrom(ctx, sourceObjectType, sources)
	diags.Append(sourceDiags...)
	model.Source = sourceList

	return diags
}

// createBody maps the planned model to the SDK create request.
func createBody(ctx context.Context, plan learningResourceModel) (*ometa.CreateLearningResource, diag.Diagnostics) {
	categories, diags := categoriesToAPI(ctx, plan.Categories)
	if diags.HasError() {
		return nil, diags
	}
	contexts, contextDiags := contextsToAPI(ctx, plan.Contexts)
	diags.Append(contextDiags...)
	if diags.HasError() {
		return nil, diags
	}
	source, sourceDiags := sourceToAPI(ctx, plan.Source)
	diags.Append(sourceDiags...)
	if diags.HasError() {
		return nil, diags
	}

	body := &ometa.CreateLearningResource{
		Name:                plan.Name.ValueString(),
		ResourceType:        ometa.CreateLearningResourceResourceType(plan.ResourceType.ValueString()),
		Categories:          categories,
		Contexts:            contexts,
		Source:              source,
		Description:         convert.StringPointer(plan.Description),
		DisplayName:         convert.StringPointer(plan.DisplayName),
		EstimatedDuration:   int32Pointer(plan.EstimatedDuration),
		CompletionThreshold: float64Pointer(plan.CompletionThreshold),
		Difficulty:          difficultyPointer(plan.Difficulty),
		Status:              statusPointer(plan.Status),
	}
	return body, diags
}

// categoriesToAPI maps a Terraform list of category names to the SDK enum
// slice. A null list yields nil.
func categoriesToAPI(ctx context.Context, value types.List) ([]ometa.CreateLearningResourceCategories, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var names []string
	diags := value.ElementsAs(ctx, &names, false)
	if diags.HasError() {
		return nil, diags
	}

	categories := make([]ometa.CreateLearningResourceCategories, 0, len(names))
	for _, name := range names {
		categories = append(categories, ometa.CreateLearningResourceCategories(name))
	}
	return categories, diags
}

// contextsToAPI maps the planned contexts to the SDK shape.
func contextsToAPI(ctx context.Context, value types.List) ([]ometa.LearningResourceContext, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var models []learningResourceContextModel
	diags := value.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}

	contexts := make([]ometa.LearningResourceContext, 0, len(models))
	for _, model := range models {
		contexts = append(contexts, ometa.LearningResourceContext{
			PageId:      model.PageID.ValueString(),
			ComponentId: convert.StringPointer(model.ComponentID),
			Priority:    int32Pointer(model.Priority),
		})
	}
	return contexts, diags
}

// sourceToAPI maps the single planned source to the SDK shape.
func sourceToAPI(ctx context.Context, value types.List) (ometa.LearningResourceSource, diag.Diagnostics) {
	var source ometa.LearningResourceSource
	if value.IsNull() || value.IsUnknown() {
		return source, nil
	}

	var models []learningResourceSourceModel
	diags := value.ElementsAs(ctx, &models, false)
	if diags.HasError() || len(models) == 0 {
		return source, diags
	}

	source.Url = models[0].URL.ValueString()
	source.Provider = convert.StringPointer(models[0].Provider)
	return source, diags
}

// difficultyPointer maps an optional difficulty to the SDK enum.
func difficultyPointer(value types.String) *ometa.CreateLearningResourceDifficulty {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	difficulty := ometa.CreateLearningResourceDifficulty(value.ValueString())
	return &difficulty
}

// statusPointer maps an optional status to the SDK enum.
func statusPointer(value types.String) *ometa.CreateLearningResourceStatus {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	status := ometa.CreateLearningResourceStatus(value.ValueString())
	return &status
}

// int32Pointer maps an optional Terraform number to the SDK's int32 pointer.
// Null and unknown values are omitted from the request.
func int32Pointer(value types.Int64) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := int32(value.ValueInt64())
	return &v
}

// float64Pointer maps an optional Terraform number to a float pointer. Null and
// unknown values are omitted from the request.
func float64Pointer(value types.Float64) *float64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueFloat64()
	return &v
}
