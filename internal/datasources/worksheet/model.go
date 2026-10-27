package worksheet

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// worksheetModel is the Terraform representation of a worksheet lookup. The id
// and fully_qualified_name attributes are both selectors and results.
type worksheetModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Spreadsheet        types.String  `tfsdk:"spreadsheet"`
	Service            types.String  `tfsdk:"service"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Missing
// spreadsheet and service references leave the prior model value untouched.
func apply(worksheet *ometa.Worksheet, model *worksheetModel) {
	model.ID = types.StringValue(worksheet.Id.String())
	model.Name = types.StringValue(worksheet.Name)
	model.FullyQualifiedName = convert.StringValue(worksheet.FullyQualifiedName)
	model.DisplayName = convert.StringValue(worksheet.DisplayName)
	model.Description = convert.StringValueOr(worksheet.Description, "")
	if fqn := worksheet.Spreadsheet.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Spreadsheet = types.StringValue(*fqn)
	} else if name := worksheet.Spreadsheet.Name; name != nil && *name != "" {
		model.Spreadsheet = types.StringValue(*name)
	}
	if fqn := worksheet.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := worksheet.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	model.Version = convert.Float64Value(worksheet.Version)
	model.UpdatedAt = convert.TimestampValue(worksheet.UpdatedAt)
	model.UpdatedBy = convert.StringValue(worksheet.UpdatedBy)
}
