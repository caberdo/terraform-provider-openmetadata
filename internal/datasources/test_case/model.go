package test_case

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// testCaseParameterValueObjectType describes one nested test-case parameter
// returned by the lookup.
var testCaseParameterValueObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}}

// testCaseParameterValueModel is the shape ElementsAs decodes a nested
// parameter into.
type testCaseParameterValueModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// testCaseModel is the Terraform representation of a test case lookup. The id
// and fully_qualified_name attributes are both selectors and results.
type testCaseModel struct {
	ID                          types.String  `tfsdk:"id"`
	Name                        types.String  `tfsdk:"name"`
	FullyQualifiedName          types.String  `tfsdk:"fully_qualified_name"`
	DisplayName                 types.String  `tfsdk:"display_name"`
	Description                 types.String  `tfsdk:"description"`
	EntityLink                  types.String  `tfsdk:"entity_link"`
	TestDefinition              types.String  `tfsdk:"test_definition"`
	EntityFQN                   types.String  `tfsdk:"entity_fqn"`
	TestSuite                   types.String  `tfsdk:"test_suite"`
	ParameterValues             types.List    `tfsdk:"parameter_values"`
	DimensionColumns            types.List    `tfsdk:"dimension_columns"`
	ComputePassedFailedRowCount types.Bool    `tfsdk:"compute_passed_failed_row_count"`
	UseDynamicAssertion         types.Bool    `tfsdk:"use_dynamic_assertion"`
	TopDimensions               types.Int64   `tfsdk:"top_dimensions"`
	TestCaseStatus              types.String  `tfsdk:"test_case_status"`
	Version                     types.Float64 `tfsdk:"version"`
	UpdatedAt                   types.String  `tfsdk:"updated_at"`
	UpdatedBy                   types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Fields the
// server leaves out are preserved rather than zeroed.
func apply(testCase *ometa.TestCase, model *testCaseModel) {
	if testCase.Id != nil {
		model.ID = types.StringValue(testCase.Id.String())
	}
	model.Name = types.StringValue(testCase.Name)
	model.FullyQualifiedName = convert.StringValue(testCase.FullyQualifiedName)
	model.DisplayName = convert.StringValue(testCase.DisplayName)
	model.Description = convert.StringValueOr(testCase.Description, "")
	model.EntityLink = types.StringValue(testCase.EntityLink)
	model.EntityFQN = convert.StringValue(testCase.EntityFQN)
	model.TestDefinition = entityReferenceFQN(testCase.TestDefinition)
	model.TestSuite = entityReferenceFQN(testCase.TestSuite)
	model.ComputePassedFailedRowCount = convert.BoolValueOr(testCase.ComputePassedFailedRowCount, false)
	model.UseDynamicAssertion = convert.BoolValueOr(testCase.UseDynamicAssertion, false)
	model.TopDimensions = convert.Int32Value(testCase.TopDimensions)
	model.TestCaseStatus = testCaseStatusValue(testCase.TestCaseStatus)
	model.Version = convert.Float64Value(testCase.Version)
	model.UpdatedAt = convert.TimestampValue(testCase.UpdatedAt)
	model.UpdatedBy = convert.StringValue(testCase.UpdatedBy)

	model.ParameterValues = types.ListNull(testCaseParameterValueObjectType)
	if testCase.ParameterValues != nil {
		model.ParameterValues = parameterValuesValue(*testCase.ParameterValues)
	}
	model.DimensionColumns = types.ListNull(types.StringType)
	if testCase.DimensionColumns != nil {
		model.DimensionColumns = stringListValue(*testCase.DimensionColumns)
	}
}

// entityReferenceFQN renders a required entity reference, preferring the fully
// qualified name and falling back to the name. When the server omits both the
// attribute becomes null so the computed value stays known.
func entityReferenceFQN(reference ometa.EntityReference) types.String {
	if reference.FullyQualifiedName != nil {
		return types.StringValue(*reference.FullyQualifiedName)
	}
	if reference.Name != nil {
		return types.StringValue(*reference.Name)
	}
	return types.StringNull()
}

// testCaseStatusValue maps the optional status enum. A missing status stays
// null instead of pretending the test passed or failed.
func testCaseStatusValue(status *ometa.TestCaseTestCaseStatus) types.String {
	if status == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*status))
}

// parameterValuesValue maps API parameters to a Terraform list in the server's
// order.
func parameterValuesValue(values []ometa.TestCaseParameterValue) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, parameter := range values {
		elements = append(elements, types.ObjectValueMust(testCaseParameterValueObjectType.AttrTypes, map[string]attr.Value{
			"name":  convert.StringValue(parameter.Name),
			"value": convert.StringValue(parameter.Value),
		}))
	}

	return types.ListValueMust(testCaseParameterValueObjectType, elements)
}

// stringListValue maps a string slice to a Terraform list.
func stringListValue(values []string) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}

	return types.ListValueMust(types.StringType, elements)
}
