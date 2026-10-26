package search_index

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// searchIndexFieldObjectType describes one nested search index field returned
// by the lookup.
var searchIndexFieldObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"name":         types.StringType,
	"data_type":    types.StringType,
	"description":  types.StringType,
	"display_name": types.StringType,
}}

// searchIndexFieldModel is the Terraform representation of one OpenMetadata
// search index field returned by the lookup.
type searchIndexFieldModel struct {
	Name        types.String `tfsdk:"name"`
	DataType    types.String `tfsdk:"data_type"`
	Description types.String `tfsdk:"description"`
	DisplayName types.String `tfsdk:"display_name"`
}

// searchIndexModel is the Terraform representation of a search index lookup.
// The id and fully_qualified_name attributes are both selectors and results.
type searchIndexModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Fields             types.List    `tfsdk:"fields"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The service
// reference keeps the prior model value when the server omits it.
func apply(index *ometa.SearchIndex, model *searchIndexModel) {
	model.ID = types.StringValue(index.Id.String())
	model.Name = types.StringValue(index.Name)
	model.FullyQualifiedName = convert.StringValue(index.FullyQualifiedName)
	model.DisplayName = convert.StringValue(index.DisplayName)
	model.Description = convert.StringValueOr(index.Description, "")
	model.Fields = fieldsValue(index.Fields)
	model.Version = convert.Float64Value(index.Version)
	model.UpdatedAt = convert.TimestampValue(index.UpdatedAt)
	model.UpdatedBy = convert.StringValue(index.UpdatedBy)

	if fqn := index.Service.FullyQualifiedName; fqn != nil {
		model.Service = types.StringValue(*fqn)
	} else if name := index.Service.Name; name != nil {
		model.Service = types.StringValue(*name)
	}
}

// fieldsValue maps API fields to a Terraform list in the server's order.
func fieldsValue(fields []ometa.SearchIndexField) types.List {
	elements := make([]attr.Value, 0, len(fields))
	for _, field := range fields {
		elements = append(elements, types.ObjectValueMust(searchIndexFieldObjectType.AttrTypes, map[string]attr.Value{
			"name":         types.StringValue(field.Name),
			"data_type":    types.StringValue(string(field.DataType)),
			"description":  convert.StringValue(field.Description),
			"display_name": convert.StringValue(field.DisplayName),
		}))
	}

	return types.ListValueMust(searchIndexFieldObjectType, elements)
}
