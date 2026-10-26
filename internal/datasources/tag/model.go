package tag

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// tagModel is the Terraform representation of a tag lookup. The id and
// fully_qualified_name attributes are both selectors and results.
type tagModel struct {
	ID                 types.String  `tfsdk:"id"`
	Classification     types.String  `tfsdk:"classification"`
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
func apply(tag *ometa.Tag, model *tagModel) {
	model.ID = types.StringValue(tag.Id.String())
	model.Name = types.StringValue(tag.Name)
	model.FullyQualifiedName = convert.StringValue(tag.FullyQualifiedName)
	model.DisplayName = convert.StringValue(tag.DisplayName)
	model.Description = convert.StringValueOr(&tag.Description, "")
	if tag.Classification != nil {
		if fqn := tag.Classification.FullyQualifiedName; fqn != nil {
			model.Classification = types.StringValue(*fqn)
		} else if name := tag.Classification.Name; name != nil {
			model.Classification = types.StringValue(*name)
		}
	}
	model.MutuallyExclusive = convert.BoolValueOr(tag.MutuallyExclusive, false)
	model.Version = convert.Float64Value(tag.Version)
	model.UpdatedAt = convert.TimestampValue(tag.UpdatedAt)
	model.UpdatedBy = convert.StringValue(tag.UpdatedBy)
}
