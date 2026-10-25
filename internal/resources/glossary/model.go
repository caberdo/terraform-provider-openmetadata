package glossary

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// glossaryModel is the Terraform representation of an OpenMetadata glossary.
// Fields are deliberately flat: nested OpenMetadata shapes get their own
// model types as entities need them.
type glossaryModel struct {
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
func apply(glossary *ometa.Glossary, model *glossaryModel) {
	model.ID = types.StringValue(glossary.Id.String())
	model.Name = types.StringValue(glossary.Name)
	model.FullyQualifiedName = convert.StringValue(glossary.FullyQualifiedName)
	model.DisplayName = convert.StringValue(glossary.DisplayName)
	model.Description = convert.StringValue(&glossary.Description)
	model.MutuallyExclusive = convert.BoolValueOr(glossary.MutuallyExclusive, false)
	model.Version = convert.Float64Value(glossary.Version)
	model.UpdatedAt = convert.TimestampValue(glossary.UpdatedAt)
	model.UpdatedBy = convert.StringValue(glossary.UpdatedBy)
}
