package stored_procedure

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// storedProcedureModel is the Terraform representation of a stored procedure
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type storedProcedureModel struct {
	ID                  types.String  `tfsdk:"id"`
	Name                types.String  `tfsdk:"name"`
	DatabaseSchema      types.String  `tfsdk:"database_schema"`
	FullyQualifiedName  types.String  `tfsdk:"fully_qualified_name"`
	DisplayName         types.String  `tfsdk:"display_name"`
	Description         types.String  `tfsdk:"description"`
	Code                types.String  `tfsdk:"code"`
	Language            types.String  `tfsdk:"language"`
	StoredProcedureType types.String  `tfsdk:"stored_procedure_type"`
	Version             types.Float64 `tfsdk:"version"`
	UpdatedAt           types.String  `tfsdk:"updated_at"`
	UpdatedBy           types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(storedProcedure *ometa.StoredProcedure, model *storedProcedureModel) {
	model.ID = types.StringValue(storedProcedure.Id.String())
	model.Name = types.StringValue(storedProcedure.Name)
	model.FullyQualifiedName = convert.StringValue(storedProcedure.FullyQualifiedName)
	model.DisplayName = convert.StringValue(storedProcedure.DisplayName)
	model.Description = convert.StringValueOr(storedProcedure.Description, "")

	if code := storedProcedure.StoredProcedureCode.Code; code != nil {
		model.Code = types.StringValue(*code)
	} else {
		model.Code = convert.StringValue(storedProcedure.Code)
	}

	if language := storedProcedure.StoredProcedureCode.Language; language != nil {
		model.Language = types.StringValue(string(*language))
	} else {
		model.Language = types.StringNull()
	}
	if procedureType := storedProcedure.StoredProcedureType; procedureType != nil {
		model.StoredProcedureType = types.StringValue(string(*procedureType))
	} else {
		model.StoredProcedureType = types.StringNull()
	}

	if parent := storedProcedure.DatabaseSchema; parent != nil {
		if fqn := parent.FullyQualifiedName; fqn != nil {
			model.DatabaseSchema = types.StringValue(*fqn)
		} else {
			model.DatabaseSchema = convert.StringValue(parent.Name)
		}
	} else {
		model.DatabaseSchema = types.StringNull()
	}

	model.Version = convert.Float64Value(storedProcedure.Version)
	model.UpdatedAt = convert.TimestampValue(storedProcedure.UpdatedAt)
	model.UpdatedBy = convert.StringValue(storedProcedure.UpdatedBy)
}
