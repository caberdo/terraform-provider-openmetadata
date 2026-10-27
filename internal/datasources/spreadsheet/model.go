package spreadsheet

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// spreadsheetModel is the Terraform representation of a spreadsheet lookup. The
// id and fully_qualified_name attributes are both selectors and results.
type spreadsheetModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	Parent             types.String  `tfsdk:"parent"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Missing service
// and directory references leave the prior model value untouched.
func apply(spreadsheet *ometa.Spreadsheet, model *spreadsheetModel) {
	model.ID = types.StringValue(spreadsheet.Id.String())
	model.Name = types.StringValue(spreadsheet.Name)
	model.FullyQualifiedName = convert.StringValue(spreadsheet.FullyQualifiedName)
	model.DisplayName = convert.StringValue(spreadsheet.DisplayName)
	model.Description = convert.StringValueOr(spreadsheet.Description, "")
	if fqn := spreadsheet.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := spreadsheet.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	if spreadsheet.Directory != nil {
		if fqn := spreadsheet.Directory.FullyQualifiedName; fqn != nil && *fqn != "" {
			model.Parent = types.StringValue(*fqn)
		} else if name := spreadsheet.Directory.Name; name != nil && *name != "" {
			model.Parent = types.StringValue(*name)
		}
	}
	model.Version = convert.Float64Value(spreadsheet.Version)
	model.UpdatedAt = convert.TimestampValue(spreadsheet.UpdatedAt)
	model.UpdatedBy = convert.StringValue(spreadsheet.UpdatedBy)
}
