package stored_procedure

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// storedProcedureModel is the Terraform representation of an OpenMetadata
// stored procedure.
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

// apply copies the API representation into the Terraform model. Required
// selectors and fields the server omits keep their prior value.
func apply(storedProcedure *ometa.StoredProcedure, model *storedProcedureModel) {
	model.ID = types.StringValue(storedProcedure.Id.String())
	model.Name = types.StringValue(storedProcedure.Name)
	model.FullyQualifiedName = convert.StringValue(storedProcedure.FullyQualifiedName)
	model.DisplayName = convert.StringValue(storedProcedure.DisplayName)
	model.Description = convert.StringValueOr(storedProcedure.Description, "")

	// Older servers return the body only in the deprecated top-level code
	// field, so fall back to it when the nested object has no code.
	if code := storedProcedure.StoredProcedureCode.Code; code != nil {
		model.Code = types.StringValue(*code)
	} else if storedProcedure.Code != nil {
		model.Code = types.StringValue(*storedProcedure.Code)
	}

	if language := storedProcedure.StoredProcedureCode.Language; language != nil {
		model.Language = types.StringValue(string(*language))
	}
	if procedureType := storedProcedure.StoredProcedureType; procedureType != nil {
		model.StoredProcedureType = types.StringValue(string(*procedureType))
	}

	if parent := storedProcedure.DatabaseSchema; parent != nil {
		if fqn := parent.FullyQualifiedName; fqn != nil {
			model.DatabaseSchema = types.StringValue(*fqn)
		} else if name := parent.Name; name != nil {
			model.DatabaseSchema = types.StringValue(*name)
		}
	}

	model.Version = convert.Float64Value(storedProcedure.Version)
	model.UpdatedAt = convert.TimestampValue(storedProcedure.UpdatedAt)
	model.UpdatedBy = convert.StringValue(storedProcedure.UpdatedBy)
}
