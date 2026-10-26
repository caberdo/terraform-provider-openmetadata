package data_product

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_data_product" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_data_product")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "name", "description", "domains", "updated_at"} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "CustomerInsights"
	name := "CustomerInsights"
	domain := "Marketing"

	var model dataProductModel
	apply(&ometa.DataProduct{
		Id:                 id,
		Name:               name,
		FullyQualifiedName: &fqn,
		Description:        "Curated customer analytics.",
		Domains: &[]ometa.EntityReference{{
			FullyQualifiedName: &domain,
			Type:               "domain",
		}},
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
	var domains []string
	if diags := model.Domains.ElementsAs(context.Background(), &domains, false); diags.HasError() {
		t.Fatalf("Domains conversion: %v", diags.Errors())
	}
	if !reflect.DeepEqual(domains, []string{domain}) {
		t.Fatalf("Domains = %#v", domains)
	}
}
