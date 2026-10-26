package dashboard

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	rdefaults "github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_dashboard" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_dashboard")
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

	service, ok := resp.Schema.Attributes["service"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("service attribute is %T, want StringAttribute", resp.Schema.Attributes["service"])
	}
	if !service.Required {
		t.Fatal("service must be required")
	}
	if len(service.PlanModifiers) == 0 {
		t.Fatal("service must require replacement")
	}

	dashboardType, ok := resp.Schema.Attributes["dashboard_type"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("dashboard_type attribute is %T, want StringAttribute", resp.Schema.Attributes["dashboard_type"])
	}
	if !dashboardType.Optional || !dashboardType.Computed {
		t.Fatal("dashboard_type must be optional and computed")
	}
	if dashboardType.Default == nil {
		t.Fatal("dashboard_type must default to Dashboard")
	}
	var defaultResp rdefaults.StringResponse
	dashboardType.Default.DefaultString(context.Background(), rdefaults.StringRequest{}, &defaultResp)
	if defaultResp.Diagnostics.HasError() {
		t.Fatalf("dashboard_type default diagnostics = %v", defaultResp.Diagnostics.Errors())
	}
	if defaultResp.PlanValue.ValueString() != "Dashboard" {
		t.Fatalf("dashboard_type default = %q, want %q", defaultResp.PlanValue.ValueString(), "Dashboard")
	}

	sourceURL, ok := resp.Schema.Attributes["source_url"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("source_url attribute is %T, want StringAttribute", resp.Schema.Attributes["source_url"])
	}
	if !sourceURL.Optional || !sourceURL.Computed || sourceURL.Default == nil {
		t.Fatal("source_url must be optional, computed and defaulted")
	}

	for _, attribute := range []string{"charts", "data_models"} {
		list, ok := resp.Schema.Attributes[attribute].(rschema.ListAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want ListAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !list.Optional || !list.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if list.ElementType != types.StringType {
			t.Fatalf("%s element type = %v, want string", attribute, list.ElementType)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "sales_dashboards.revenue"
	displayName := "Revenue"
	description := "Revenue dashboard."
	serviceFQN := "sales_dashboards"
	dashboardType := ometa.DashboardDashboardTypeReport
	sourceURL := "https://bi.example.com/revenue"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	chartFQN := "sales_dashboards.chart.revenue_by_month"
	chartName := "revenue_by_month"
	dataModelFQN := "sales_dashboards.model.revenue"

	var model dashboardModel
	apply(&ometa.Dashboard{
		Id:                 id,
		Name:               "revenue",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
		},
		DashboardType: &dashboardType,
		SourceUrl:     &sourceURL,
		Charts: &[]ometa.EntityReference{
			{FullyQualifiedName: &chartFQN},
			{Name: &chartName},
		},
		DataModels: &[]ometa.EntityReference{
			{FullyQualifiedName: &dataModelFQN},
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "revenue" {
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
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.DashboardType.ValueString() != "Report" {
		t.Fatalf("DashboardType = %q", model.DashboardType.ValueString())
	}
	if model.SourceURL.ValueString() != sourceURL {
		t.Fatalf("SourceURL = %q", model.SourceURL.ValueString())
	}
	if got := listValues(t, model.Charts); !reflect.DeepEqual(got, []string{chartFQN, chartName}) {
		t.Fatalf("Charts = %v", got)
	}
	if got := listValues(t, model.DataModels); !reflect.DeepEqual(got, []string{dataModelFQN}) {
		t.Fatalf("DataModels = %v", got)
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
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceFQN := "sales_dashboards"
	charts := listOfStrings(t, "sales_dashboards.chart.revenue")

	model := dashboardModel{
		Service:    types.StringValue(serviceFQN),
		SourceURL:  types.StringValue("https://bi.example.com/revenue"),
		Charts:     charts,
		DataModels: listOfStrings(t),
	}
	apply(&ometa.Dashboard{
		Id:   id,
		Name: "revenue",
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
		},
	}, &model)

	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.SourceURL.ValueString() != "https://bi.example.com/revenue" {
		t.Fatalf("SourceURL = %q", model.SourceURL.ValueString())
	}
	if model.DashboardType.ValueString() != "Dashboard" {
		t.Fatalf("DashboardType = %q, want the Dashboard default", model.DashboardType.ValueString())
	}
	if got := listValues(t, model.Charts); !reflect.DeepEqual(got, []string{"sales_dashboards.chart.revenue"}) {
		t.Fatalf("Charts = %v", got)
	}
	if got := listValues(t, model.DataModels); len(got) != 0 {
		t.Fatalf("DataModels = %v", got)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	charts := listOfStrings(t, "sales_dashboards.chart.revenue")
	dataModels := listOfStrings(t, "sales_dashboards.model.revenue")

	base := dashboardModel{
		Name:          types.StringValue("revenue"),
		Service:       types.StringValue("sales_dashboards"),
		DisplayName:   types.StringValue("Display"),
		Description:   types.StringValue("before"),
		DashboardType: types.StringValue("Dashboard"),
		SourceURL:     types.StringValue(""),
		Charts:        charts,
		DataModels:    dataModels,
	}

	tests := []struct {
		name       string
		plan       dashboardModel
		state      dashboardModel
		charts     []ometa.EntityReference
		dataModels []ometa.EntityReference
		wantOps    int
		wantPath   string
		wantOp     string
		wantValue  any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   types.StringValue("New display"),
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        base.Charts,
				DataModels:    base.DataModels,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   types.StringNull(),
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        base.Charts,
				DataModels:    base.DataModels,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   base.DisplayName,
				Description:   types.StringValue("after"),
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        base.Charts,
				DataModels:    base.DataModels,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "dashboard type changed",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
				DashboardType: types.StringValue("Report"),
				SourceURL:     base.SourceURL,
				Charts:        base.Charts,
				DataModels:    base.DataModels,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/dashboardType",
			wantOp:    "add",
			wantValue: "Report",
		},
		{
			name: "source url changed",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     types.StringValue("https://bi.example.com/revenue"),
				Charts:        base.Charts,
				DataModels:    base.DataModels,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/sourceUrl",
			wantOp:    "add",
			wantValue: "https://bi.example.com/revenue",
		},
		{
			name: "charts added",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        listOfStrings(t, "sales_dashboards.chart.revenue", "sales_dashboards.chart.costs"),
				DataModels:    base.DataModels,
			},
			state:     base,
			charts:    []ometa.EntityReference{{Type: "chart"}},
			wantOps:   1,
			wantPath:  "/charts",
			wantOp:    "add",
			wantValue: []ometa.EntityReference{{Type: "chart"}},
		},
		{
			name: "charts cleared",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        listOfStrings(t),
				DataModels:    base.DataModels,
			},
			state:     base,
			charts:    []ometa.EntityReference{},
			wantOps:   1,
			wantPath:  "/charts",
			wantOp:    "add",
			wantValue: []ometa.EntityReference{},
		},
		{
			name: "data models changed",
			plan: dashboardModel{
				Name:          base.Name,
				Service:       base.Service,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        base.Charts,
				DataModels:    listOfStrings(t, "sales_dashboards.model.costs"),
			},
			state:      base,
			dataModels: []ometa.EntityReference{{Type: "dashboardDataModel"}},
			wantOps:    1,
			wantPath:   "/dataModels",
			wantOp:     "add",
			wantValue:  []ometa.EntityReference{{Type: "dashboardDataModel"}},
		},
		{
			name: "service and name are not patched",
			plan: dashboardModel{
				Name:          types.StringValue("renamed"),
				Service:       types.StringValue("other_service"),
				DisplayName:   base.DisplayName,
				Description:   base.Description,
				DashboardType: base.DashboardType,
				SourceURL:     base.SourceURL,
				Charts:        base.Charts,
				DataModels:    base.DataModels,
			},
			state:   base,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state, tt.charts, tt.dataModels)
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
			if tt.wantValue != nil && !reflect.DeepEqual(ops[0].Value, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}

func listOfStrings(t *testing.T, values ...string) types.List {
	t.Helper()

	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	return types.ListValueMust(types.StringType, elements)
}

func listValues(t *testing.T, list types.List) []string {
	t.Helper()

	var values []string
	if diags := list.ElementsAs(context.Background(), &values, false); diags.HasError() {
		t.Fatalf("list elements: %v", diags.Errors())
	}
	return values
}
