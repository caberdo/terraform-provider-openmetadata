package team

import (
	"context"
	"reflect"
	"sort"
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

	if resp.TypeName != "openmetadata_team" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_team")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !name.Required {
		t.Fatal("name must be required")
	}
	if len(name.PlanModifiers) == 0 {
		t.Fatal("name must force replacement")
	}

	teamType := resp.Schema.Attributes["team_type"].(rschema.StringAttribute)
	if !teamType.Required {
		t.Fatal("team_type must be required")
	}

	email := resp.Schema.Attributes["email"].(rschema.StringAttribute)
	if !email.Optional || email.Computed {
		t.Fatal("email must be optional and not computed")
	}

	for _, attribute := range []string{"is_joinable", "default_persona"} {
		computed, ok := resp.Schema.Attributes[attribute]
		if !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
		switch value := computed.(type) {
		case rschema.BoolAttribute:
			if !value.Optional || !value.Computed {
				t.Fatalf("%s must be optional and computed", attribute)
			}
		case rschema.StringAttribute:
			if !value.Optional || !value.Computed {
				t.Fatalf("%s must be optional and computed", attribute)
			}
		}
	}

	for _, attribute := range []string{"parents", "policies", "default_roles", "users"} {
		set, ok := resp.Schema.Attributes[attribute].(rschema.SetAttribute)
		if !ok {
			t.Fatalf("%s is %T, want SetAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !set.Optional || !set.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if set.ElementType != types.StringType {
			t.Fatalf("%s element type = %v, want string", attribute, set.ElementType)
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "version", "updated_at", "updated_by", "description"} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	parentID := uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	policyID := uuid.MustParse("4f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	roleID := uuid.MustParse("5f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	personaID := uuid.MustParse("6f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	userID := uuid.MustParse("3f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "department.group"
	parentFQN := "department"
	policyName := "team-policy"
	roleFQN := "team-role"
	personaFQN := "team-persona"
	userFQN := "jane.doe"
	teamType := ometa.TeamTeamTypeGroup
	email := "group@example.com"
	displayName := "Platform group"
	description := "Group of platform engineers."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	joinable := false

	var model teamModel
	apply(&ometa.Team{
		Id:                 id,
		Name:               "group",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		TeamType:           &teamType,
		Email:              &email,
		IsJoinable:         &joinable,
		Parents:            &[]ometa.EntityReference{{Id: parentID, Type: "team", FullyQualifiedName: &parentFQN}},
		Policies:           &[]ometa.EntityReference{{Id: policyID, Type: "policy", Name: &policyName}},
		DefaultRoles:       &[]ometa.EntityReference{{Id: roleID, Type: "role", FullyQualifiedName: &roleFQN}},
		DefaultPersona:     &ometa.EntityReference{Id: personaID, Type: "persona", FullyQualifiedName: &personaFQN},
		Users:              &[]ometa.EntityReference{{Id: userID, Type: "user", FullyQualifiedName: &userFQN}},
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "group" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.TeamType.ValueString() != "Group" {
		t.Fatalf("TeamType = %q", model.TeamType.ValueString())
	}
	if model.Email.ValueString() != email {
		t.Fatalf("Email = %q", model.Email.ValueString())
	}
	if model.IsJoinable.ValueBool() {
		t.Fatal("IsJoinable = true, want false")
	}
	if got := setValues(t, model.Parents); !reflect.DeepEqual(got, []string{parentFQN}) {
		t.Fatalf("Parents = %v", got)
	}
	if got := setValues(t, model.Policies); !reflect.DeepEqual(got, []string{policyName}) {
		t.Fatalf("Policies = %v", got)
	}
	if got := setValues(t, model.DefaultRoles); !reflect.DeepEqual(got, []string{roleFQN}) {
		t.Fatalf("DefaultRoles = %v", got)
	}
	if model.DefaultPersona.ValueString() != personaFQN {
		t.Fatalf("DefaultPersona = %q", model.DefaultPersona.ValueString())
	}
	if got := setValues(t, model.Users); !reflect.DeepEqual(got, []string{userFQN}) {
		t.Fatalf("Users = %v", got)
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

	var model teamModel
	empty := ""
	apply(&ometa.Team{Id: id, Name: "bare", Description: &empty}, &model)

	if !model.Email.IsNull() {
		t.Fatalf("Email = %v, want null", model.Email)
	}
	if !model.IsJoinable.ValueBool() {
		t.Fatal("IsJoinable = false, want the server default true")
	}
	if model.Parents.IsNull() || len(model.Parents.Elements()) != 0 {
		t.Fatalf("Parents = %v, want an empty set", model.Parents)
	}
	for name, set := range map[string]types.Set{
		"Policies":     model.Policies,
		"DefaultRoles": model.DefaultRoles,
		"Users":        model.Users,
	} {
		if set.IsNull() || len(set.Elements()) != 0 {
			t.Fatalf("%s = %v, want an empty set", name, set)
		}
	}
	if !model.DefaultPersona.IsNull() {
		t.Fatalf("DefaultPersona = %v, want null", model.DefaultPersona)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	lookup := func(_ context.Context, kind string, names []string) ([]ometa.EntityReference, error) {
		references := make([]ometa.EntityReference, 0, len(names))
		for _, name := range names {
			references = append(references, ometa.EntityReference{Id: uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)), Type: kind})
		}
		return references, nil
	}

	base := teamModel{
		Name:           types.StringValue("group"),
		TeamType:       types.StringValue("Group"),
		Email:          types.StringValue("group@example.com"),
		IsJoinable:     types.BoolValue(true),
		DisplayName:    types.StringValue("Group"),
		Description:    types.StringValue("before"),
		Parents:        stringSetValue("department"),
		Policies:       stringSetValue("team-policy"),
		DefaultRoles:   stringSetValue("team-role"),
		DefaultPersona: types.StringValue("team-persona"),
		Users:          stringSetValue("jane.doe"),
	}

	tests := []struct {
		name       string
		plan       teamModel
		wantOps    map[string]string
		wantValues map[string]any
	}{
		{
			name: "no changes",
			plan: base,
		},
		{
			name: "scalar fields changed",
			plan: withTeam(base, func(model *teamModel) {
				model.Description = types.StringValue("after")
				model.TeamType = types.StringValue("Department")
				model.IsJoinable = types.BoolValue(false)
			}),
			wantOps: map[string]string{
				"/teamType":    "add",
				"/isJoinable":  "add",
				"/description": "add",
			},
			wantValues: map[string]any{
				"/teamType":    "Department",
				"/isJoinable":  false,
				"/description": "after",
			},
		},
		{
			name: "email removed",
			plan: withTeam(base, func(model *teamModel) {
				model.Email = types.StringNull()
			}),
			wantOps: map[string]string{"/email": "remove"},
		},
		{
			name: "display name removed",
			plan: withTeam(base, func(model *teamModel) {
				model.DisplayName = types.StringNull()
			}),
			wantOps: map[string]string{"/displayName": "remove"},
		},
		{
			name: "parents replace the reference set",
			plan: withTeam(base, func(model *teamModel) {
				model.Parents = stringSetValue("division")
			}),
			wantOps: map[string]string{"/parents": "add"},
			wantValues: map[string]any{
				"/parents": []ometa.EntityReference{{
					Id:   uuid.NewSHA1(uuid.NameSpaceOID, []byte("division")),
					Type: "team",
				}},
			},
		},
		{
			name: "users replace the reference set",
			plan: withTeam(base, func(model *teamModel) {
				model.Users = stringSetValue("john.doe")
			}),
			wantOps: map[string]string{"/users": "add"},
			wantValues: map[string]any{
				"/users": []ometa.EntityReference{{
					Id:   uuid.NewSHA1(uuid.NameSpaceOID, []byte("john.doe")),
					Type: "user",
				}},
			},
		},
		{
			name: "persona removed",
			plan: withTeam(base, func(model *teamModel) {
				model.DefaultPersona = types.StringNull()
			}),
			wantOps: map[string]string{"/defaultPersona": "remove"},
		},
		{
			name: "persona changed",
			plan: withTeam(base, func(model *teamModel) {
				model.DefaultPersona = types.StringValue("other-persona")
			}),
			wantOps: map[string]string{"/defaultPersona": "add"},
			wantValues: map[string]any{
				"/defaultPersona": ometa.EntityReference{
					Id:   uuid.NewSHA1(uuid.NameSpaceOID, []byte("other-persona")),
					Type: "persona",
				},
			},
		},
		{
			name: "name is never patched",
			plan: withTeam(base, func(model *teamModel) {
				model.Name = types.StringValue("renamed")
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := patchOps(context.Background(), tt.plan, base, lookup)
			if err != nil {
				t.Fatalf("patchOps() error = %v", err)
			}

			byPath := make(map[string]ometa.JSONPatchOp, len(ops))
			for _, op := range ops {
				byPath[op.Path] = op
			}

			if len(byPath) != len(tt.wantOps) {
				t.Fatalf("paths = %v, want %v", keys(byPath), keys(tt.wantOps))
			}
			for path, wantOp := range tt.wantOps {
				op, ok := byPath[path]
				if !ok {
					t.Fatalf("missing operation for %s: %#v", path, ops)
				}
				if op.Op != wantOp {
					t.Fatalf("%s op = %q, want %q", path, op.Op, wantOp)
				}
				if want, ok := tt.wantValues[path]; ok && !reflect.DeepEqual(op.Value, want) {
					t.Fatalf("%s value = %#v, want %#v", path, op.Value, want)
				}
			}
			if _, ok := byPath["/name"]; ok {
				t.Fatal("name must never be patched")
			}
		})
	}
}

func TestPatchOpsSkipsUnknownSets(t *testing.T) {
	base := teamModel{
		TeamType:       types.StringValue("Group"),
		Email:          types.StringValue("group@example.com"),
		IsJoinable:     types.BoolValue(true),
		Description:    types.StringValue("before"),
		Parents:        stringSetValue("department"),
		Policies:       stringSetValue("team-policy"),
		DefaultRoles:   stringSetValue("team-role"),
		DefaultPersona: types.StringValue("team-persona"),
		Users:          stringSetValue("jane.doe"),
	}
	plan := withTeam(base, func(model *teamModel) {
		model.Parents = types.SetUnknown(types.StringType)
		model.Users = types.SetUnknown(types.StringType)
		model.DefaultPersona = types.StringUnknown()
	})

	ops, err := patchOps(context.Background(), plan, base, func(context.Context, string, []string) ([]ometa.EntityReference, error) {
		t.Fatal("lookup must not be called for unknown values")
		return nil, nil
	})
	if err != nil {
		t.Fatalf("patchOps() error = %v", err)
	}
	if len(ops) != 0 {
		t.Fatalf("ops = %#v, want none", ops)
	}
}

// stringSetValue builds a Terraform string set from the given values.
func stringSetValue(values ...string) types.Set {
	elements := make([]attr.Value, len(values))
	for index, value := range values {
		elements[index] = types.StringValue(value)
	}
	return types.SetValueMust(types.StringType, elements)
}

// setValues reads a Terraform string set back into a sorted Go slice. Sets are
// unordered, so assertions sort before comparing.
func setValues(t *testing.T, value types.Set) []string {
	t.Helper()
	var out []string
	if diags := value.ElementsAs(context.Background(), &out, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	sort.Strings(out)
	return out
}

// withTeam copies a model and applies a mutation, so table cases stay readable.
func withTeam(base teamModel, mutate func(*teamModel)) teamModel {
	copied := base
	mutate(&copied)
	return copied
}

// keys returns the paths of a patch operation map.
func keys[T any](ops map[string]T) []string {
	paths := make([]string, 0, len(ops))
	for path := range ops {
		paths = append(paths, path)
	}
	return paths
}
