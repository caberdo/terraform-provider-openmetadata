package test_definition

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

	if resp.TypeName != "openmetadata_test_definition" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_test_definition")
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
		"entity_type",
		"test_platforms",
		"data_quality_dimension",
		"supported_data_types",
		"supported_services",
		"sql_expression",
		"source",
		"updated_at",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "custom_table_row_count"
	displayName := "Custom table row count"
	description := "Counts rows in a table."
	sqlExpression := "SELECT COUNT(*) FROM {{ table }}"
	entityType := ometa.TestDefinitionEntityTypeTABLE
	dimension := ometa.TestDefinitionDataQualityDimensionAccuracy
	provider := ometa.TestDefinitionProviderUser
	dataTypes := []ometa.TestDefinitionSupportedDataTypes{ometa.TestDefinitionSupportedDataTypesINT}
	services := []string{"Mysql"}

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
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "custom_table_row_count" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Counts rows in a table." {
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

	var platforms []string
	if diags := model.TestPlatforms.ElementsAs(context.Background(), &platforms, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	if len(platforms) != 1 || platforms[0] != "Other" {
		t.Fatalf("TestPlatforms = %v", platforms)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model testDefinitionModel
	apply(&ometa.TestDefinition{Id: &id, Name: "bare"}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty fallback", model.Description.ValueString())
	}
	if !model.Provider.IsNull() {
		t.Fatalf("Provider = %v, want null", model.Provider)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.SupportedDataTypes.IsNull() {
		t.Fatalf("SupportedDataTypes = %v, want null", model.SupportedDataTypes)
	}
}
