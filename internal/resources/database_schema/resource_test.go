package database_schema

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_database_schema" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_database_schema")
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

	database, ok := resp.Schema.Attributes["database"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("database attribute is %T, want StringAttribute", resp.Schema.Attributes["database"])
	}
	if !database.Required {
		t.Fatal("database must be required")
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"version",
		"updated_at",
		"updated_by",
	} {
		if attribute == "version" {
			floatAttribute, floatOK := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !floatOK || !floatAttribute.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed || description.Default == nil {
		t.Fatal("description must be optional, computed, and default to an empty string")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	displayName := "Public schema"
	fqn := "local_postgres.sales.public"
	databaseFQN := "local_postgres.sales"
	databaseName := "sales"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model databaseSchemaModel
	apply(&ometa.DatabaseSchema{
		Id:                 id,
		Name:               "public",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        stringPointer("Default schema"),
		Database: ometa.EntityReference{
			Name:               &databaseName,
			FullyQualifiedName: &databaseFQN,
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "public" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Database.ValueString() != databaseFQN {
		t.Fatalf("Database = %q", model.Database.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Default schema" {
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
}

func TestApplyFallsBackToDatabaseName(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	databaseName := "sales"

	var model databaseSchemaModel
	apply(&ometa.DatabaseSchema{
		Id:       id,
		Name:     "public",
		Database: ometa.EntityReference{Name: &databaseName},
	}, &model)

	if model.Database.ValueString() != databaseName {
		t.Fatalf("Database = %q, want %q", model.Database.ValueString(), databaseName)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestApplyKeepsDatabaseWhenReferenceMissing(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := databaseSchemaModel{Database: types.StringValue("local_postgres.sales")}
	apply(&ometa.DatabaseSchema{Id: id, Name: "public"}, &model)

	if model.Database.ValueString() != "local_postgres.sales" {
		t.Fatalf("Database = %q, want the prior value", model.Database.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := databaseSchemaModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      databaseSchemaModel
		state     databaseSchemaModel
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
			plan: databaseSchemaModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: databaseSchemaModel{
				DisplayName: types.StringNull(),
				Description: base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: databaseSchemaModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
			if len(ops) != tt.wantOps {
				t.Fatalf("len(ops) = %d, want %d: %#v", len(ops), tt.wantOps, ops)
			}
			if tt.wantOps == 0 {
				return
			}
			if ops[0].Path != tt.wantPath {
				t.Fatalf("Path = %q, want %q", ops[0].Path, tt.wantPath)
			}
			if ops[0].Op != tt.wantOp {
				t.Fatalf("Op = %q, want %q", ops[0].Op, tt.wantOp)
			}
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("Value = %v, want %v", ops[0].Value, tt.wantValue)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
