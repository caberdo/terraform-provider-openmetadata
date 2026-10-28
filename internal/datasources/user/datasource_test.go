package user

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_user" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_user")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"name",
		"display_name",
		"description",
		"email",
		"is_admin",
		"is_bot",
		"teams",
		"roles",
		"default_persona",
		"timezone",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be an Optional+Computed selector", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "jane"
	displayName := "Jane Doe"
	description := "Engineer"
	timezone := "Europe/Amsterdam"
	teamFQN := "engineering"
	roleFQN := "data-steward"
	personaFQN := "default-persona"
	isAdmin := true
	isBot := false

	var model userModel
	apply(&ometa.User{
		Id:                 id,
		Name:               "jane",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Email:              "jane@example.com",
		IsAdmin:            &isAdmin,
		IsBot:              &isBot,
		Teams:              &[]ometa.EntityReference{{Id: uuid.New(), Type: "team", FullyQualifiedName: &teamFQN}},
		Roles:              &[]ometa.EntityReference{{Id: uuid.New(), Type: "role", FullyQualifiedName: &roleFQN}},
		DefaultPersona:     &ometa.EntityReference{Id: uuid.New(), Type: "persona", FullyQualifiedName: &personaFQN},
		Timezone:           &timezone,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "jane" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Email.ValueString() != "jane@example.com" {
		t.Fatalf("Email = %q", model.Email.ValueString())
	}
	if model.Description.ValueString() != "Engineer" {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if !model.IsAdmin.ValueBool() || model.IsBot.ValueBool() {
		t.Fatalf("IsAdmin = %v, IsBot = %v", model.IsAdmin.ValueBool(), model.IsBot.ValueBool())
	}
	if got := dataSourceSetElements(t, model.Teams); !reflect.DeepEqual(got, []string{teamFQN}) {
		t.Fatalf("Teams = %v", got)
	}
	if got := dataSourceSetElements(t, model.Roles); !reflect.DeepEqual(got, []string{roleFQN}) {
		t.Fatalf("Roles = %v", got)
	}
	if model.DefaultPersona.ValueString() != personaFQN {
		t.Fatalf("DefaultPersona = %q", model.DefaultPersona.ValueString())
	}
	if model.Timezone.ValueString() != timezone {
		t.Fatalf("Timezone = %q", model.Timezone.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model userModel
	apply(&ometa.User{Id: id, Name: "bare", Email: "bare@example.com"}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.IsAdmin.ValueBool() || model.IsBot.ValueBool() {
		t.Fatalf("IsAdmin = %v, IsBot = %v, want false", model.IsAdmin.ValueBool(), model.IsBot.ValueBool())
	}
	if !model.Teams.IsNull() && len(model.Teams.Elements()) != 0 {
		t.Fatalf("Teams = %v, want an empty list", model.Teams)
	}
	if !model.DefaultPersona.IsNull() {
		t.Fatalf("DefaultPersona = %v, want null", model.DefaultPersona)
	}
	if !model.Timezone.IsNull() {
		t.Fatalf("Timezone = %v, want null", model.Timezone)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func dataSourceSetElements(t *testing.T, value types.Set) []string {
	t.Helper()
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var out []string
	if diags := value.ElementsAs(context.Background(), &out, false); diags.HasError() {
		t.Fatalf("ElementsAs: %v", diags.Errors())
	}
	return out
}
