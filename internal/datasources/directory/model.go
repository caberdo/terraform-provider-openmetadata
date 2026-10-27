package directory

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// directoryModel is the Terraform representation of a directory lookup. The id
// and fully_qualified_name attributes are both selectors and results.
type directoryModel struct {
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

// apply copies the API representation into the Terraform model. Missing
// service and parent references leave the prior model value untouched.
func apply(directory *ometa.Directory, model *directoryModel) {
	model.ID = types.StringValue(directory.Id.String())
	model.Name = types.StringValue(directory.Name)
	model.FullyQualifiedName = convert.StringValue(directory.FullyQualifiedName)
	model.DisplayName = convert.StringValue(directory.DisplayName)
	model.Description = convert.StringValueOr(directory.Description, "")
	if fqn := directory.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := directory.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	if directory.Parent != nil {
		if fqn := directory.Parent.FullyQualifiedName; fqn != nil && *fqn != "" {
			model.Parent = types.StringValue(*fqn)
		} else if name := directory.Parent.Name; name != nil && *name != "" {
			model.Parent = types.StringValue(*name)
		}
	}
	model.Version = convert.Float64Value(directory.Version)
	model.UpdatedAt = convert.TimestampValue(directory.UpdatedAt)
	model.UpdatedBy = convert.StringValue(directory.UpdatedBy)
}
