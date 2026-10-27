package file

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// fileModel is the Terraform representation of an OpenMetadata file. A file
// lives inside a directory of a drive service, so both selectors are flat
// strings holding fully qualified names.
type fileModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	Directory          types.String  `tfsdk:"directory"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The service
// and directory selectors are left untouched when the server omits the
// reference, so a missing directory never zeroes state.
func apply(file *ometa.File, model *fileModel) {
	model.ID = types.StringValue(file.Id.String())
	model.Name = types.StringValue(file.Name)
	model.FullyQualifiedName = convert.StringValue(file.FullyQualifiedName)
	model.DisplayName = convert.StringValue(file.DisplayName)
	model.Description = convert.StringValueOr(file.Description, "")
	if fqn := file.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := file.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	if file.Directory != nil {
		if fqn := file.Directory.FullyQualifiedName; fqn != nil && *fqn != "" {
			model.Directory = types.StringValue(*fqn)
		} else if name := file.Directory.Name; name != nil && *name != "" {
			model.Directory = types.StringValue(*name)
		}
	}
	model.Version = convert.Float64Value(file.Version)
	model.UpdatedAt = convert.TimestampValue(file.UpdatedAt)
	model.UpdatedBy = convert.StringValue(file.UpdatedBy)
}
