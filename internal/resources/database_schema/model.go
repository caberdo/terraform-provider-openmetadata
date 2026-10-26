package database_schema

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// databaseSchemaModel is the Terraform representation of an OpenMetadata
// database schema.
type databaseSchemaModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Database           types.String  `tfsdk:"database"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The required
// database selector is left untouched when the server omits the reference.
func apply(schema *ometa.DatabaseSchema, model *databaseSchemaModel) {
	model.ID = types.StringValue(schema.Id.String())
	model.Name = types.StringValue(schema.Name)
	model.FullyQualifiedName = convert.StringValue(schema.FullyQualifiedName)
	model.DisplayName = convert.StringValue(schema.DisplayName)
	model.Description = convert.StringValueOr(schema.Description, "")
	model.Version = convert.Float64Value(schema.Version)
	model.UpdatedAt = convert.TimestampValue(schema.UpdatedAt)
	model.UpdatedBy = convert.StringValue(schema.UpdatedBy)

	if fqn := schema.Database.FullyQualifiedName; fqn != nil {
		model.Database = types.StringValue(*fqn)
	} else if name := schema.Database.Name; name != nil {
		model.Database = types.StringValue(*name)
	}
}
