package dashboard_data_model

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_dashboard_data_model" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_dashboard_data_model")
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
		"service",
		"data_model_type",
		"sql",
		"version",
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
	for _, attribute := range []string{"name", "data_type", "display_name", "description"} {
		if _, ok := columns.NestedObject.Attributes[attribute].(dschema.StringAttribute); !ok {
			t.Fatalf("columns.%s must be a string", attribute)
		}
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
	columnDescription := "Region name."

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
		Service: &ometa.EntityReference{FullyQualifiedName: &serviceFQN},
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply diagnostics = %v", diags.Errors())
	}

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "power_bi_model" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.DataModelType.ValueString() != "PowerBIDataModel" {
		t.Fatalf("DataModelType = %q", model.DataModelType.ValueString())
	}
	if model.SQL.ValueString() != sql {
		t.Fatalf("SQL = %q", model.SQL.ValueString())
	}
	if model.Columns.IsNull() {
		t.Fatal("Columns is null")
	}
	var columns []columnModel
	diags = model.Columns.ElementsAs(ctx, &columns, false)
	if diags.HasError() {
		t.Fatalf("ElementsAs diagnostics = %v", diags.Errors())
	}
	if len(columns) != 1 || columns[0].Name.ValueString() != "region" {
		t.Fatalf("Columns = %#v", columns)
	}
	if columns[0].DataType.ValueString() != "VARCHAR" {
		t.Fatalf("column data type = %q", columns[0].DataType.ValueString())
	}
	if columns[0].Description.ValueString() != columnDescription {
		t.Fatalf("column description = %q", columns[0].Description.ValueString())
	}
}
