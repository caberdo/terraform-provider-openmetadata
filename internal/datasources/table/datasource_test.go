package table

import (
	"context"
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

	if resp.TypeName != "openmetadata_table" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_table")
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
		"database_schema",
		"display_name",
		"description",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	columns, ok := resp.Schema.Attributes["columns"].(dschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("columns attribute is %T, want ListNestedAttribute", resp.Schema.Attributes["columns"])
	}
	if !columns.Computed {
		t.Fatal("columns must be computed")
	}

	for _, attribute := range []string{"name", "data_type", "data_length", "description", "display_name", "ordinal_position", "precision", "scale"} {
		if _, ok := columns.NestedObject.Attributes[attribute]; !ok {
			t.Fatalf("columns is missing nested attribute %q", attribute)
		}
	}
}

// tableColumnModel is the shape ElementsAs decodes a nested column into.
type tableColumnModel struct {
	Name            types.String `tfsdk:"name"`
	DataType        types.String `tfsdk:"data_type"`
	DataLength      types.Int64  `tfsdk:"data_length"`
	Description     types.String `tfsdk:"description"`
	DisplayName     types.String `tfsdk:"display_name"`
	OrdinalPosition types.Int64  `tfsdk:"ordinal_position"`
	Precision       types.Int64  `tfsdk:"precision"`
	Scale           types.Int64  `tfsdk:"scale"`
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "svc.db.public.users"
	schemaFQN := "svc.db.public"
	displayName := "Users"
	ordinal := int32(2)
	precision := int32(255)
	dataLength := int32(255)
	description := "One row per user."

	var model tableModel
	apply(&ometa.Table{
		Id:                 id,
		Name:               "users",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		DatabaseSchema:     &ometa.EntityReference{FullyQualifiedName: &schemaFQN},
		Columns: []ometa.Column{
			{Name: "id", DataType: ometa.ColumnDataTypeBIGINT},
			{
				Name:            "email",
				DataType:        ometa.ColumnDataTypeVARCHAR,
				OrdinalPosition: &ordinal,
				Precision:       &precision,
				DataLength:      &dataLength,
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
	if columns[1].OrdinalPosition.ValueInt64() != 2 || columns[1].Precision.ValueInt64() != 255 ||
		columns[1].DataLength.ValueInt64() != 255 {
		t.Fatalf("second column = %#v", columns[1])
	}
}

func TestApplyKeepsMissingFieldsNull(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model tableModel
	empty := ""
	apply(&ometa.Table{Id: id, Name: "bare", Description: &empty}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.DatabaseSchema.IsNull() {
		t.Fatalf("DatabaseSchema = %v, want null", model.DatabaseSchema)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
	if model.Columns.IsNull() || len(model.Columns.Elements()) != 0 {
		t.Fatalf("Columns = %v, want an empty list", model.Columns)
	}
}
