package persona

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// personaModel is the Terraform representation of an OpenMetadata persona.
type personaModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Default            types.Bool    `tfsdk:"default"`
	Users              types.Set     `tfsdk:"users"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Fields the
// server omits keep their configured default rather than being zeroed.
func apply(ctx context.Context, persona *ometa.Persona, model *personaModel) diag.Diagnostics {
	model.ID = types.StringValue(persona.Id.String())
	model.Name = types.StringValue(persona.Name)
	model.FullyQualifiedName = convert.StringValue(persona.FullyQualifiedName)
	model.DisplayName = convert.StringValue(persona.DisplayName)
	model.Description = convert.StringValueOr(persona.Description, "")
	model.Default = convert.BoolValueOr(persona.Default, false)
	model.Version = convert.Float64Value(persona.Version)
	model.UpdatedAt = convert.TimestampValue(persona.UpdatedAt)
	model.UpdatedBy = convert.StringValue(persona.UpdatedBy)
	return userNamesValue(ctx, persona.Users, model)
}

// userNamesValue maps assigned users to the fully qualified names the
// configuration uses. An omitted list becomes an empty set, matching what the
// server reports for a persona without users.
func userNamesValue(ctx context.Context, users *[]ometa.EntityReference, model *personaModel) diag.Diagnostics {
	names := make([]string, 0)
	if users != nil {
		for _, user := range *users {
			switch {
			case user.FullyQualifiedName != nil && *user.FullyQualifiedName != "":
				names = append(names, *user.FullyQualifiedName)
			case user.Name != nil && *user.Name != "":
				names = append(names, *user.Name)
			}
		}
	}

	set, diags := types.SetValueFrom(ctx, types.StringType, names)
	model.Users = set
	return diags
}
