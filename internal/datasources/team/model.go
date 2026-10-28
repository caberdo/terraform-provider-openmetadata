package team

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// teamModel is the Terraform representation of an OpenMetadata team. Every
// entity reference is flattened to a string list, because the provider works
// with fully qualified names while the API works with UUIDs.
type teamModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	TeamType           types.String  `tfsdk:"team_type"`
	Email              types.String  `tfsdk:"email"`
	IsJoinable         types.Bool    `tfsdk:"is_joinable"`
	Parents            types.Set     `tfsdk:"parents"`
	Policies           types.Set     `tfsdk:"policies"`
	DefaultRoles       types.Set     `tfsdk:"default_roles"`
	DefaultPersona     types.String  `tfsdk:"default_persona"`
	Users              types.Set     `tfsdk:"users"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. References are
// rendered as fully qualified names, and an omitted reference list becomes an
// empty set so a team without parents, policies or users does not churn. Sets
// are used because the server sorts every reference list on read.
func apply(team *ometa.Team, model *teamModel) {
	model.ID = types.StringValue(team.Id.String())
	model.Name = types.StringValue(team.Name)
	model.FullyQualifiedName = convert.StringValue(team.FullyQualifiedName)
	model.DisplayName = convert.StringValue(team.DisplayName)
	model.Description = convert.StringValueOr(team.Description, "")
	if team.TeamType != nil {
		model.TeamType = types.StringValue(string(*team.TeamType))
	}
	model.Email = convert.StringValue(team.Email)
	model.IsJoinable = convert.BoolValueOr(team.IsJoinable, true)
	model.Parents = referenceNames(team.Parents)
	model.Policies = referenceNames(team.Policies)
	model.DefaultRoles = referenceNames(team.DefaultRoles)
	model.DefaultPersona = referenceName(team.DefaultPersona)
	model.Users = referenceNames(team.Users)
	model.Version = convert.Float64Value(team.Version)
	model.UpdatedAt = convert.TimestampValue(team.UpdatedAt)
	model.UpdatedBy = convert.StringValue(team.UpdatedBy)
}

// referenceNames renders an API reference list as fully qualified names. The
// server omits an empty list, so absence becomes an empty Terraform set.
func referenceNames(references *[]ometa.EntityReference) types.Set {
	elements := make([]attr.Value, 0)
	if references != nil {
		for _, reference := range *references {
			if name := referenceIdentifier(&reference); name != "" {
				elements = append(elements, types.StringValue(name))
			}
		}
	}
	return types.SetValueMust(types.StringType, elements)
}

// referenceName renders one API reference as a fully qualified name.
func referenceName(reference *ometa.EntityReference) types.String {
	if reference == nil {
		return types.StringNull()
	}
	if name := referenceIdentifier(reference); name != "" {
		return types.StringValue(name)
	}
	return types.StringNull()
}

// referenceIdentifier prefers the fully qualified name and falls back to the
// plain name, because some relationships only carry one of the two.
func referenceIdentifier(reference *ometa.EntityReference) string {
	if reference.FullyQualifiedName != nil && *reference.FullyQualifiedName != "" {
		return *reference.FullyQualifiedName
	}
	if reference.Name != nil {
		return *reference.Name
	}
	return ""
}
