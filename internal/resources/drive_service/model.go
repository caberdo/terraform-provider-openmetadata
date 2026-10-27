package drive_service

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// driveServiceModel is the Terraform representation of an OpenMetadata drive
// service.
type driveServiceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	ServiceType        types.String  `tfsdk:"service_type"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. A service type
// the server omits keeps the prior model value instead of being zeroed.
func apply(driveService *ometa.DriveService, model *driveServiceModel) {
	model.ID = types.StringValue(driveService.Id.String())
	model.Name = types.StringValue(driveService.Name)
	if driveService.ServiceType != "" {
		model.ServiceType = types.StringValue(string(driveService.ServiceType))
	}
	model.FullyQualifiedName = convert.StringValue(driveService.FullyQualifiedName)
	model.DisplayName = convert.StringValue(driveService.DisplayName)
	model.Description = convert.StringValueOr(driveService.Description, "")
	model.Version = convert.Float64Value(driveService.Version)
	model.UpdatedAt = convert.TimestampValue(driveService.UpdatedAt)
	model.UpdatedBy = convert.StringValue(driveService.UpdatedBy)
}
