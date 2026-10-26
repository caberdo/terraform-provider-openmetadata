package metric

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// metricModel is the Terraform representation of an OpenMetadata metric. The
// expression is flattened into two attributes so the schema stays free of
// nested blocks.
type metricModel struct {
	ID                      types.String  `tfsdk:"id"`
	Name                    types.String  `tfsdk:"name"`
	FullyQualifiedName      types.String  `tfsdk:"fully_qualified_name"`
	DisplayName             types.String  `tfsdk:"display_name"`
	Description             types.String  `tfsdk:"description"`
	MetricType              types.String  `tfsdk:"metric_type"`
	UnitOfMeasurement       types.String  `tfsdk:"unit_of_measurement"`
	CustomUnitOfMeasurement types.String  `tfsdk:"custom_unit_of_measurement"`
	Granularity             types.String  `tfsdk:"granularity"`
	ExpressionCode          types.String  `tfsdk:"expression_code"`
	ExpressionLanguage      types.String  `tfsdk:"expression_language"`
	Version                 types.Float64 `tfsdk:"version"`
	UpdatedAt               types.String  `tfsdk:"updated_at"`
	UpdatedBy               types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(metric *ometa.Metric, model *metricModel) {
	model.ID = types.StringValue(metric.Id.String())
	model.Name = types.StringValue(metric.Name)
	model.FullyQualifiedName = convert.StringValue(metric.FullyQualifiedName)
	model.DisplayName = convert.StringValue(metric.DisplayName)
	model.Description = convert.StringValueOr(metric.Description, "")
	model.MetricType = convert.StringValue(metricType(metric.MetricType))
	model.UnitOfMeasurement = convert.StringValue(unitOfMeasurement(metric.UnitOfMeasurement))
	model.CustomUnitOfMeasurement = convert.StringValue(metric.CustomUnitOfMeasurement)
	model.Granularity = convert.StringValue(granularity(metric.Granularity))
	model.ExpressionCode = types.StringNull()
	model.ExpressionLanguage = types.StringNull()
	if expression := metric.MetricExpression; expression != nil {
		model.ExpressionCode = convert.StringValue(expression.Code)
		if expression.Language != nil {
			model.ExpressionLanguage = types.StringValue(string(*expression.Language))
		}
	}
	model.Version = convert.Float64Value(metric.Version)
	model.UpdatedAt = convert.TimestampValue(metric.UpdatedAt)
	model.UpdatedBy = convert.StringValue(metric.UpdatedBy)
}

// metricType renders an optional enum as a Terraform string.
func metricType(value *ometa.MetricMetricType) *string {
	if value == nil {
		return nil
	}
	rendered := string(*value)
	return &rendered
}

// unitOfMeasurement renders an optional enum as a Terraform string.
func unitOfMeasurement(value *ometa.MetricUnitOfMeasurement) *string {
	if value == nil {
		return nil
	}
	rendered := string(*value)
	return &rendered
}

// granularity renders an optional enum as a Terraform string.
func granularity(value *ometa.MetricGranularity) *string {
	if value == nil {
		return nil
	}
	rendered := string(*value)
	return &rendered
}
