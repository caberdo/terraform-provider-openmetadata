package persona

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

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_persona" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_persona")
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

	defaultAttribute, ok := resp.Schema.Attributes["default"].(rschema.BoolAttribute)
	if !ok {
		t.Fatalf("default attribute is %T, want BoolAttribute", resp.Schema.Attributes["default"])
	}
	if !defaultAttribute.Optional || !defaultAttribute.Computed {
		t.Fatal("default must be optional and computed")
	}
	if defaultAttribute.Default != nil {
		t.Fatal("default must not carry a schema default")
	}

	users, ok := resp.Schema.Attributes["users"].(rschema.SetAttribute)
	if !ok {
		t.Fatalf("users attribute is %T, want SetAttribute", resp.Schema.Attributes["users"])
	}
	if !users.Optional || !users.Computed {
		t.Fatal("users must be optional and computed")
	}
	if users.ElementType != types.StringType {
		t.Fatalf("users element type = %v, want string", users.ElementType)
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"version",
		"updated_at",
		"updated_by",
	} {
		if attribute == "version" {
			version, ok := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !ok || !version.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	name := "DataSteward"
	fqn := "DataSteward"
	displayName := "Data Steward"
	description := "Owns data quality."
	userFQN := "alice"
	userName := "alice"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	isDefault := true

	var model personaModel
	diags := apply(context.Background(), &ometa.Persona{
		Id:                 id,
		Name:               name,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Default:            &isDefault,
		Users: &[]ometa.EntityReference{
			{Id: uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"), Type: "user", FullyQualifiedName: &userFQN, Name: &userName},
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != name {
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
	if !model.Default.ValueBool() {
		t.Fatal("Default = false")
	}
	users := model.Users.Elements()
	if len(users) != 1 || users[0].(types.String).ValueString() != userFQN {
		t.Fatalf("Users = %v", users)
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

	var model personaModel
	diags := apply(context.Background(), &ometa.Persona{Id: id, Name: "Bare"}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want empty", model.Description.ValueString())
	}
	if model.Default.ValueBool() {
		t.Fatal("Default = true, want the false default")
	}
	if model.Users.IsNull() || len(model.Users.Elements()) != 0 {
		t.Fatalf("Users = %v, want an empty set", model.Users)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func stringSet(t *testing.T, values ...string) types.Set {
	t.Helper()

	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	set, diags := types.SetValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("building set: %v", diags.Errors())
	}
	return set
}

func TestPatchOps(t *testing.T) {
	base := personaModel{
		Name:        types.StringValue("DataSteward"),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Default:     types.BoolValue(false),
		Users:       stringSet(t, "alice"),
	}

	reference := ometa.EntityReference{Id: uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"), Type: "user"}

	tests := []struct {
		name      string
		plan      personaModel
		state     personaModel
		users     []ometa.EntityReference
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
			name: "display name replaced",
			plan: personaModel{
				Name:        base.Name,
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Default:     base.Default,
				Users:       base.Users,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: personaModel{
				Name:        base.Name,
				DisplayName: types.StringNull(),
				Description: base.Description,
				Default:     base.Default,
				Users:       base.Users,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: personaModel{
				Name:        base.Name,
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Default:     base.Default,
				Users:       base.Users,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "default changed",
			plan: personaModel{
				Name:        base.Name,
				DisplayName: base.DisplayName,
				Description: base.Description,
				Default:     types.BoolValue(true),
				Users:       base.Users,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/default",
			wantOp:    "add",
			wantValue: true,
		},
		{
			name: "users changed",
			plan: personaModel{
				Name:        base.Name,
				DisplayName: base.DisplayName,
				Description: base.Description,
				Default:     base.Default,
				Users:       stringSet(t, "bob"),
			},
			state:     base,
			users:     []ometa.EntityReference{reference},
			wantOps:   1,
			wantPath:  "/users",
			wantOp:    "add",
			wantValue: reference,
		},
		{
			name: "name cannot be patched",
			plan: personaModel{
				Name:        types.StringValue("Renamed"),
				DisplayName: base.DisplayName,
				Description: base.Description,
				Default:     base.Default,
				Users:       base.Users,
			},
			state:   base,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state, tt.users)
			if len(ops) != tt.wantOps {
				t.Fatalf("got %d ops, want %d: %+v", len(ops), tt.wantOps, ops)
			}
			if tt.wantOps == 0 {
				return
			}

			op := ops[0]
			if op.Path != tt.wantPath || op.Op != tt.wantOp {
				t.Fatalf("op = %s %s, want %s %s", op.Op, op.Path, tt.wantOp, tt.wantPath)
			}
			if tt.wantValue != nil {
				switch want := tt.wantValue.(type) {
				case ometa.EntityReference:
					refs, ok := op.Value.([]ometa.EntityReference)
					if !ok || len(refs) != 1 || refs[0].Id != want.Id || refs[0].Type != want.Type {
						t.Fatalf("op value = %v, want %v", op.Value, want)
					}
				default:
					if op.Value != want {
						t.Fatalf("op value = %v, want %v", op.Value, want)
					}
				}
			}
		})
	}
}
