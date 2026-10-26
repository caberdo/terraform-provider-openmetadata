package ml_model_service

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// mlModelServiceModel is the Terraform representation of an OpenMetadata ML
// model service.
type mlModelServiceModel struct {
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
func apply(service *ometa.MlModelService, model *mlModelServiceModel) {
	model.ID = types.StringValue(service.Id.String())
	model.Name = types.StringValue(service.Name)
	if service.ServiceType != "" {
		model.ServiceType = types.StringValue(string(service.ServiceType))
	}
	model.FullyQualifiedName = convert.StringValue(service.FullyQualifiedName)
	model.DisplayName = convert.StringValue(service.DisplayName)
	model.Description = convert.StringValueOr(service.Description, "")
	model.Version = convert.Float64Value(service.Version)
	model.UpdatedAt = convert.TimestampValue(service.UpdatedAt)
	model.UpdatedBy = convert.StringValue(service.UpdatedBy)
}
