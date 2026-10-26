package chart

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

	if resp.TypeName != "openmetadata_chart" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_chart")
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
		"chart_type",
		"source_url",
		"dashboards",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	id, ok := resp.Schema.Attributes["id"].(dschema.StringAttribute)
	if !ok || !id.Optional || !id.Computed {
		t.Fatal("id must be an optional computed string")
	}

	dashboards, ok := resp.Schema.Attributes["dashboards"].(dschema.ListAttribute)
	if !ok || !dashboards.Computed {
		t.Fatal("dashboards must be a computed list")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "looker.revenue_by_region"
	displayName := "Revenue by region"
	serviceName := "looker"
	sourceURL := "https://looker.example.com/charts/1"
	chartType := ometa.ChartChartType("Bar")
	description := "Revenue chart."
	dashboardFQN := "looker.sales"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model chartModel
	apply(&ometa.Chart{
		Id:                 id,
		Name:               "revenue_by_region",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			Name:               &serviceName,
			FullyQualifiedName: &serviceName,
		},
		ChartType: &chartType,
		SourceUrl: &sourceURL,
		Dashboards: &[]ometa.EntityReference{
			{FullyQualifiedName: &dashboardFQN},
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "revenue_by_region" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Revenue chart." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Service.ValueString() != serviceName {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.ChartType.ValueString() != "Bar" {
		t.Fatalf("ChartType = %q", model.ChartType.ValueString())
	}
	if model.SourceURL.ValueString() != sourceURL {
		t.Fatalf("SourceURL = %q", model.SourceURL.ValueString())
	}
	if len(model.Dashboards.Elements()) != 1 {
		t.Fatalf("Dashboards = %v", model.Dashboards)
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

	var model chartModel
	apply(&ometa.Chart{Id: id, Name: "bare"}, &model)

	if model.ChartType.ValueString() != "Other" {
		t.Fatalf("ChartType = %q, want the Other default", model.ChartType.ValueString())
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if len(model.Dashboards.Elements()) != 0 {
		t.Fatalf("Dashboards = %v, want an empty list", model.Dashboards)
	}
}
