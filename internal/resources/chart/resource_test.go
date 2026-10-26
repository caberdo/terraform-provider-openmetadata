package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_chart" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_chart")
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
		t.Fatal("service must carry RequiresReplace")
	}

	chartType, ok := resp.Schema.Attributes["chart_type"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("chart_type attribute is %T, want StringAttribute", resp.Schema.Attributes["chart_type"])
	}
	if !chartType.Optional || !chartType.Computed {
		t.Fatal("chart_type must be optional and computed")
	}
	if chartType.Default == nil {
		t.Fatal("chart_type must have the Other default")
	}

	dashboards, ok := resp.Schema.Attributes["dashboards"].(rschema.ListAttribute)
	if !ok {
		t.Fatalf("dashboards attribute is %T, want ListAttribute", resp.Schema.Attributes["dashboards"])
	}
	if !dashboards.Optional || !dashboards.Computed {
		t.Fatal("dashboards must be optional and computed")
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"version",
		"updated_at",
		"updated_by",
	} {
		if attribute == "version" {
			version, ok := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !ok || !version.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	displayName := "Revenue by region"
	fqn := "looker.revenue_by_region"
	serviceName := "looker"
	sourceURL := "https://looker.example.com/charts/1"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	chartType := ometa.ChartChartType("Bar")
	description := "Revenue chart."
	dashboardName := "sales"
	dashboardFQN := "looker.sales"

	var model chartModel
	apply(&ometa.Chart{
		Id:                 id,
		Name:               "revenue_by_region",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			Name:               &serviceName,
			FullyQualifiedName: &fqn,
		},
		ChartType: &chartType,
		SourceUrl: &sourceURL,
		Dashboards: &[]ometa.EntityReference{
			{Name: &dashboardName, FullyQualifiedName: &dashboardFQN},
			{Name: &dashboardName},
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "revenue_by_region" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Revenue chart." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Service.ValueString() != fqn {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.ChartType.ValueString() != "Bar" {
		t.Fatalf("ChartType = %q", model.ChartType.ValueString())
	}
	if model.SourceURL.ValueString() != sourceURL {
		t.Fatalf("SourceURL = %q", model.SourceURL.ValueString())
	}
	if len(model.Dashboards.Elements()) != 2 {
		t.Fatalf("Dashboards = %v", model.Dashboards)
	}
	if model.Dashboards.Elements()[0].(types.String).ValueString() != dashboardFQN {
		t.Fatalf("Dashboards[0] = %v", model.Dashboards.Elements()[0])
	}
	if model.Dashboards.Elements()[1].(types.String).ValueString() != dashboardName {
		t.Fatalf("Dashboards[1] = %v", model.Dashboards.Elements()[1])
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

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "looker"

	model := chartModel{
		Service:    types.StringValue("looker"),
		Dashboards: types.ListUnknown(types.StringType),
	}
	apply(&ometa.Chart{
		Id:      id,
		Name:    "bare",
		Service: ometa.EntityReference{Name: &serviceName},
	}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.ChartType.ValueString() != "Other" {
		t.Fatalf("ChartType = %q, want the Other default", model.ChartType.ValueString())
	}
	if !model.SourceURL.IsNull() {
		t.Fatalf("SourceURL = %v, want null so state keeps its prior value", model.SourceURL)
	}
	if model.Service.ValueString() != serviceName {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if len(model.Dashboards.Elements()) != 0 {
		t.Fatalf("Dashboards = %v, want an empty list", model.Dashboards)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := chartModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		ChartType:   types.StringValue("Bar"),
		SourceURL:   types.StringValue("https://example.com/old"),
		Dashboards:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sales")}),
	}

	tests := []struct {
		name       string
		plan       chartModel
		state      chartModel
		dashboards []ometa.EntityReference
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
			plan: chartModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				ChartType:   base.ChartType,
				SourceURL:   base.SourceURL,
				Dashboards:  base.Dashboards,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: chartModel{
				DisplayName: types.StringNull(),
				Description: base.Description,
				ChartType:   base.ChartType,
				SourceURL:   base.SourceURL,
				Dashboards:  base.Dashboards,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: chartModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				ChartType:   base.ChartType,
				SourceURL:   base.SourceURL,
				Dashboards:  base.Dashboards,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "chart type changed",
			plan: chartModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ChartType:   types.StringValue("Line"),
				SourceURL:   base.SourceURL,
				Dashboards:  base.Dashboards,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/chartType",
			wantOp:    "add",
			wantValue: "Line",
		},
		{
			name: "source url changed",
			plan: chartModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ChartType:   base.ChartType,
				SourceURL:   types.StringValue("https://example.com/new"),
				Dashboards:  base.Dashboards,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/sourceUrl",
			wantOp:    "add",
			wantValue: "https://example.com/new",
		},
		{
			name: "dashboards changed",
			plan: chartModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ChartType:   base.ChartType,
				SourceURL:   base.SourceURL,
				Dashboards:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sales"), types.StringValue("marketing")}),
			},
			state:      base,
			dashboards: []ometa.EntityReference{{Type: "dashboard"}},
			wantOps:    1,
			wantPath:   "/dashboards",
			wantOp:     "add",
			wantValue:  []ometa.EntityReference{{Type: "dashboard"}},
		},
		{
			name: "dashboards cleared",
			plan: chartModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ChartType:   base.ChartType,
				SourceURL:   base.SourceURL,
				Dashboards:  types.ListValueMust(types.StringType, []attr.Value{}),
			},
			state:      base,
			dashboards: []ometa.EntityReference{},
			wantOps:    1,
			wantPath:   "/dashboards",
			wantOp:     "add",
			wantValue:  []ometa.EntityReference{},
		},
		{
			name: "service cannot be patched",
			plan: chartModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				ChartType:   base.ChartType,
				SourceURL:   base.SourceURL,
				Dashboards:  base.Dashboards,
				Service:     types.StringValue("another-service"),
			},
			state:   base,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state, tt.dashboards)
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
				switch want := tt.wantValue.(type) {
				case string:
					if ops[0].Value != want {
						t.Fatalf("value = %#v, want %#v", ops[0].Value, want)
					}
				case []ometa.EntityReference:
					got, ok := ops[0].Value.([]ometa.EntityReference)
					if !ok {
						t.Fatalf("value = %#v, want []ometa.EntityReference", ops[0].Value)
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("value = %#v, want %#v", got, want)
					}
				case []string:
					got, ok := ops[0].Value.([]string)
					if !ok {
						t.Fatalf("value = %#v, want []string", ops[0].Value)
					}
					if len(got) != len(want) {
						t.Fatalf("value = %v, want %v", got, want)
					}
					for i := range want {
						if got[i] != want[i] {
							t.Fatalf("value = %v, want %v", got, want)
						}
					}
				}
			}
		})
	}
}
