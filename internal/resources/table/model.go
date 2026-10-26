package table

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// tableColumnObjectType describes one nested column. The resource and the data
// source share this shape so a column read from the API can be written straight
// back into state.
var tableColumnObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"name":             types.StringType,
	"data_type":        types.StringType,
	"data_length":      types.Int64Type,
	"description":      types.StringType,
	"display_name":     types.StringType,
	"ordinal_position": types.Int64Type,
	"precision":        types.Int64Type,
	"scale":            types.Int64Type,
}}

// tableColumnModel is the Terraform representation of one OpenMetadata column.
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

// tableModel is the Terraform representation of an OpenMetadata table.
type tableModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	DatabaseSchema     types.String  `tfsdk:"database_schema"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Columns            types.List    `tfsdk:"columns"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Required
// selectors are left untouched when the server omits the reference.
func apply(tbl *ometa.Table, model *tableModel) {
	model.ID = types.StringValue(tbl.Id.String())
	model.Name = types.StringValue(tbl.Name)
	model.FullyQualifiedName = convert.StringValue(tbl.FullyQualifiedName)
	model.DisplayName = convert.StringValue(tbl.DisplayName)
	model.Description = convert.StringValueOr(tbl.Description, "")
	model.Columns = columnsValue(tbl.Columns)
	model.Version = convert.Float64Value(tbl.Version)
	model.UpdatedAt = convert.TimestampValue(tbl.UpdatedAt)
	model.UpdatedBy = convert.StringValue(tbl.UpdatedBy)

	if tbl.DatabaseSchema != nil {
		if fqn := tbl.DatabaseSchema.FullyQualifiedName; fqn != nil {
			model.DatabaseSchema = types.StringValue(*fqn)
		} else if name := tbl.DatabaseSchema.Name; name != nil {
			model.DatabaseSchema = types.StringValue(*name)
		}
	}
}

// columnsValue maps API columns to a Terraform list in the server's order.
func columnsValue(columns []ometa.Column) types.List {
	elements := make([]attr.Value, 0, len(columns))
	for i, column := range columns {
		ordinal := types.Int64Value(int64(i + 1))
		if column.OrdinalPosition != nil {
			ordinal = types.Int64Value(int64(*column.OrdinalPosition))
		}

		elements = append(elements, types.ObjectValueMust(tableColumnObjectType.AttrTypes, map[string]attr.Value{
			"name":             types.StringValue(column.Name),
			"data_type":        types.StringValue(string(column.DataType)),
			"data_length":      convert.Int32Value(column.DataLength),
			"description":      convert.StringValue(column.Description),
			"display_name":     convert.StringValue(column.DisplayName),
			"ordinal_position": ordinal,
			"precision":        convert.Int32Value(column.Precision),
			"scale":            convert.Int32Value(column.Scale),
		}))
	}

	return types.ListValueMust(tableColumnObjectType, elements)
}

// columnsToAPI maps planned columns to the SDK shape.
func columnsToAPI(ctx context.Context, value types.List) ([]ometa.Column, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var models []tableColumnModel
	diags := value.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}

	columns := make([]ometa.Column, 0, len(models))
	for _, model := range models {
		column := ometa.Column{
			Name:            model.Name.ValueString(),
			DataType:        ometa.ColumnDataType(model.DataType.ValueString()),
			DataLength:      int32Pointer(model.DataLength),
			Description:     convert.StringPointer(model.Description),
			DisplayName:     convert.StringPointer(model.DisplayName),
			OrdinalPosition: int32Pointer(model.OrdinalPosition),
			Precision:       int32Pointer(model.Precision),
			Scale:           int32Pointer(model.Scale),
		}
		columns = append(columns, column)
	}

	return columns, diags
}

// int32Pointer maps an optional Terraform number to the SDK's int32 pointer.
// Null and unknown values are omitted from the request.
func int32Pointer(value types.Int64) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	converted := int32(value.ValueInt64())
	return &converted
}
