package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func policySet(t *testing.T, values ...string) types.Set {
	t.Helper()

	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, elements)
}

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_role" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_role")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want StringAttribute", resp.Schema.Attributes["name"])
	}
	if !name.Required {
		t.Fatal("name must be required")
	}

	policies, ok := resp.Schema.Attributes["policies"].(rschema.SetAttribute)
	if !ok {
		t.Fatalf("policies attribute is %T, want SetAttribute", resp.Schema.Attributes["policies"])
	}
	if !policies.Required {
		t.Fatal("policies must be required")
	}
	if len(policies.Validators) == 0 {
		t.Fatal("policies must require at least one entry")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Required {
		t.Fatal("display_name must be optional")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed {
		t.Fatal("description must be optional and computed")
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
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
	displayName := "Data Steward"
	fqn := "DataSteward"
	policyName := "DataStewardPolicy"
	policyFQN := "DataStewardPolicy"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	description := "Can curate data assets."

	var model roleModel
	apply(&ometa.Role{
		Id:                 id,
		Name:               "DataSteward",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Policies: &[]ometa.EntityReference{{
			Id:                 uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"),
			Type:               "policy",
			Name:               &policyName,
			FullyQualifiedName: &policyFQN,
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "DataSteward" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Can curate data assets." {
		t.Fatalf("Description = %q", model.Description.ValueString())
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

	var policies []string
	model.Policies.ElementsAs(context.Background(), &policies, false)
	if len(policies) != 1 || policies[0] != policyFQN {
		t.Fatalf("Policies = %#v, want [%q]", policies, policyFQN)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	policyName := "OnlyPolicy"

	model := roleModel{
		Policies: policySet(t, "OnlyPolicy"),
	}
	apply(&ometa.Role{Id: id, Name: "bare"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}

	// A server response without the relationship must not zero the required
	// selector.
	var policies []string
	model.Policies.ElementsAs(context.Background(), &policies, false)
	if len(policies) != 1 || policies[0] != policyName {
		t.Fatalf("Policies = %#v, want the previous value [%q]", policies, policyName)
	}
}

func TestApplyEmptyPolicySet(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model roleModel
	apply(&ometa.Role{Id: id, Name: "bare", Policies: &[]ometa.EntityReference{}}, &model)

	if model.Policies.IsNull() {
		t.Fatal("Policies = null, want an empty list")
	}
	if len(model.Policies.Elements()) != 0 {
		t.Fatalf("Policies = %v, want an empty list", model.Policies)
	}
}

func TestPatchOps(t *testing.T) {
	base := roleModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Policies:    policySet(t, "FirstPolicy"),
	}
	reference := ometa.EntityReference{
		Id:   uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"),
		Type: "policy",
	}

	tests := []struct {
		name      string
		plan      roleModel
		state     roleModel
		policies  []ometa.EntityReference
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "name is never patched",
			plan: roleModel{
				Name:        types.StringValue("Renamed"),
				DisplayName: base.DisplayName,
				Description: base.Description,
				Policies:    base.Policies,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: roleModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Policies:    base.Policies,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: roleModel{
				DisplayName: types.StringNull(),
				Description: base.Description,
				Policies:    base.Policies,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: roleModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Policies:    base.Policies,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "policies replaced",
			plan: roleModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Policies:    policySet(t, "SecondPolicy"),
			},
			state:     base,
			policies:  []ometa.EntityReference{reference},
			wantOps:   1,
			wantPath:  "/policies",
			wantOp:    "add",
			wantValue: []ometa.EntityReference{reference},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state, tt.policies)
			if len(ops) != tt.wantOps {
				t.Fatalf("len(ops) = %d, want %d: %#v", len(ops), tt.wantOps, ops)
			}
			if tt.wantOps == 0 {
				return
			}
			if ops[0].Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", ops[0].Path, tt.wantPath)
			}
			if ops[0].Op != tt.wantOp {
				t.Fatalf("op = %q, want %q", ops[0].Op, tt.wantOp)
			}
			if tt.wantValue != nil {
				if refs, ok := tt.wantValue.([]ometa.EntityReference); ok {
					got, ok := ops[0].Value.([]ometa.EntityReference)
					if !ok || len(got) != len(refs) || got[0].Id != refs[0].Id || got[0].Type != refs[0].Type {
						t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
					}
					return
				}
				if ops[0].Value != tt.wantValue {
					t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
				}
			}
		})
	}
}
