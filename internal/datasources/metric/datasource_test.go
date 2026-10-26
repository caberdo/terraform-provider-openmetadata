package metric

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_metric" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_metric")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{
		"id", "fully_qualified_name", "name", "description", "metric_type",
		"unit_of_measurement", "granularity", "expression_code", "updated_at",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "WeeklyRevenue"
	name := "WeeklyRevenue"
	metricType := ometa.MetricMetricTypeSUM

	var model metricModel
	apply(&ometa.Metric{
		Id:                 id,
		Name:               name,
		FullyQualifiedName: &fqn,
		Description:        stringPointer("Revenue grouped by week."),
		MetricType:         &metricType,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != name {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.MetricType.ValueString() != "SUM" {
		t.Fatalf("MetricType = %q", model.MetricType.ValueString())
	}
}

func stringPointer(value string) *string { return &value }
