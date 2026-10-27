package test_definition

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// testDefinitionModel is the Terraform representation of a test definition
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type testDefinitionModel struct {
	ID                   types.String  `tfsdk:"id"`
	Name                 types.String  `tfsdk:"name"`
	FullyQualifiedName   types.String  `tfsdk:"fully_qualified_name"`
	DisplayName          types.String  `tfsdk:"display_name"`
	Description          types.String  `tfsdk:"description"`
	EntityType           types.String  `tfsdk:"entity_type"`
	TestPlatforms        types.List    `tfsdk:"test_platforms"`
	DataQualityDimension types.String  `tfsdk:"data_quality_dimension"`
	SupportedDataTypes   types.List    `tfsdk:"supported_data_types"`
	SupportedServices    types.List    `tfsdk:"supported_services"`
	SQLExpression        types.String  `tfsdk:"sql_expression"`
	Provider             types.String  `tfsdk:"source"`
	Version              types.Float64 `tfsdk:"version"`
	UpdatedAt            types.String  `tfsdk:"updated_at"`
	UpdatedBy            types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Every field of
// the data source is computed, so an omitted value becomes null.
func apply(definition *ometa.TestDefinition, model *testDefinitionModel) {
	model.ID = types.StringNull()
	if definition.Id != nil {
		model.ID = types.StringValue(definition.Id.String())
	}
	model.Name = types.StringValue(definition.Name)
	model.FullyQualifiedName = convert.StringValue(definition.FullyQualifiedName)
	model.DisplayName = convert.StringValue(definition.DisplayName)
	model.Description = convert.StringValueOr(definition.Description, "")
	model.EntityType = types.StringNull()
	if definition.EntityType != nil {
		model.EntityType = types.StringValue(string(*definition.EntityType))
	}
	model.TestPlatforms = types.ListNull(types.StringType)
	if definition.TestPlatforms != nil {
		model.TestPlatforms = platformListValue(definition.TestPlatforms)
	}
	model.DataQualityDimension = convert.StringValue(dimensionValue(definition.DataQualityDimension))
	model.SupportedDataTypes = types.ListNull(types.StringType)
	if definition.SupportedDataTypes != nil && len(*definition.SupportedDataTypes) > 0 {
		model.SupportedDataTypes = dataTypeListValue(definition.SupportedDataTypes)
	}
	model.SupportedServices = types.ListNull(types.StringType)
	if definition.SupportedServices != nil && len(*definition.SupportedServices) > 0 {
		model.SupportedServices = stringListValue(definition.SupportedServices)
	}
	model.SQLExpression = convert.StringValue(definition.SqlExpression)
	model.Provider = convert.StringValue(providerValue(definition.Provider))
	model.Version = convert.Float64Value(definition.Version)
	model.UpdatedAt = convert.TimestampValue(definition.UpdatedAt)
	model.UpdatedBy = convert.StringValue(definition.UpdatedBy)
}

// dimensionValue renders an optional enum as a Terraform string pointer.
func dimensionValue(value *ometa.TestDefinitionDataQualityDimension) *string {
	if value == nil {
		return nil
	}
	rendered := string(*value)
	return &rendered
}

// providerValue renders an optional enum as a Terraform string pointer.
func providerValue(value *ometa.TestDefinitionProvider) *string {
	if value == nil {
		return nil
	}
	rendered := string(*value)
	return &rendered
}

// platformListValue maps the platform slice to a Terraform list.
func platformListValue(values []ometa.TestDefinitionTestPlatforms) types.List {
	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(string(value))
	}
	return types.ListValueMust(types.StringType, elements)
}

// dataTypeListValue maps an optional data-type slice to a Terraform list. A nil
// slice becomes null.
func dataTypeListValue(values *[]ometa.TestDefinitionSupportedDataTypes) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}
	elements := make([]attr.Value, len(*values))
	for i, value := range *values {
		elements[i] = types.StringValue(string(value))
	}
	return types.ListValueMust(types.StringType, elements)
}

// stringListValue maps an optional string slice to a Terraform list. A nil
// slice becomes null.
func stringListValue(values *[]string) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}
	elements := make([]attr.Value, len(*values))
	for i, value := range *values {
		elements[i] = types.StringValue(value)
	}
	return types.ListValueMust(types.StringType, elements)
}
