package database

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// databaseModel is the Terraform representation of an OpenMetadata database.
type databaseModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The parent
// service selector is left untouched when the server omits the reference.
func apply(database *ometa.Database, model *databaseModel) {
	model.ID = types.StringValue(database.Id.String())
	model.Name = types.StringValue(database.Name)
	model.FullyQualifiedName = convert.StringValue(database.FullyQualifiedName)
	model.DisplayName = convert.StringValue(database.DisplayName)
	model.Description = convert.StringValueOr(database.Description, "")
	if fqn := database.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := database.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	model.Version = convert.Float64Value(database.Version)
	model.UpdatedAt = convert.TimestampValue(database.UpdatedAt)
	model.UpdatedBy = convert.StringValue(database.UpdatedBy)
}
