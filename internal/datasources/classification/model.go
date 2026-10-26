package classification

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// classificationModel is the Terraform representation of a classification
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type classificationModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	MutuallyExclusive  types.Bool    `tfsdk:"mutually_exclusive"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(classification *ometa.Classification, model *classificationModel) {
	model.ID = types.StringValue(classification.Id.String())
	model.Name = types.StringValue(classification.Name)
	model.FullyQualifiedName = convert.StringValue(classification.FullyQualifiedName)
	model.DisplayName = convert.StringValue(classification.DisplayName)
	model.Description = convert.StringValueOr(&classification.Description, "")
	model.MutuallyExclusive = convert.BoolValueOr(classification.MutuallyExclusive, false)
	model.Version = convert.Float64Value(classification.Version)
	model.UpdatedAt = convert.TimestampValue(classification.UpdatedAt)
	model.UpdatedBy = convert.StringValue(classification.UpdatedBy)
}
