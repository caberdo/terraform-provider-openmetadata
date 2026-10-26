package database_schema

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

	if resp.TypeName != "openmetadata_database_schema" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_database_schema")
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
		"database",
		"display_name",
		"description",
		"updated_at",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	for _, selector := range []string{"id", "fully_qualified_name"} {
		attribute, ok := resp.Schema.Attributes[selector].(dschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", selector, resp.Schema.Attributes[selector])
		}
		if !attribute.Optional || !attribute.Computed {
			t.Fatalf("%s must be optional and computed", selector)
		}
		if len(attribute.Validators) == 0 {
			t.Fatalf("%s must carry a mutually exclusive selector validator", selector)
		}
	}

	database, ok := resp.Schema.Attributes["database"].(dschema.StringAttribute)
	if !ok || !database.Computed || database.Optional {
		t.Fatal("database must be computed only")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "local_postgres.sales.public"
	databaseFQN := "local_postgres.sales"
	displayName := "Public schema"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model databaseSchemaModel
	apply(&ometa.DatabaseSchema{
		Id:                 id,
		Name:               "public",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        stringPointer("Default schema"),
		Database: ometa.EntityReference{
			FullyQualifiedName: &databaseFQN,
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
	if model.Name.ValueString() != "public" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Database.ValueString() != databaseFQN {
		t.Fatalf("Database = %q", model.Database.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Default schema" {
		t.Fatalf("Description = %q", model.Description.ValueString())
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

func stringPointer(value string) *string { return &value }
