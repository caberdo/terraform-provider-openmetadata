package dashboard

import (
	"context"
	"reflect"
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

	if resp.TypeName != "openmetadata_dashboard" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_dashboard")
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
		"description",
		"service",
		"dashboard_type",
		"source_url",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	for _, attribute := range []string{"charts", "data_models"} {
		list, ok := resp.Schema.Attributes[attribute].(dschema.ListAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want ListAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !list.Computed {
			t.Fatalf("%s must be computed", attribute)
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
	serviceFQN := "sales_dashboards"
	dashboardType := ometa.DashboardDashboardTypeReport
	sourceURL := "https://bi.example.com/revenue"
	chartFQN := "sales_dashboards.chart.revenue_by_month"
	dataModelFQN := "sales_dashboards.model.revenue"

	var model dashboardModel
	apply(&ometa.Dashboard{
		Id:                 id,
		Name:               "revenue",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        ometa.Str("Revenue dashboard."),
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
		},
		DashboardType: &dashboardType,
		SourceUrl:     &sourceURL,
		Charts: &[]ometa.EntityReference{
			{FullyQualifiedName: &chartFQN},
		},
		DataModels: &[]ometa.EntityReference{
			{FullyQualifiedName: &dataModelFQN},
		},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "revenue" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Revenue dashboard." {
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
	if got := listValues(t, model.Charts); !reflect.DeepEqual(got, []string{chartFQN}) {
		t.Fatalf("Charts = %v", got)
	}
	if got := listValues(t, model.DataModels); !reflect.DeepEqual(got, []string{dataModelFQN}) {
		t.Fatalf("DataModels = %v", got)
	}
}

func listValues(t *testing.T, list types.List) []string {
	t.Helper()

	var values []string
	if diags := list.ElementsAs(context.Background(), &values, false); diags.HasError() {
		t.Fatalf("list elements: %v", diags.Errors())
	}
	return values
}
