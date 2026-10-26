package glossary_term

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// glossaryTermModel is the Terraform representation of a glossary term lookup.
// The id and fully_qualified_name attributes are both selectors and results.
type glossaryTermModel struct {
	ID                 types.String  `tfsdk:"id"`
	Glossary           types.String  `tfsdk:"glossary"`
	Parent             types.String  `tfsdk:"parent"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Synonyms           types.List    `tfsdk:"synonyms"`
	MutuallyExclusive  types.Bool    `tfsdk:"mutually_exclusive"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(term *ometa.GlossaryTerm, model *glossaryTermModel) {
	model.ID = types.StringValue(term.Id.String())
	model.Name = types.StringValue(term.Name)
	model.FullyQualifiedName = convert.StringValue(term.FullyQualifiedName)
	model.DisplayName = convert.StringValue(term.DisplayName)
	model.Description = convert.StringValueOr(&term.Description, "")
	model.Synonyms = stringListValue(term.Synonyms)
	model.MutuallyExclusive = convert.BoolValueOr(term.MutuallyExclusive, false)
	model.Version = convert.Float64Value(term.Version)
	model.UpdatedAt = convert.TimestampValue(term.UpdatedAt)
	model.UpdatedBy = convert.StringValue(term.UpdatedBy)

	if fqn := term.Glossary.FullyQualifiedName; fqn != nil {
		model.Glossary = types.StringValue(*fqn)
	} else if name := term.Glossary.Name; name != nil {
		model.Glossary = types.StringValue(*name)
	}

	if term.Parent != nil {
		if fqn := term.Parent.FullyQualifiedName; fqn != nil {
			model.Parent = types.StringValue(*fqn)
		} else if name := term.Parent.Name; name != nil {
			model.Parent = types.StringValue(*name)
		}
	}
}

// stringListValue maps a nullable slice to a Terraform list. A nil slice
// becomes null.
func stringListValue(values *[]string) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}
	elements := make([]attr.Value, len(*values))
	for i, value := range *values {
		elements[i] = types.StringValue(value)
	}
	return types.ListValueMust(types.StringType, elements)
}
