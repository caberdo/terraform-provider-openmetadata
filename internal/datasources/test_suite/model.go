package test_suite

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// testSuiteModel is the Terraform representation of a test suite lookup. The
// id and fully_qualified_name attributes are both selectors and results.
type testSuiteModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Basic              types.Bool    `tfsdk:"basic"`
	Executable         types.Bool    `tfsdk:"executable"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(testSuite *ometa.TestSuite, model *testSuiteModel) {
	if testSuite.Id != nil {
		model.ID = types.StringValue(testSuite.Id.String())
	}
	model.Name = types.StringValue(testSuite.Name)
	model.FullyQualifiedName = convert.StringValue(testSuite.FullyQualifiedName)
	model.DisplayName = convert.StringValue(testSuite.DisplayName)
	model.Description = convert.StringValueOr(testSuite.Description, "")
	model.Basic = convert.BoolValueOr(testSuite.Basic, false)
	model.Executable = convert.BoolValueOr(testSuite.Executable, false)
	model.Version = convert.Float64Value(testSuite.Version)
	model.UpdatedAt = convert.TimestampValue(testSuite.UpdatedAt)
	model.UpdatedBy = convert.StringValue(testSuite.UpdatedBy)
}
