package test_definition

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

	if resp.TypeName != "openmetadata_test_definition" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_test_definition")
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
		t.Fatalf("name must be a required string attribute, got %T", resp.Schema.Attributes["name"])
	}

	entityType, ok := resp.Schema.Attributes["entity_type"].(rschema.StringAttribute)
	if !ok || !entityType.Required {
		t.Fatalf("entity_type must be a required string attribute, got %T", resp.Schema.Attributes["entity_type"])
	}

	platforms, ok := resp.Schema.Attributes["test_platforms"].(rschema.ListAttribute)
	if !ok || !platforms.Required {
		t.Fatalf("test_platforms must be a required list attribute, got %T", resp.Schema.Attributes["test_platforms"])
	}

	provider, ok := resp.Schema.Attributes["source"].(rschema.StringAttribute)
	if !ok || !provider.Optional || !provider.Computed {
		t.Fatalf("source must be an optional computed string attribute, got %T", resp.Schema.Attributes["source"])
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
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
		t.Fatalf("version must be a computed float")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed {
		t.Fatalf("description must be optional and computed")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "custom_table_row_count"
	displayName := "Custom table row count"
	description := "Counts rows in a table."
	sqlExpression := "SELECT COUNT(*) FROM {{ table }}"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	entityType := ometa.TestDefinitionEntityTypeTABLE
	dimension := ometa.TestDefinitionDataQualityDimensionAccuracy
	provider := ometa.TestDefinitionProviderUser
	dataTypes := []ometa.TestDefinitionSupportedDataTypes{
		ometa.TestDefinitionSupportedDataTypesINT,
		ometa.TestDefinitionSupportedDataTypesVARCHAR,
	}
	services := []string{"Mysql", "Snowflake"}

	var model testDefinitionModel
	apply(&ometa.TestDefinition{
		Id:                   &id,
		Name:                 "custom_table_row_count",
		FullyQualifiedName:   &fqn,
		DisplayName:          &displayName,
		Description:          &description,
		EntityType:           &entityType,
		TestPlatforms:        []ometa.TestDefinitionTestPlatforms{ometa.TestDefinitionTestPlatformsOther},
		DataQualityDimension: &dimension,
		SupportedDataTypes:   &dataTypes,
		SupportedServices:    &services,
		SqlExpression:        &sqlExpression,
		Provider:             &provider,
		Version:              &version,
		UpdatedAt:            &updatedAt,
		UpdatedBy:            &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "custom_table_row_count" {
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
	if model.EntityType.ValueString() != "TABLE" {
		t.Fatalf("EntityType = %q", model.EntityType.ValueString())
	}
	if model.DataQualityDimension.ValueString() != "Accuracy" {
		t.Fatalf("DataQualityDimension = %q", model.DataQualityDimension.ValueString())
	}
	if model.SQLExpression.ValueString() != sqlExpression {
		t.Fatalf("SQLExpression = %q", model.SQLExpression.ValueString())
	}
	if model.Provider.ValueString() != "user" {
		t.Fatalf("Provider = %q", model.Provider.ValueString())
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

	platformValues := listStrings(t, model.TestPlatforms)
	if len(platformValues) != 1 || platformValues[0] != "Other" {
		t.Fatalf("TestPlatforms = %v", platformValues)
	}

	if got := listStrings(t, model.SupportedDataTypes); len(got) != 2 || got[0] != "INT" || got[1] != "VARCHAR" {
		t.Fatalf("SupportedDataTypes = %v", got)
	}

	if got := listStrings(t, model.SupportedServices); len(got) != 2 || got[0] != "Mysql" || got[1] != "Snowflake" {
		t.Fatalf("SupportedServices = %v", got)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := testDefinitionModel{
		EntityType:    types.StringValue("COLUMN"),
		TestPlatforms: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Other")}),
	}
	apply(&ometa.TestDefinition{Id: &id, Name: "bare"}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.Provider.ValueString() != "user" {
		t.Fatalf("Provider = %q, want the user default", model.Provider.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.DataQualityDimension.IsNull() {
		t.Fatalf("DataQualityDimension = %v, want null", model.DataQualityDimension)
	}
	if !model.SQLExpression.IsNull() {
		t.Fatalf("SQLExpression = %v, want null", model.SQLExpression)
	}
	if !model.SupportedServices.IsNull() {
		t.Fatalf("SupportedServices = %v, want null", model.SupportedServices)
	}
	if model.EntityType.ValueString() != "COLUMN" {
		t.Fatalf("EntityType = %q, want the plan value preserved", model.EntityType.ValueString())
	}
	if model.TestPlatforms.IsNull() {
		t.Fatal("TestPlatforms = null, want the plan value preserved")
	}
}

func TestPatchOps(t *testing.T) {
	base := testDefinitionModel{
		DisplayName:          types.StringValue("Display"),
		Description:          types.StringValue("before"),
		EntityType:           types.StringValue("TABLE"),
		TestPlatforms:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Other")}),
		DataQualityDimension: types.StringValue("Accuracy"),
		SupportedDataTypes:   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("INT")}),
		SupportedServices:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Mysql")}),
		SQLExpression:        types.StringValue("SELECT 1"),
		Provider:             types.StringValue("user"),
	}

	tests := []struct {
		name      string
		plan      testDefinitionModel
		state     testDefinitionModel
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
			name:      "display name replaced",
			plan:      with(base, func(m *testDefinitionModel) { m.DisplayName = types.StringValue("New display") }),
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name:     "display name removed",
			plan:     with(base, func(m *testDefinitionModel) { m.DisplayName = types.StringNull() }),
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name:      "description changed",
			plan:      with(base, func(m *testDefinitionModel) { m.Description = types.StringValue("after") }),
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name:      "dimension changed",
			plan:      with(base, func(m *testDefinitionModel) { m.DataQualityDimension = types.StringValue("Completeness") }),
			state:     base,
			wantOps:   1,
			wantPath:  "/dataQualityDimension",
			wantOp:    "add",
			wantValue: "Completeness",
		},
		{
			name:     "dimension removed",
			plan:     with(base, func(m *testDefinitionModel) { m.DataQualityDimension = types.StringNull() }),
			state:    base,
			wantOps:  1,
			wantPath: "/dataQualityDimension",
			wantOp:   "remove",
		},
		{
			name: "supported data types changed",
			plan: with(base, func(m *testDefinitionModel) {
				m.SupportedDataTypes = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("VARCHAR")})
			}),
			state:    base,
			wantOps:  1,
			wantPath: "/supportedDataTypes",
			wantOp:   "add",
		},
		{
			name: "supported services removed",
			plan: with(base, func(m *testDefinitionModel) {
				m.SupportedServices = types.ListNull(types.StringType)
			}),
			state:    base,
			wantOps:  1,
			wantPath: "/supportedServices",
			wantOp:   "remove",
		},
		{
			name:      "sql expression changed",
			plan:      with(base, func(m *testDefinitionModel) { m.SQLExpression = types.StringValue("SELECT 2") }),
			state:     base,
			wantOps:   1,
			wantPath:  "/sqlExpression",
			wantOp:    "add",
			wantValue: "SELECT 2",
		},
		{
			name:     "sql expression removed",
			plan:     with(base, func(m *testDefinitionModel) { m.SQLExpression = types.StringNull() }),
			state:    base,
			wantOps:  1,
			wantPath: "/sqlExpression",
			wantOp:   "remove",
		},
		{
			name: "replace-only attributes are never patched",
			plan: with(base, func(m *testDefinitionModel) {
				m.EntityType = types.StringValue("COLUMN")
				m.Provider = types.StringValue("system")
				m.TestPlatforms = types.ListValueMust(types.StringType, []attr.Value{types.StringValue("OpenMetadata")})
			}),
			state:   base,
			wantOps: 0,
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

// with copies a model and applies a mutation, so table cases stay readable.
func with(base testDefinitionModel, mutate func(*testDefinitionModel)) testDefinitionModel {
	clone := base
	mutate(&clone)
	return clone
}

// listStrings renders a Terraform string list for assertions.
func listStrings(t *testing.T, value types.List) []string {
	t.Helper()

	var out []string
	diags := value.ElementsAs(context.Background(), &out, false)
	if diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	return out
}
