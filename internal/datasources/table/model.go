package table

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// tableColumnObjectType describes one nested column returned by the lookup.
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

// tableModel is the Terraform representation of a table lookup. The id and
// fully_qualified_name attributes are both selectors and results.
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

// apply copies the API representation into the Terraform model.
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
