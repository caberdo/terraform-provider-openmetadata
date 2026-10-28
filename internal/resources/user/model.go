package user

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// userModel is the Terraform representation of an OpenMetadata user.
// jwt_token_expiry is write-only for bot users: the API response carries a JWT
// secret inside the authentication mechanism, so the mechanism is never read
// back into state.
type userModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Email              types.String  `tfsdk:"email"`
	IsAdmin            types.Bool    `tfsdk:"is_admin"`
	IsBot              types.Bool    `tfsdk:"is_bot"`
	JWTTokenExpiry     types.String  `tfsdk:"jwt_token_expiry"`
	Teams              types.Set     `tfsdk:"teams"`
	Roles              types.Set     `tfsdk:"roles"`
	DefaultPersona     types.String  `tfsdk:"default_persona"`
	Timezone           types.String  `tfsdk:"timezone"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The
// authentication mechanism is deliberately ignored so a JWT secret can never
// reach Terraform state.
func apply(user *ometa.User, model *userModel) {
	model.ID = types.StringValue(user.Id.String())
	model.Name = types.StringValue(user.Name)
	model.FullyQualifiedName = convert.StringValue(user.FullyQualifiedName)
	model.DisplayName = convert.StringValue(user.DisplayName)
	model.Description = convert.StringValueOr(user.Description, "")
	model.Email = types.StringValue(user.Email)
	model.IsAdmin = convert.BoolValueOr(user.IsAdmin, false)
	model.IsBot = convert.BoolValueOr(user.IsBot, false)
	model.Teams = referenceNamesValue(user.Teams)
	model.Roles = referenceNamesValue(user.Roles)
	model.DefaultPersona = referenceNameValue(user.DefaultPersona)
	model.Timezone = convert.StringValue(user.Timezone)
	model.Version = convert.Float64Value(user.Version)
	model.UpdatedAt = convert.TimestampValue(user.UpdatedAt)
	model.UpdatedBy = convert.StringValue(user.UpdatedBy)
}

// referenceNamesValue maps entity references to their fully qualified names.
// An absent list becomes an empty list, matching the schema's computed shape.
func referenceNamesValue(references *[]ometa.EntityReference) types.Set {
	if references == nil || len(*references) == 0 {
		return types.SetValueMust(types.StringType, nil)
	}

	elements := make([]attr.Value, 0, len(*references))
	for _, reference := range *references {
		name := reference.FullyQualifiedName
		if name == nil {
			name = reference.Name
		}
		if name == nil {
			continue
		}
		elements = append(elements, types.StringValue(*name))
	}
	return types.SetValueMust(types.StringType, elements)
}

// referenceNameValue maps one entity reference to its fully qualified name.
func referenceNameValue(reference *ometa.EntityReference) types.String {
	if reference == nil {
		return types.StringNull()
	}
	if reference.FullyQualifiedName != nil {
		return types.StringValue(*reference.FullyQualifiedName)
	}
	if reference.Name != nil {
		return types.StringValue(*reference.Name)
	}
	return types.StringNull()
}
