package chart

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// chartModel is the Terraform representation of an OpenMetadata chart.
type chartModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Service            types.String  `tfsdk:"service"`
	ChartType          types.String  `tfsdk:"chart_type"`
	SourceURL          types.String  `tfsdk:"source_url"`
	Dashboards         types.List    `tfsdk:"dashboards"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Fields the API
// omits keep their prior value so a partial response never clears state.
func apply(chart *ometa.Chart, model *chartModel) {
	model.ID = types.StringValue(chart.Id.String())
	model.Name = types.StringValue(chart.Name)
	model.FullyQualifiedName = convert.StringValue(chart.FullyQualifiedName)
	model.DisplayName = convert.StringValue(chart.DisplayName)
	model.Description = convert.StringValueOr(chart.Description, "")

	var chartType *string
	if chart.ChartType != nil {
		value := string(*chart.ChartType)
		chartType = &value
	}
	model.ChartType = convert.StringValueOr(chartType, "Other")

	if chart.SourceUrl != nil {
		model.SourceURL = convert.StringValueOr(chart.SourceUrl, "")
	}

	switch {
	case chart.Service.FullyQualifiedName != nil && *chart.Service.FullyQualifiedName != "":
		model.Service = types.StringValue(*chart.Service.FullyQualifiedName)
	case chart.Service.Name != nil && *chart.Service.Name != "":
		model.Service = types.StringValue(*chart.Service.Name)
	}

	switch {
	case chart.Dashboards != nil:
		model.Dashboards = referenceListValue(*chart.Dashboards)
	case model.Dashboards.IsUnknown():
		model.Dashboards = types.ListValueMust(types.StringType, []attr.Value{})
	}

	model.Version = convert.Float64Value(chart.Version)
	model.UpdatedAt = convert.TimestampValue(chart.UpdatedAt)
	model.UpdatedBy = convert.StringValue(chart.UpdatedBy)
}

// referenceListValue maps entity references to the fully qualified names the
// API accepts on write.
func referenceListValue(references []ometa.EntityReference) types.List {
	elements := make([]attr.Value, 0, len(references))
	for _, reference := range references {
		if fqn := reference.FullyQualifiedName; fqn != nil {
			elements = append(elements, types.StringValue(*fqn))
		} else if name := reference.Name; name != nil {
			elements = append(elements, types.StringValue(*name))
		}
	}
	return types.ListValueMust(types.StringType, elements)
}
