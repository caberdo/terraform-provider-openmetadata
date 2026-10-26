package dashboard_data_model

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

	if resp.TypeName != "openmetadata_dashboard_data_model" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_dashboard_data_model")
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

	for _, attribute := range []string{"service", "data_model_type"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}

	sql, ok := resp.Schema.Attributes["sql"].(rschema.StringAttribute)
	if !ok || !sql.Optional || !sql.Computed {
		t.Fatal("sql must be optional and computed")
	}

	columns, ok := resp.Schema.Attributes["columns"].(rschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("columns attribute is %T, want ListNestedAttribute", resp.Schema.Attributes["columns"])
	}
	if !columns.Required {
		t.Fatal("columns must be required")
	}

	for _, attribute := range []string{"name", "data_type"} {
		columnAttribute, ok := columns.NestedObject.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !columnAttribute.Required {
			t.Fatalf("columns.%s must be a required string", attribute)
		}
	}
	for _, attribute := range []string{"display_name", "description"} {
		if _, ok := columns.NestedObject.Attributes[attribute].(rschema.StringAttribute); !ok {
			t.Fatalf("columns.%s must be a string", attribute)
		}
	}
}

func TestColumnsRoundTrip(t *testing.T) {
	ctx := context.Background()
	displayName := "Total Sales"
	description := "Sum of all sales."

	apiColumns := []ometa.Column{
		{
			Name:        "total_sales",
			DataType:    ometa.ColumnDataType("DOUBLE"),
			DisplayName: &displayName,
			Description: &description,
		},
		{
			Name:     "region",
			DataType: ometa.ColumnDataType("VARCHAR"),
		},
	}

	list, diags := columnsToList(ctx, apiColumns)
	if diags.HasError() {
		t.Fatalf("columnsToList diagnostics = %v", diags.Errors())
	}

	roundTripped, diags := columnsFromList(ctx, list)
	if diags.HasError() {
		t.Fatalf("columnsFromList diagnostics = %v", diags.Errors())
	}
	if len(roundTripped) != 2 {
		t.Fatalf("len(columns) = %d, want 2", len(roundTripped))
	}
	if roundTripped[0].Name != "total_sales" || roundTripped[0].DataType != ometa.ColumnDataType("DOUBLE") {
		t.Fatalf("first column = %#v", roundTripped[0])
	}
	if roundTripped[0].DisplayName == nil || *roundTripped[0].DisplayName != displayName {
		t.Fatalf("first column display name = %v", roundTripped[0].DisplayName)
	}
	if roundTripped[0].Description == nil || *roundTripped[0].Description != description {
		t.Fatalf("first column description = %v", roundTripped[0].Description)
	}
	if roundTripped[1].DisplayName != nil {
		t.Fatalf("second column display name = %v, want nil", roundTripped[1].DisplayName)
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "warehouse.model.power_bi_model"
	displayName := "Power BI Model"
	description := "Model behind the finance dashboard."
	sql := "SELECT * FROM sales"
	serviceFQN := "warehouse"
	serviceName := "warehouse"
	columnDescription := "Region name."
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model dashboardDataModelModel
	diags := apply(ctx, &ometa.DashboardDataModel{
		Id:                 id,
		Name:               "power_bi_model",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Sql:                &sql,
		DataModelType:      ometa.DashboardDataModelDataModelType("PowerBIDataModel"),
		Columns: []ometa.Column{
			{
				Name:        "region",
				DataType:    ometa.ColumnDataType("VARCHAR"),
				Description: &columnDescription,
			},
		},
		Service: &ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Name:               &serviceName,
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "power_bi_model" {
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
	if model.SQL.ValueString() != sql {
		t.Fatalf("SQL = %q", model.SQL.ValueString())
	}
	if model.DataModelType.ValueString() != "PowerBIDataModel" {
		t.Fatalf("DataModelType = %q", model.DataModelType.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	columns, diags := columnsFromList(ctx, model.Columns)
	if diags.HasError() {
		t.Fatalf("columnsFromList diagnostics = %v", diags.Errors())
	}
	if len(columns) != 1 || columns[0].Name != "region" {
		t.Fatalf("Columns = %#v", columns)
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

func TestApplyPreservesOmittedFields(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := dashboardDataModelModel{
		Service: types.StringValue("warehouse"),
		SQL:     types.StringValue("SELECT 1"),
	}

	diags := apply(ctx, &ometa.DashboardDataModel{
		Id:            id,
		Name:          "sparse",
		DataModelType: ometa.DashboardDataModelDataModelType("PowerBIDataModel"),
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply diagnostics = %v", diags.Errors())
	}

	if model.Service.ValueString() != "warehouse" {
		t.Fatalf("Service = %q, want the prior value", model.Service.ValueString())
	}
	if model.SQL.ValueString() != "SELECT 1" {
		t.Fatalf("SQL = %q, want the prior value", model.SQL.ValueString())
	}
	if !model.Columns.IsNull() {
		t.Fatalf("Columns = %v, want null", model.Columns)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	ctx := context.Background()

	baseColumns := []ometa.Column{
		{Name: "region", DataType: ometa.ColumnDataType("VARCHAR")},
	}
	baseList, diags := columnsToList(ctx, baseColumns)
	if diags.HasError() {
		t.Fatalf("columnsToList diagnostics = %v", diags.Errors())
	}

	base := dashboardDataModelModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		SQL:         types.StringValue("SELECT 1"),
		Columns:     baseList,
	}

	updatedColumns := []ometa.Column{
		{Name: "region", DataType: ometa.ColumnDataType("VARCHAR")},
		{Name: "sales", DataType: ometa.ColumnDataType("DOUBLE")},
	}
	updatedList, diags := columnsToList(ctx, updatedColumns)
	if diags.HasError() {
		t.Fatalf("columnsToList diagnostics = %v", diags.Errors())
	}

	tests := []struct {
		name      string
		plan      dashboardDataModelModel
		state     dashboardDataModelModel
		columns   []ometa.Column
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			columns: baseColumns,
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: dashboardDataModelModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				SQL:         base.SQL,
				Columns:     base.Columns,
			},
			state:     base,
			columns:   baseColumns,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: dashboardDataModelModel{
				DisplayName: types.StringNull(),
				Description: base.Description,
				SQL:         base.SQL,
				Columns:     base.Columns,
			},
			state:    base,
			columns:  baseColumns,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: dashboardDataModelModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				SQL:         base.SQL,
				Columns:     base.Columns,
			},
			state:     base,
			columns:   baseColumns,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "sql changed",
			plan: dashboardDataModelModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				SQL:         types.StringValue("SELECT 2"),
				Columns:     base.Columns,
			},
			state:     base,
			columns:   baseColumns,
			wantOps:   1,
			wantPath:  "/sql",
			wantOp:    "add",
			wantValue: "SELECT 2",
		},
		{
			name: "columns changed",
			plan: dashboardDataModelModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				SQL:         base.SQL,
				Columns:     updatedList,
			},
			state:     base,
			columns:   updatedColumns,
			wantOps:   1,
			wantPath:  "/columns",
			wantOp:    "add",
			wantValue: updatedColumns,
		},
		{
			name: "requires replace fields are not patched",
			plan: dashboardDataModelModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				SQL:         base.SQL,
				Columns:     base.Columns,
				Service:     types.StringValue("other-service"),
				Name:        types.StringValue("other-name"),
			},
			state:   base,
			columns: baseColumns,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state, tt.columns)
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
				if want, ok := tt.wantValue.([]ometa.Column); ok {
					got, ok := ops[0].Value.([]ometa.Column)
					if !ok || len(got) != len(want) {
						t.Fatalf("value = %#v, want %#v", ops[0].Value, want)
					}
					for i := range want {
						if got[i].Name != want[i].Name || got[i].DataType != want[i].DataType {
							t.Fatalf("value = %#v, want %#v", got, want)
						}
					}
				} else if ops[0].Value != tt.wantValue {
					t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
				}
			}
		})
	}
}
