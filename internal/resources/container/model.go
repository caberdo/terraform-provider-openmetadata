package container

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// containerModel is the Terraform representation of an OpenMetadata container.
type containerModel struct {
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

// apply copies the API representation into the Terraform model. The service
// and parent selectors are left untouched when the server omits the
// reference, so an omitted parent never zeroes state.
func apply(container *ometa.Container, model *containerModel) {
	model.ID = types.StringValue(container.Id.String())
	model.Name = types.StringValue(container.Name)
	model.FullyQualifiedName = convert.StringValue(container.FullyQualifiedName)
	model.DisplayName = convert.StringValue(container.DisplayName)
	model.Description = convert.StringValueOr(container.Description, "")
	if fqn := container.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := container.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	if container.Parent != nil {
		if fqn := container.Parent.FullyQualifiedName; fqn != nil && *fqn != "" {
			model.Parent = types.StringValue(*fqn)
		} else if name := container.Parent.Name; name != nil && *name != "" {
			model.Parent = types.StringValue(*name)
		}
	}
	model.Version = convert.Float64Value(container.Version)
	model.UpdatedAt = convert.TimestampValue(container.UpdatedAt)
	model.UpdatedBy = convert.StringValue(container.UpdatedBy)
}
