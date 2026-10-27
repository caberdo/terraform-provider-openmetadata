package test_case

import (
	"context"
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

	if resp.TypeName != "openmetadata_test_case" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_test_case")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "entity_link", "test_definition"} {
		stringAttribute, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !stringAttribute.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	for _, attribute := range []string{
		"fully_qualified_name",
		"entity_fqn",
		"test_suite",
		"test_case_status",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}

	parameters, ok := resp.Schema.Attributes["parameter_values"].(rschema.ListNestedAttribute)
	if !ok || !parameters.Optional {
		t.Fatalf("parameter_values attribute is %T, want an optional ListNestedAttribute", resp.Schema.Attributes["parameter_values"])
	}
	for _, attribute := range []string{"name", "value"} {
		if _, ok := parameters.NestedObject.Attributes[attribute]; !ok {
			t.Fatalf("parameter_values is missing nested attribute %q", attribute)
		}
	}

	dimensions, ok := resp.Schema.Attributes["dimension_columns"].(rschema.ListAttribute)
	if !ok || !dimensions.Optional {
		t.Fatalf("dimension_columns attribute is %T, want an optional ListAttribute", resp.Schema.Attributes["dimension_columns"])
	}
	if dimensions.ElementType != types.StringType {
		t.Fatalf("dimension_columns element type = %v, want string", dimensions.ElementType)
	}

	for _, attribute := range []string{"compute_passed_failed_row_count", "use_dynamic_assertion"} {
		boolAttribute, ok := resp.Schema.Attributes[attribute].(rschema.BoolAttribute)
		if !ok || !boolAttribute.Optional || !boolAttribute.Computed {
			t.Fatalf("%s must be an optional computed bool", attribute)
		}
	}

	topDimensions, ok := resp.Schema.Attributes["top_dimensions"].(rschema.Int64Attribute)
	if !ok || !topDimensions.Optional {
		t.Fatalf("top_dimensions attribute is %T, want an optional Int64Attribute", resp.Schema.Attributes["top_dimensions"])
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "svc.db.public.users.users_row_count"
	entityFQN := "svc.db.public.users"
	displayName := "Row count"
	description := "Checks the row count."
	definitionFQN := "tableRowCountToBeBetween"
	suiteFQN := "svc.db.public.users.testSuite"
	parameterName := "min"
	parameterValue := "10"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	computePassedFailedRowCount := true
	useDynamicAssertion := true
	topDimensions := int32(5)
	status := ometa.TestCaseTestCaseStatusSuccess

	var model testCaseModel
	apply(&ometa.TestCase{
		Id:                          &id,
		Name:                        "users_row_count",
		FullyQualifiedName:          &fqn,
		DisplayName:                 &displayName,
		Description:                 &description,
		EntityLink:                  "<#E::table::svc.db.public.users>",
		EntityFQN:                   &entityFQN,
		TestDefinition:              ometa.EntityReference{FullyQualifiedName: &definitionFQN},
		TestSuite:                   ometa.EntityReference{FullyQualifiedName: &suiteFQN},
		ParameterValues:             &[]ometa.TestCaseParameterValue{{Name: &parameterName, Value: &parameterValue}},
		DimensionColumns:            &[]string{"country"},
		ComputePassedFailedRowCount: &computePassedFailedRowCount,
		UseDynamicAssertion:         &useDynamicAssertion,
		TopDimensions:               &topDimensions,
		TestCaseStatus:              &status,
		Version:                     &version,
		UpdatedAt:                   &updatedAt,
		UpdatedBy:                   &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "users_row_count" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Checks the row count." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.EntityLink.ValueString() != "<#E::table::svc.db.public.users>" {
		t.Fatalf("EntityLink = %q", model.EntityLink.ValueString())
	}
	if model.EntityFQN.ValueString() != entityFQN {
		t.Fatalf("EntityFQN = %q", model.EntityFQN.ValueString())
	}
	if model.TestDefinition.ValueString() != definitionFQN {
		t.Fatalf("TestDefinition = %q", model.TestDefinition.ValueString())
	}
	if model.TestSuite.ValueString() != suiteFQN {
		t.Fatalf("TestSuite = %q", model.TestSuite.ValueString())
	}
	if model.TestCaseStatus.ValueString() != "Success" {
		t.Fatalf("TestCaseStatus = %q", model.TestCaseStatus.ValueString())
	}
	if !model.ComputePassedFailedRowCount.ValueBool() {
		t.Fatal("ComputePassedFailedRowCount = false")
	}
	if !model.UseDynamicAssertion.ValueBool() {
		t.Fatal("UseDynamicAssertion = false")
	}
	if model.TopDimensions.ValueInt64() != 5 {
		t.Fatalf("TopDimensions = %d", model.TopDimensions.ValueInt64())
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

	var parameters []testCaseParameterValueModel
	if diags := model.ParameterValues.ElementsAs(context.Background(), &parameters, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	if len(parameters) != 1 || parameters[0].Name.ValueString() != parameterName || parameters[0].Value.ValueString() != parameterValue {
		t.Fatalf("ParameterValues = %#v", parameters)
	}

	var dimensions []string
	if diags := model.DimensionColumns.ElementsAs(context.Background(), &dimensions, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	if len(dimensions) != 1 || dimensions[0] != "country" {
		t.Fatalf("DimensionColumns = %#v", dimensions)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model testCaseModel
	apply(&ometa.TestCase{Id: &id, Name: "bare", EntityLink: "<#E::table::svc.db.users>"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.TestCaseStatus.IsNull() {
		t.Fatalf("TestCaseStatus = %v, want null", model.TestCaseStatus)
	}
	if !model.TopDimensions.IsNull() {
		t.Fatalf("TopDimensions = %v, want null", model.TopDimensions)
	}
	if model.ComputePassedFailedRowCount.ValueBool() {
		t.Fatal("ComputePassedFailedRowCount = true, want the false default")
	}
	if !model.ParameterValues.IsNull() {
		t.Fatalf("ParameterValues = %v, want null", model.ParameterValues)
	}
	if !model.TestDefinition.IsNull() {
		t.Fatalf("TestDefinition = %v, want null", model.TestDefinition)
	}
}

func TestPatchOps(t *testing.T) {
	base := testCaseModel{
		DisplayName:                 types.StringValue("Display"),
		Description:                 types.StringValue("before"),
		ParameterValues:             parametersList("min", "1"),
		DimensionColumns:            stringListValue([]string{"country"}),
		ComputePassedFailedRowCount: types.BoolValue(false),
		UseDynamicAssertion:         types.BoolValue(false),
		TopDimensions:               types.Int64Value(3),
	}

	tests := []struct {
		name      string
		plan      testCaseModel
		state     testCaseModel
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
			name: "display name replaced",
			plan: testCaseModel{
				DisplayName:                 types.StringValue("New display"),
				Description:                 base.Description,
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: testCaseModel{
				DisplayName:                 types.StringNull(),
				Description:                 base.Description,
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 types.StringValue("after"),
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "parameter values replaced",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 base.Description,
				ParameterValues:             parametersList("max", "100"),
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/parameterValues",
			wantOp:   "add",
		},
		{
			name: "parameter values removed",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 base.Description,
				ParameterValues:             types.ListNull(testCaseParameterValueObjectType),
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/parameterValues",
			wantOp:   "remove",
		},
		{
			name: "dimension columns changed",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 base.Description,
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            stringListValue([]string{"country", "device"}),
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/dimensionColumns",
			wantOp:   "add",
		},
		{
			name: "compute row count toggled",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 base.Description,
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: types.BoolValue(true),
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               base.TopDimensions,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/computePassedFailedRowCount",
			wantOp:    "add",
			wantValue: true,
		},
		{
			name: "dynamic assertion toggled",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 base.Description,
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         types.BoolValue(true),
				TopDimensions:               base.TopDimensions,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/useDynamicAssertion",
			wantOp:    "add",
			wantValue: true,
		},
		{
			name: "top dimensions changed",
			plan: testCaseModel{
				DisplayName:                 base.DisplayName,
				Description:                 base.Description,
				ParameterValues:             base.ParameterValues,
				DimensionColumns:            base.DimensionColumns,
				ComputePassedFailedRowCount: base.ComputePassedFailedRowCount,
				UseDynamicAssertion:         base.UseDynamicAssertion,
				TopDimensions:               types.Int64Value(10),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/topDimensions",
			wantOp:    "add",
			wantValue: int64(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, diags := patchOps(context.Background(), tt.plan, tt.state)
			if diags.HasError() {
				t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
			}
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
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}

// parametersList builds the Terraform shape of one parameter value.
func parametersList(name, value string) types.List {
	element := types.ObjectValueMust(testCaseParameterValueObjectType.AttrTypes, map[string]attr.Value{
		"name":  types.StringValue(name),
		"value": types.StringValue(value),
	})

	return types.ListValueMust(testCaseParameterValueObjectType, []attr.Value{element})
}
