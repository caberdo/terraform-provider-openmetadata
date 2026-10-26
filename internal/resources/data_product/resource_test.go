package data_product

import (
	"context"
	"reflect"
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

	if resp.TypeName != "openmetadata_data_product" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_data_product")
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

	domains, ok := resp.Schema.Attributes["domains"].(rschema.ListAttribute)
	if !ok || !domains.Required {
		t.Fatal("domains must be a required list")
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
	fqn := "CustomerInsights"
	displayName := "Customer Insights"
	domain := "Marketing"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)

	var model dataProductModel
	apply(&ometa.DataProduct{
		Id:                 id,
		Name:               "CustomerInsights",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        "Curated customer analytics.",
		Domains: &[]ometa.EntityReference{{
			FullyQualifiedName: &domain,
			Type:               "domain",
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "CustomerInsights" {
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
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyKeepsDomainsWhenReferencesAreMissing(t *testing.T) {
	domains := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Marketing")})
	model := dataProductModel{Domains: domains}

	apply(&ometa.DataProduct{Id: uuid.New(), Name: "bare"}, &model)

	if !model.Domains.Equal(domains) {
		t.Fatalf("Domains = %v, want the prior state", model.Domains)
	}
}

func TestPatchOps(t *testing.T) {
	base := dataProductModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Domains:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Marketing")}),
	}

	tests := []struct {
		name      string
		plan      dataProductModel
		state     dataProductModel
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
			plan: dataProductModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Domains:     base.Domains,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "description changed",
			plan: dataProductModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Domains:     base.Domains,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "domains changed",
			plan: dataProductModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Domains:     types.ListValueMust(types.StringType, []attr.Value{types.StringValue("Sales")}),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/domains",
			wantOp:    "add",
			wantValue: []string{"Sales"},
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
			if tt.wantValue != nil && !reflect.DeepEqual(ops[0].Value, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
