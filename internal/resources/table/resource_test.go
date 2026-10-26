package table

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

	if resp.TypeName != "openmetadata_table" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_table")
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

	databaseSchema, ok := resp.Schema.Attributes["database_schema"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("database_schema attribute is %T, want StringAttribute", resp.Schema.Attributes["database_schema"])
	}
	if !databaseSchema.Required {
		t.Fatal("database_schema must be required")
	}

	columns, ok := resp.Schema.Attributes["columns"].(rschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("columns attribute is %T, want ListNestedAttribute", resp.Schema.Attributes["columns"])
	}
	if !columns.Required {
		t.Fatal("columns must be required")
	}

	for _, attribute := range []string{"name", "data_type", "description", "display_name", "ordinal_position", "precision", "scale"} {
		if _, ok := columns.NestedObject.Attributes[attribute]; !ok {
			t.Fatalf("columns is missing nested attribute %q", attribute)
		}
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
	fqn := "svc.db.public.users"
	schemaFQN := "svc.db.public"
	displayName := "Users"
	description := "One row per user."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	ordinal := int32(2)
	precision := int32(255)
	columnDescription := "Login address."

	model := tableModel{DatabaseSchema: types.StringValue("previous.schema")}
	apply(&ometa.Table{
		Id:                 id,
		Name:               "users",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
		DatabaseSchema:     &ometa.EntityReference{FullyQualifiedName: &schemaFQN},
		Columns: []ometa.Column{
			{Name: "id", DataType: ometa.ColumnDataTypeBIGINT},
			{
				Name:            "email",
				DataType:        ometa.ColumnDataTypeVARCHAR,
				Description:     &columnDescription,
				OrdinalPosition: &ordinal,
				Precision:       &precision,
			},
		},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "users" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DatabaseSchema.ValueString() != schemaFQN {
		t.Fatalf("DatabaseSchema = %q", model.DatabaseSchema.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
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

	var columns []tableColumnModel
	if diags := model.Columns.ElementsAs(context.Background(), &columns, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	if len(columns) != 2 {
		t.Fatalf("len(columns) = %d, want 2", len(columns))
	}
	if columns[0].Name.ValueString() != "id" || columns[0].DataType.ValueString() != "BIGINT" {
		t.Fatalf("first column = %#v", columns[0])
	}
	if columns[0].OrdinalPosition.ValueInt64() != 1 {
		t.Fatalf("first column ordinal = %d, want 1", columns[0].OrdinalPosition.ValueInt64())
	}
	if columns[1].OrdinalPosition.ValueInt64() != 2 || columns[1].Precision.ValueInt64() != 255 {
		t.Fatalf("second column = %#v", columns[1])
	}
	if columns[1].Description.ValueString() != columnDescription {
		t.Fatalf("second column description = %q", columns[1].Description.ValueString())
	}
}

func TestApplyKeepsDatabaseSchemaWhenOmitted(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := tableModel{DatabaseSchema: types.StringValue("svc.db.public")}
	empty := ""
	apply(&ometa.Table{Id: id, Name: "bare", Description: &empty}, &model)

	if model.DatabaseSchema.ValueString() != "svc.db.public" {
		t.Fatalf("DatabaseSchema = %q, want the prior selector", model.DatabaseSchema.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
	if !model.Columns.IsNull() && len(model.Columns.Elements()) != 0 {
		t.Fatalf("Columns = %v, want an empty list", model.Columns)
	}
}

func TestColumnsToAPI(t *testing.T) {
	columnDescription := "Login address."
	columns := types.ListValueMust(tableColumnObjectType, []attr.Value{
		types.ObjectValueMust(tableColumnObjectType.AttrTypes, map[string]attr.Value{
			"name":             types.StringValue("id"),
			"data_type":        types.StringValue("BIGINT"),
			"description":      types.StringNull(),
			"display_name":     types.StringNull(),
			"ordinal_position": types.Int64Value(1),
			"precision":        types.Int64Null(),
			"scale":            types.Int64Null(),
		}),
		types.ObjectValueMust(tableColumnObjectType.AttrTypes, map[string]attr.Value{
			"name":             types.StringValue("email"),
			"data_type":        types.StringValue("VARCHAR"),
			"description":      types.StringValue(columnDescription),
			"display_name":     types.StringValue("Email"),
			"ordinal_position": types.Int64Value(2),
			"precision":        types.Int64Value(255),
			"scale":            types.Int64Null(),
		}),
	})

	converted, diags := columnsToAPI(context.Background(), columns)
	if diags.HasError() {
		t.Fatalf("columnsToAPI() diagnostics = %v", diags.Errors())
	}
	if len(converted) != 2 {
		t.Fatalf("len(converted) = %d, want 2", len(converted))
	}
	if converted[0].Name != "id" || converted[0].DataType != ometa.ColumnDataTypeBIGINT {
		t.Fatalf("first column = %#v", converted[0])
	}
	if converted[0].Description != nil || converted[0].Precision != nil {
		t.Fatalf("first column optional fields = %#v", converted[0])
	}
	if converted[1].DataType != ometa.ColumnDataTypeVARCHAR {
		t.Fatalf("second column type = %q", converted[1].DataType)
	}
	if converted[1].Precision == nil || *converted[1].Precision != 255 {
		t.Fatalf("second column precision = %#v", converted[1].Precision)
	}
	if converted[1].Description == nil || *converted[1].Description != columnDescription {
		t.Fatalf("second column description = %#v", converted[1].Description)
	}
	if converted[1].Scale != nil {
		t.Fatalf("second column scale = %#v, want nil", converted[1].Scale)
	}
}

func TestPatchOps(t *testing.T) {
	columns := types.ListValueMust(tableColumnObjectType, []attr.Value{
		types.ObjectValueMust(tableColumnObjectType.AttrTypes, map[string]attr.Value{
			"name":             types.StringValue("id"),
			"data_type":        types.StringValue("BIGINT"),
			"description":      types.StringNull(),
			"display_name":     types.StringNull(),
			"ordinal_position": types.Int64Value(1),
			"precision":        types.Int64Null(),
			"scale":            types.Int64Null(),
		}),
	})
	extraColumns := types.ListValueMust(tableColumnObjectType, []attr.Value{
		columns.Elements()[0],
		types.ObjectValueMust(tableColumnObjectType.AttrTypes, map[string]attr.Value{
			"name":             types.StringValue("email"),
			"data_type":        types.StringValue("VARCHAR"),
			"description":      types.StringNull(),
			"display_name":     types.StringNull(),
			"ordinal_position": types.Int64Value(2),
			"precision":        types.Int64Value(255),
			"scale":            types.Int64Null(),
		}),
	})

	base := tableModel{
		Name:           types.StringValue("users"),
		DatabaseSchema: types.StringValue("svc.db.public"),
		DisplayName:    types.StringValue("Users"),
		Description:    types.StringValue("before"),
		Columns:        columns,
	}

	tests := []struct {
		name      string
		plan      tableModel
		state     tableModel
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
			plan: tableModel{
				Name:           base.Name,
				DatabaseSchema: base.DatabaseSchema,
				DisplayName:    types.StringValue("New display"),
				Description:    base.Description,
				Columns:        base.Columns,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: tableModel{
				Name:           base.Name,
				DatabaseSchema: base.DatabaseSchema,
				DisplayName:    types.StringNull(),
				Description:    base.Description,
				Columns:        base.Columns,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: tableModel{
				Name:           base.Name,
				DatabaseSchema: base.DatabaseSchema,
				DisplayName:    base.DisplayName,
				Description:    types.StringValue("after"),
				Columns:        base.Columns,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "columns changed",
			plan: tableModel{
				Name:           base.Name,
				DatabaseSchema: base.DatabaseSchema,
				DisplayName:    base.DisplayName,
				Description:    base.Description,
				Columns:        extraColumns,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/columns",
			wantOp:   "add",
		},
		{
			name: "name and database schema are never patched",
			plan: tableModel{
				Name:           types.StringValue("renamed"),
				DatabaseSchema: types.StringValue("svc.db.other"),
				DisplayName:    base.DisplayName,
				Description:    base.Description,
				Columns:        base.Columns,
			},
			state:   base,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, diags := patchOps(context.Background(), tt.plan, tt.state)
			if diags.HasError() {
				t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
			}
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
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
			if tt.wantPath == "/columns" {
				columns, ok := ops[0].Value.([]ometa.Column)
				if !ok {
					t.Fatalf("columns value is %T, want []ometa.Column", ops[0].Value)
				}
				if len(columns) != 2 || columns[1].Name != "email" {
					t.Fatalf("columns value = %#v", columns)
				}
			}
		})
	}
}
