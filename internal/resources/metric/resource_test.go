package metric

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_metric" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_metric")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatal("name must be a required string")
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "version", "updated_at", "updated_by"} {
		if attribute == "version" {
			floatAttribute, ok := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !ok || !floatAttribute.Computed {
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
	fqn := "WeeklyRevenue"
	displayName := "Weekly revenue"
	description := "Revenue grouped by week."
	metricType := ometa.MetricMetricTypeSUM
	unit := ometa.MetricUnitOfMeasurementDOLLARS
	granularity := ometa.MetricGranularityWEEK
	language := ometa.MetricExpressionLanguageSQL
	code := "sum(revenue)"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)

	var model metricModel
	apply(&ometa.Metric{
		Id:                 id,
		Name:               "WeeklyRevenue",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		MetricType:         &metricType,
		UnitOfMeasurement:  &unit,
		Granularity:        &granularity,
		MetricExpression: &ometa.MetricExpression{
			Code:     &code,
			Language: &language,
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "WeeklyRevenue" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.MetricType.ValueString() != "SUM" {
		t.Fatalf("MetricType = %q", model.MetricType.ValueString())
	}
	if model.UnitOfMeasurement.ValueString() != "DOLLARS" {
		t.Fatalf("UnitOfMeasurement = %q", model.UnitOfMeasurement.ValueString())
	}
	if model.Granularity.ValueString() != "WEEK" {
		t.Fatalf("Granularity = %q", model.Granularity.ValueString())
	}
	if model.ExpressionCode.ValueString() != code {
		t.Fatalf("ExpressionCode = %q", model.ExpressionCode.ValueString())
	}
	if model.ExpressionLanguage.ValueString() != "SQL" {
		t.Fatalf("ExpressionLanguage = %q", model.ExpressionLanguage.ValueString())
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model metricModel
	apply(&ometa.Metric{Id: id, Name: "bare"}, &model)

	for name, value := range map[string]types.String{
		"MetricType":         model.MetricType,
		"UnitOfMeasurement":  model.UnitOfMeasurement,
		"Granularity":        model.Granularity,
		"ExpressionCode":     model.ExpressionCode,
		"ExpressionLanguage": model.ExpressionLanguage,
	} {
		if !value.IsNull() {
			t.Fatalf("%s = %v, want null", name, value)
		}
	}
}

func TestPatchOps(t *testing.T) {
	base := metricModel{
		DisplayName:        types.StringValue("Display"),
		Description:        types.StringValue("before"),
		MetricType:         types.StringValue("SUM"),
		Granularity:        types.StringValue("WEEK"),
		ExpressionCode:     types.StringValue("sum(revenue)"),
		ExpressionLanguage: types.StringValue("SQL"),
	}

	tests := []struct {
		name      string
		plan      metricModel
		state     metricModel
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "metric type changed",
			plan: metricModel{
				DisplayName:        base.DisplayName,
				Description:        base.Description,
				MetricType:         types.StringValue("AVERAGE"),
				Granularity:        base.Granularity,
				ExpressionCode:     base.ExpressionCode,
				ExpressionLanguage: base.ExpressionLanguage,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/metricType",
			wantOp:    "add",
			wantValue: "AVERAGE",
		},
		{
			name: "granularity removed",
			plan: metricModel{
				DisplayName:        base.DisplayName,
				Description:        base.Description,
				MetricType:         base.MetricType,
				Granularity:        types.StringNull(),
				ExpressionCode:     base.ExpressionCode,
				ExpressionLanguage: base.ExpressionLanguage,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/granularity",
			wantOp:   "remove",
		},
		{
			name: "expression changed",
			plan: metricModel{
				DisplayName:        base.DisplayName,
				Description:        base.Description,
				MetricType:         base.MetricType,
				Granularity:        base.Granularity,
				ExpressionCode:     types.StringValue("avg(revenue)"),
				ExpressionLanguage: types.StringValue("Python"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/metricExpression",
			wantOp:    "add",
			wantValue: map[string]any{"code": "avg(revenue)", "language": "Python"},
		},
		{
			name: "expression removed",
			plan: metricModel{
				DisplayName:        base.DisplayName,
				Description:        base.Description,
				MetricType:         base.MetricType,
				Granularity:        base.Granularity,
				ExpressionCode:     types.StringNull(),
				ExpressionLanguage: types.StringNull(),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/metricExpression",
			wantOp:   "remove",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
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

func TestCreatePayloadConversions(t *testing.T) {
	if createMetricType(types.StringValue("SUM")) == nil {
		t.Fatal("createMetricType() = nil for a known value")
	}
	if createMetricType(types.StringNull()) != nil {
		t.Fatal("createMetricType() != nil for a null value")
	}
	if createUnitOfMeasurement(types.StringValue("DOLLARS")) == nil {
		t.Fatal("createUnitOfMeasurement() = nil for a known value")
	}
	if createGranularity(types.StringValue("DAY")) == nil {
		t.Fatal("createGranularity() = nil for a known value")
	}

	expression := createExpression(types.StringValue("sum(revenue)"), types.StringValue("SQL"))
	if expression == nil || expression.Code == nil || *expression.Code != "sum(revenue)" {
		t.Fatalf("createExpression() = %#v", expression)
	}
	if expression.Language == nil || *expression.Language != ometa.MetricExpressionLanguageSQL {
		t.Fatalf("expression.Language = %#v", expression.Language)
	}
	if createExpression(types.StringNull(), types.StringNull()) != nil {
		t.Fatal("createExpression() != nil when both fields are null")
	}
}
