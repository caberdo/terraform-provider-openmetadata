package test_case

import (
	"context"
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

	if resp.TypeName != "openmetadata_test_case" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_test_case")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be an optional computed string", attribute)
		}
	}

	for _, attribute := range []string{
		"name",
		"display_name",
		"description",
		"entity_link",
		"test_definition",
		"entity_fqn",
		"test_suite",
		"test_case_status",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	parameters, ok := resp.Schema.Attributes["parameter_values"].(dschema.ListNestedAttribute)
	if !ok || !parameters.Computed {
		t.Fatalf("parameter_values attribute is %T, want a computed ListNestedAttribute", resp.Schema.Attributes["parameter_values"])
	}
	for _, attribute := range []string{"name", "value"} {
		if _, ok := parameters.NestedObject.Attributes[attribute]; !ok {
			t.Fatalf("parameter_values is missing nested attribute %q", attribute)
		}
	}

	dimensions, ok := resp.Schema.Attributes["dimension_columns"].(dschema.ListAttribute)
	if !ok || !dimensions.Computed {
		t.Fatalf("dimension_columns attribute is %T, want a computed ListAttribute", resp.Schema.Attributes["dimension_columns"])
	}
	if dimensions.ElementType != types.StringType {
		t.Fatalf("dimension_columns element type = %v, want string", dimensions.ElementType)
	}

	for _, attribute := range []string{"compute_passed_failed_row_count", "use_dynamic_assertion"} {
		boolAttribute, ok := resp.Schema.Attributes[attribute].(dschema.BoolAttribute)
		if !ok || !boolAttribute.Computed {
			t.Fatalf("%s must be a computed bool", attribute)
		}
	}

	if _, ok := resp.Schema.Attributes["top_dimensions"].(dschema.Int64Attribute); !ok {
		t.Fatal("top_dimensions must be an Int64Attribute")
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
	status := ometa.TestCaseTestCaseStatusFailed

	var model testCaseModel
	apply(&ometa.TestCase{
		Id:                 &id,
		Name:               "users_row_count",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		EntityLink:         "<#E::table::svc.db.public.users>",
		EntityFQN:          &entityFQN,
		TestDefinition:     ometa.EntityReference{FullyQualifiedName: &definitionFQN},
		TestSuite:          ometa.EntityReference{Name: &suiteFQN},
		ParameterValues:    &[]ometa.TestCaseParameterValue{{Name: &parameterName, Value: &parameterValue}},
		TestCaseStatus:     &status,
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
	if model.TestDefinition.ValueString() != definitionFQN {
		t.Fatalf("TestDefinition = %q", model.TestDefinition.ValueString())
	}
	if model.TestSuite.ValueString() != suiteFQN {
		t.Fatalf("TestSuite = %q (want the name fallback)", model.TestSuite.ValueString())
	}
	if model.TestCaseStatus.ValueString() != "Failed" {
		t.Fatalf("TestCaseStatus = %q", model.TestCaseStatus.ValueString())
	}
	if model.EntityFQN.ValueString() != entityFQN {
		t.Fatalf("EntityFQN = %q", model.EntityFQN.ValueString())
	}

	var parameters []testCaseParameterValueModel
	if diags := model.ParameterValues.ElementsAs(context.Background(), &parameters, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	if len(parameters) != 1 || parameters[0].Name.ValueString() != parameterName {
		t.Fatalf("ParameterValues = %#v", parameters)
	}
}

func TestApplyKeepsMissingFieldsNull(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model testCaseModel
	apply(&ometa.TestCase{Id: &id, Name: "bare", EntityLink: "<#E::table::svc.db.users>"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.EntityFQN.IsNull() {
		t.Fatalf("EntityFQN = %v, want null", model.EntityFQN)
	}
	if !model.TestCaseStatus.IsNull() {
		t.Fatalf("TestCaseStatus = %v, want null", model.TestCaseStatus)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
	if !model.ParameterValues.IsNull() {
		t.Fatalf("ParameterValues = %v, want null", model.ParameterValues)
	}
	if !model.DimensionColumns.IsNull() {
		t.Fatalf("DimensionColumns = %v, want null", model.DimensionColumns)
	}
}
