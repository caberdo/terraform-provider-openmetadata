package ml_model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// mlModelModel is the Terraform representation of an OpenMetadata ML model.
type mlModelModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	Algorithm          types.String  `tfsdk:"algorithm"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The service
// selector and the algorithm are left untouched when the server omits them.
func apply(mlModel *ometa.MlModel, model *mlModelModel) {
	model.ID = types.StringValue(mlModel.Id.String())
	model.Name = types.StringValue(mlModel.Name)
	if mlModel.Algorithm != "" {
		model.Algorithm = types.StringValue(mlModel.Algorithm)
	}
	if fqn := mlModel.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := mlModel.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	model.FullyQualifiedName = convert.StringValue(mlModel.FullyQualifiedName)
	model.DisplayName = convert.StringValue(mlModel.DisplayName)
	model.Description = convert.StringValueOr(mlModel.Description, "")
	model.Version = convert.Float64Value(mlModel.Version)
	model.UpdatedAt = convert.TimestampValue(mlModel.UpdatedAt)
	model.UpdatedBy = convert.StringValue(mlModel.UpdatedBy)
}
