package query

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

	if resp.TypeName != "openmetadata_query" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_query")
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
		"service",
		"name",
		"display_name",
		"description",
		"query",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "sample_data.ecommerce_db.select_orders"
	serviceFQN := "sample_data.ecommerce_db"
	description := "Orders placed in the last day."

	var model queryModel
	apply(&ometa.Query{
		Id:                 id,
		Name:               "select_orders",
		FullyQualifiedName: &fqn,
		Description:        &description,
		Query:              "SELECT * FROM orders",
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Type:               "databaseService",
		},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "select_orders" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.Query.ValueString() != "SELECT * FROM orders" {
		t.Fatalf("Query = %q", model.Query.ValueString())
	}
	if model.Description.ValueString() != "Orders placed in the last day." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
}
