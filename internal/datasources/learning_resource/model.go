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

// learningResourceModel is the Terraform representation of a learning resource
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type learningResourceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	ResourceType       types.String  `tfsdk:"resource_type"`
	Categories         types.List    `tfsdk:"categories"`
	Contexts           types.List    `tfsdk:"contexts"`
	Source             types.List    `tfsdk:"source"`
	Difficulty         types.String  `tfsdk:"difficulty"`
	EstimatedDuration  types.Int64   `tfsdk:"estimated_duration"`
	Status             types.String  `tfsdk:"status"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
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
