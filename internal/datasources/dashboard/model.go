package dashboard

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// dashboardModel is the Terraform representation of a dashboard lookup. The
// id and fully_qualified_name attributes are both selectors and results.
type dashboardModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Service            types.String  `tfsdk:"service"`
	DashboardType      types.String  `tfsdk:"dashboard_type"`
	SourceURL          types.String  `tfsdk:"source_url"`
	Charts             types.List    `tfsdk:"charts"`
	DataModels         types.List    `tfsdk:"data_models"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(dashboard *ometa.Dashboard, model *dashboardModel) {
	model.ID = types.StringValue(dashboard.Id.String())
	model.Name = types.StringValue(dashboard.Name)
	model.FullyQualifiedName = convert.StringValue(dashboard.FullyQualifiedName)
	model.DisplayName = convert.StringValue(dashboard.DisplayName)
	model.Description = convert.StringValueOr(dashboard.Description, "")
	model.Service = entityReferenceValue(dashboard.Service, model.Service)

	var dashboardType *string
	if dashboard.DashboardType != nil {
		value := string(*dashboard.DashboardType)
		dashboardType = &value
	}
	model.DashboardType = convert.StringValueOr(dashboardType, "Dashboard")

	if dashboard.SourceUrl != nil {
		model.SourceURL = convert.StringValueOr(dashboard.SourceUrl, "")
	}

	if dashboard.Charts != nil {
		model.Charts = referenceListValue(*dashboard.Charts)
	} else if model.Charts.IsUnknown() || model.Charts.IsNull() {
		model.Charts = emptyStringList()
	}

	if dashboard.DataModels != nil {
		model.DataModels = referenceListValue(*dashboard.DataModels)
	} else if model.DataModels.IsUnknown() || model.DataModels.IsNull() {
		model.DataModels = emptyStringList()
	}

	model.Version = convert.Float64Value(dashboard.Version)
	model.UpdatedAt = convert.TimestampValue(dashboard.UpdatedAt)
	model.UpdatedBy = convert.StringValue(dashboard.UpdatedBy)
}

// entityReferenceValue prefers the fully qualified name and falls back to the
// name. An empty reference leaves the prior value untouched.
func entityReferenceValue(reference ometa.EntityReference, current types.String) types.String {
	if reference.FullyQualifiedName != nil && *reference.FullyQualifiedName != "" {
		return types.StringValue(*reference.FullyQualifiedName)
	}
	if reference.Name != nil && *reference.Name != "" {
		return types.StringValue(*reference.Name)
	}
	return current
}

// referenceListValue maps entity references to fully qualified names.
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

func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}
