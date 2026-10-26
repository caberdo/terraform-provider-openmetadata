package dashboard_data_model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// columnAttrTypes is the element type shared by the columns list and its
// conversion helpers.
var columnAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"data_type":    types.StringType,
	"display_name": types.StringType,
	"description":  types.StringType,
}

// columnModel is the Terraform representation of a single data model column.
type columnModel struct {
	Name        types.String `tfsdk:"name"`
	DataType    types.String `tfsdk:"data_type"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
}

// dashboardDataModelModel is the Terraform representation of an OpenMetadata
// dashboard data model.
type dashboardDataModelModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Service            types.String  `tfsdk:"service"`
	DataModelType      types.String  `tfsdk:"data_model_type"`
	SQL                types.String  `tfsdk:"sql"`
	Columns            types.List    `tfsdk:"columns"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Fields the
// server omits keep their prior model value so a sparse response never clears
// configured attributes.
func apply(ctx context.Context, dataModel *ometa.DashboardDataModel, model *dashboardDataModelModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(dataModel.Id.String())
	model.Name = types.StringValue(dataModel.Name)
	model.FullyQualifiedName = convert.StringValue(dataModel.FullyQualifiedName)
	model.DisplayName = convert.StringValue(dataModel.DisplayName)
	model.Description = convert.StringValueOr(dataModel.Description, "")
	model.DataModelType = types.StringValue(string(dataModel.DataModelType))
	if dataModel.Sql != nil {
		model.SQL = convert.StringValue(dataModel.Sql)
	}
	if dataModel.Columns != nil {
		columns, columnDiags := columnsToList(ctx, dataModel.Columns)
		diags.Append(columnDiags...)
		if !columnDiags.HasError() {
			model.Columns = columns
		}
	}
	if dataModel.Service != nil {
		model.Service = referenceValue(dataModel.Service)
	}
	model.Version = convert.Float64Value(dataModel.Version)
	model.UpdatedAt = convert.TimestampValue(dataModel.UpdatedAt)
	model.UpdatedBy = convert.StringValue(dataModel.UpdatedBy)

	return diags
}

// referenceValue maps an entity reference to the fully qualified name the API
// accepts on write, falling back to the plain name.
func referenceValue(reference *ometa.EntityReference) types.String {
	if reference.FullyQualifiedName != nil {
		return convert.StringValue(reference.FullyQualifiedName)
	}
	return convert.StringValue(reference.Name)
}

// columnsFromList maps the Terraform columns list to API columns.
func columnsFromList(ctx context.Context, value types.List) ([]ometa.Column, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	var models []columnModel
	diags.Append(value.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil, diags
	}

	columns := make([]ometa.Column, 0, len(models))
	for _, model := range models {
		columns = append(columns, ometa.Column{
			Name:        model.Name.ValueString(),
			DataType:    ometa.ColumnDataType(model.DataType.ValueString()),
			DisplayName: convert.StringPointer(model.DisplayName),
			Description: convert.StringPointer(model.Description),
		})
	}

	return columns, diags
}

// columnsToList maps API columns to the Terraform columns list.
func columnsToList(ctx context.Context, columns []ometa.Column) (types.List, diag.Diagnostics) {
	models := make([]columnModel, 0, len(columns))
	for _, column := range columns {
		models = append(models, columnModel{
			Name:        types.StringValue(column.Name),
			DataType:    types.StringValue(string(column.DataType)),
			DisplayName: convert.StringValue(column.DisplayName),
			Description: convert.StringValueOr(column.Description, ""),
		})
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: columnAttrTypes}, models)
}
