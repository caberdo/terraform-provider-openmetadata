package search_index

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

	if resp.TypeName != "openmetadata_search_index" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_search_index")
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
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if len(selector.Validators) == 0 {
			t.Fatalf("%s must carry an ExactlyOneOf validator", attribute)
		}
	}

	for _, attribute := range []string{
		"name",
		"service",
		"display_name",
		"description",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed || computed.Optional {
			t.Fatalf("%s must be a computed-only string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(dschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}

	fields, ok := resp.Schema.Attributes["fields"].(dschema.ListNestedAttribute)
	if !ok || !fields.Computed || fields.Optional || fields.Required {
		t.Fatal("fields must be a computed-only list")
	}

	for _, attribute := range []string{"name", "data_type", "description", "display_name"} {
		computed, ok := fields.NestedObject.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed || computed.Optional || computed.Required {
			t.Fatalf("fields.%s must be a computed-only string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "elastic"
	fqn := "elastic.messages"
	displayName := "Messages"
	description := "Message search index."

	var model searchIndexModel
	apply(&ometa.SearchIndex{
		Id:                 id,
		Name:               "messages",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			Name: &serviceName,
			Type: "searchService",
		},
		Fields: []ometa.SearchIndexField{
			{Name: "message", DataType: ometa.SearchIndexFieldDataTypeTEXT},
		},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "messages" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Service.ValueString() != serviceName {
		t.Fatalf("Service = %q", model.Service.ValueString())
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

	if model.Fields.IsNull() || model.Fields.IsUnknown() {
		t.Fatal("Fields must be a known list")
	}
	fields := model.Fields.Elements()
	if len(fields) != 1 {
		t.Fatalf("len(Fields) = %d, want 1", len(fields))
	}

	var fieldModels []searchIndexFieldModel
	if diags := model.Fields.ElementsAs(context.Background(), &fieldModels, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	if fieldModels[0].Name.ValueString() != "message" {
		t.Fatalf("Fields[0].Name = %q", fieldModels[0].Name.ValueString())
	}
	if fieldModels[0].DataType.ValueString() != "TEXT" {
		t.Fatalf("Fields[0].DataType = %q", fieldModels[0].DataType.ValueString())
	}
}
