package role

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// roleModel is the Terraform representation of a role lookup. The id and
// fully_qualified_name attributes are both selectors and results.
type roleModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Policies           types.Set     `tfsdk:"policies"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(role *ometa.Role, model *roleModel) {
	model.ID = types.StringValue(role.Id.String())
	model.Name = types.StringValue(role.Name)
	model.FullyQualifiedName = convert.StringValue(role.FullyQualifiedName)
	model.DisplayName = convert.StringValue(role.DisplayName)
	model.Description = convert.StringValueOr(role.Description, "")
	if role.Policies != nil {
		model.Policies = policySetValue(*role.Policies)
	}
	model.Version = convert.Float64Value(role.Version)
	model.UpdatedAt = convert.TimestampValue(role.UpdatedAt)
	model.UpdatedBy = convert.StringValue(role.UpdatedBy)
}

// policySetValue maps policy references to the fully qualified names shown to
// Terraform users.
func policySetValue(references []ometa.EntityReference) types.Set {
	elements := make([]attr.Value, 0, len(references))
	for _, reference := range references {
		if fqn := reference.FullyQualifiedName; fqn != nil {
			elements = append(elements, types.StringValue(*fqn))
		} else if name := reference.Name; name != nil {
			elements = append(elements, types.StringValue(*name))
		}
	}
	return types.SetValueMust(types.StringType, elements)
}
