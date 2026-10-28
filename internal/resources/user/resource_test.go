package user

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_user" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_user")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatalf("name attribute = %T, want a required StringAttribute", resp.Schema.Attributes["name"])
	}

	email, ok := resp.Schema.Attributes["email"].(rschema.StringAttribute)
	if !ok || !email.Required {
		t.Fatalf("email attribute = %T, want a required StringAttribute", resp.Schema.Attributes["email"])
	}

	isBot, ok := resp.Schema.Attributes["is_bot"].(rschema.BoolAttribute)
	if !ok || !isBot.Optional || !isBot.Computed {
		t.Fatalf("is_bot attribute = %T, want Optional+Computed", resp.Schema.Attributes["is_bot"])
	}

	teams, ok := resp.Schema.Attributes["teams"].(rschema.SetAttribute)
	if !ok || !teams.Optional || !teams.Computed {
		t.Fatalf("teams attribute = %T, want an Optional+Computed SetAttribute", resp.Schema.Attributes["teams"])
	}

	jwt, ok := resp.Schema.Attributes["jwt_token_expiry"].(rschema.StringAttribute)
	if !ok || !jwt.Optional || !jwt.Computed {
		t.Fatalf("jwt_token_expiry attribute = %T, want Optional+Computed", resp.Schema.Attributes["jwt_token_expiry"])
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
		value, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !value.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
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
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
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
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "jane" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Engineer" {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Email.ValueString() != "jane@example.com" {
		t.Fatalf("Email = %q", model.Email.ValueString())
	}
	if !model.IsAdmin.ValueBool() {
		t.Fatal("IsAdmin = false")
	}
	if model.IsBot.ValueBool() {
		t.Fatal("IsBot = true")
	}
	if got := setElements(t, model.Teams); !reflect.DeepEqual(got, []string{teamFQN}) {
		t.Fatalf("Teams = %v", got)
	}
	if got := setElements(t, model.Roles); !reflect.DeepEqual(got, []string{roleFQN}) {
		t.Fatalf("Roles = %v", got)
	}
	if model.DefaultPersona.ValueString() != personaFQN {
		t.Fatalf("DefaultPersona = %q", model.DefaultPersona.ValueString())
	}
	if model.Timezone.ValueString() != timezone {
		t.Fatalf("Timezone = %q", model.Timezone.ValueString())
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
	if model.UpdatedBy.ValueString() != updatedBy {
		t.Fatalf("UpdatedBy = %q", model.UpdatedBy.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := userModel{JWTTokenExpiry: types.StringValue("30")}
	apply(&ometa.User{Id: id, Name: "bare", Email: "bare@example.com"}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.IsAdmin.ValueBool() {
		t.Fatal("IsAdmin = true, want the false default")
	}
	if model.IsBot.ValueBool() {
		t.Fatal("IsBot = true, want the false default")
	}
	if !model.Teams.IsNull() && len(model.Teams.Elements()) != 0 {
		t.Fatalf("Teams = %v, want an empty list", model.Teams)
	}
	if !model.Roles.IsNull() && len(model.Roles.Elements()) != 0 {
		t.Fatalf("Roles = %v, want an empty list", model.Roles)
	}
	if !model.DefaultPersona.IsNull() {
		t.Fatalf("DefaultPersona = %v, want null", model.DefaultPersona)
	}
	if !model.Timezone.IsNull() {
		t.Fatalf("Timezone = %v, want null", model.Timezone)
	}
	if model.JWTTokenExpiry.ValueString() != "30" {
		t.Fatalf("JWTTokenExpiry = %q, apply must not touch it", model.JWTTokenExpiry.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func setElements(t *testing.T, value types.Set) []string {
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

func stringSet(values ...string) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, elements)
}

func stubResolver(mapping map[string]uuid.UUID) referenceResolver {
	return func(_ context.Context, kind, name string) (uuid.UUID, error) {
		id, ok := mapping[kind+":"+name]
		if !ok {
			return uuid.Nil, fmt.Errorf("unexpected %s %q", kind, name)
		}
		return id, nil
	}
}

func TestPatchOpsNoChanges(t *testing.T) {
	base := userModel{
		Name:               types.StringValue("jane"),
		Email:              types.StringValue("jane@example.com"),
		DisplayName:        types.StringValue("Jane"),
		Description:        types.StringValue("before"),
		IsAdmin:            types.BoolValue(false),
		IsBot:              types.BoolValue(false),
		JWTTokenExpiry:     types.StringNull(),
		Teams:              stringSet("engineering"),
		Roles:              stringSet("data-steward"),
		DefaultPersona:     types.StringNull(),
		Timezone:           types.StringNull(),
		FullyQualifiedName: types.StringValue("jane"),
	}

	ops, diags := patchOps(context.Background(), base, base, stubResolver(nil))
	if diags.HasError() {
		t.Fatalf("patchOps diagnostics = %v", diags.Errors())
	}
	if len(ops) != 0 {
		t.Fatalf("len(ops) = %d, want 0: %#v", len(ops), ops)
	}
}

func TestPatchOpsChangedFields(t *testing.T) {
	teamID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	roleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	personaID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	state := userModel{
		Name:           types.StringValue("jane"),
		Email:          types.StringValue("jane@example.com"),
		DisplayName:    types.StringValue("Jane"),
		Description:    types.StringValue("before"),
		IsAdmin:        types.BoolValue(false),
		IsBot:          types.BoolValue(false),
		Teams:          stringSet("team-a"),
		Roles:          stringSet("role-a"),
		DefaultPersona: types.StringNull(),
		Timezone:       types.StringNull(),
	}
	plan := userModel{
		Name:           types.StringValue("jane"),
		Email:          types.StringValue("jane@example.com"),
		DisplayName:    types.StringValue("Jane Doe"),
		Description:    types.StringValue("after"),
		IsAdmin:        types.BoolValue(true),
		IsBot:          types.BoolValue(false),
		Teams:          stringSet("team-b"),
		Roles:          stringSet("role-b"),
		DefaultPersona: types.StringValue("persona-a"),
		Timezone:       types.StringValue("Europe/Amsterdam"),
	}

	ops, diags := patchOps(context.Background(), plan, state, stubResolver(map[string]uuid.UUID{
		"team:team-b":       teamID,
		"role:role-b":       roleID,
		"persona:persona-a": personaID,
	}))
	if diags.HasError() {
		t.Fatalf("patchOps diagnostics = %v", diags.Errors())
	}

	want := []ometa.JSONPatchOp{
		{Op: "add", Path: "/displayName", Value: "Jane Doe"},
		{Op: "add", Path: "/description", Value: "after"},
		{Op: "add", Path: "/isAdmin", Value: true},
		{Op: "add", Path: "/timezone", Value: "Europe/Amsterdam"},
		{Op: "add", Path: "/teams", Value: []ometa.EntityReference{{Id: teamID, Type: "team"}}},
		{Op: "add", Path: "/roles", Value: []ometa.EntityReference{{Id: roleID, Type: "role"}}},
		{Op: "add", Path: "/defaultPersona", Value: ometa.EntityReference{Id: personaID, Type: "persona"}},
	}
	if !reflect.DeepEqual(ops, want) {
		t.Fatalf("ops = %#v, want %#v", ops, want)
	}
}

func TestPatchOpsRemovesOptionalFields(t *testing.T) {
	state := userModel{
		DisplayName:    types.StringValue("Jane"),
		Description:    types.StringValue(""),
		IsAdmin:        types.BoolValue(false),
		Teams:          stringSet("team-a"),
		Roles:          stringSet("role-a"),
		DefaultPersona: types.StringValue("persona-a"),
		Timezone:       types.StringValue("UTC"),
	}
	plan := userModel{
		DisplayName:    types.StringNull(),
		Description:    types.StringValue(""),
		IsAdmin:        types.BoolValue(false),
		Teams:          state.Teams,
		Roles:          state.Roles,
		DefaultPersona: types.StringNull(),
		Timezone:       types.StringNull(),
	}

	ops, diags := patchOps(context.Background(), plan, state, stubResolver(nil))
	if diags.HasError() {
		t.Fatalf("patchOps diagnostics = %v", diags.Errors())
	}

	want := []ometa.JSONPatchOp{
		{Op: "remove", Path: "/displayName"},
		{Op: "remove", Path: "/timezone"},
		{Op: "remove", Path: "/defaultPersona"},
	}
	if !reflect.DeepEqual(ops, want) {
		t.Fatalf("ops = %#v, want %#v", ops, want)
	}
}

func TestPatchOpsNeverPatchesImmutableFields(t *testing.T) {
	state := userModel{
		Name:           types.StringValue("jane"),
		Email:          types.StringValue("jane@example.com"),
		IsBot:          types.BoolValue(true),
		JWTTokenExpiry: types.StringValue("7"),
		Teams:          stringSet("team-a"),
		Roles:          stringSet("role-a"),
	}
	plan := userModel{
		Name:           types.StringValue("janet"),
		Email:          types.StringValue("janet@example.com"),
		IsBot:          types.BoolValue(false),
		JWTTokenExpiry: types.StringValue("30"),
		Teams:          state.Teams,
		Roles:          state.Roles,
	}

	ops, diags := patchOps(context.Background(), plan, state, stubResolver(nil))
	if diags.HasError() {
		t.Fatalf("patchOps diagnostics = %v", diags.Errors())
	}
	if len(ops) != 0 {
		t.Fatalf("len(ops) = %d, want 0 (immutable fields must not be patched): %#v", len(ops), ops)
	}
}

func TestPatchOpsEmptyTeamsReplacesList(t *testing.T) {
	state := userModel{
		Teams: stringSet("team-a"),
		Roles: stringSet("role-a"),
	}
	plan := userModel{
		Teams: stringSet(),
		Roles: state.Roles,
	}

	ops, diags := patchOps(context.Background(), plan, state, stubResolver(nil))
	if diags.HasError() {
		t.Fatalf("patchOps diagnostics = %v", diags.Errors())
	}
	if len(ops) != 1 || ops[0].Path != "/teams" || ops[0].Op != "add" {
		t.Fatalf("ops = %#v, want one add on /teams", ops)
	}
}
