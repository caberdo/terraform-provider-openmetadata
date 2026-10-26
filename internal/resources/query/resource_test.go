package query

import (
	"context"
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

	if resp.TypeName != "openmetadata_query" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_query")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"service", "query"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want StringAttribute", resp.Schema.Attributes["name"])
	}
	if !name.Optional || !name.Computed {
		t.Fatal("name must be optional and computed")
	}

	queryType, ok := resp.Schema.Attributes["query_type"].(rschema.StringAttribute)
	if !ok || !queryType.Optional || queryType.Computed {
		t.Fatal("query_type must be an optional string")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed || description.Default == nil {
		t.Fatal("description must be optional, computed, and default to an empty string")
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "sample_data.ecommerce_db.select_orders"
	name := "select_orders"
	displayName := "Select orders"
	queryType := "SELECT"
	version := 0.3
	updatedAt := int64(1700000000000)
	updatedBy := "admin"
	serviceFQN := "sample_data.ecommerce_db"
	description := "Orders placed in the last day."

	var model queryModel
	apply(&ometa.Query{
		Id:                 id,
		Name:               name,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Query:              "SELECT * FROM orders",
		QueryType:          &queryType,
		Service: ometa.EntityReference{
			FullyQualifiedName: &serviceFQN,
			Type:               "databaseService",
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
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
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Orders placed in the last day." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Query.ValueString() != "SELECT * FROM orders" {
		t.Fatalf("Query = %q", model.Query.ValueString())
	}
	if model.QueryType.ValueString() != queryType {
		t.Fatalf("QueryType = %q", model.QueryType.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
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

func TestApplyKeepsPriorServiceWhenOmitted(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := queryModel{Service: types.StringValue("sample_data.ecommerce_db")}
	apply(&ometa.Query{Id: id, Name: "bare", Query: "SELECT 1"}, &model)

	if model.Service.ValueString() != "sample_data.ecommerce_db" {
		t.Fatalf("Service = %q, want the prior value", model.Service.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.QueryType.IsNull() {
		t.Fatalf("QueryType = %v, want null", model.QueryType)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := queryModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Query:       types.StringValue("SELECT 1"),
	}

	tests := []struct {
		name      string
		plan      queryModel
		state     queryModel
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
			plan: queryModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Query:       base.Query,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: queryModel{
				DisplayName: types.StringNull(),
				Description: base.Description,
				Query:       base.Query,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: queryModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Query:       base.Query,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "query text changed",
			plan: queryModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Query:       types.StringValue("SELECT 2"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/query",
			wantOp:    "add",
			wantValue: "SELECT 2",
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
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
