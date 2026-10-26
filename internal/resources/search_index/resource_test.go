package search_index

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

	if resp.TypeName != "openmetadata_search_index" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_search_index")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "service"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !required.Required {
			t.Fatalf("%s must be required", attribute)
		}
		if len(required.PlanModifiers) == 0 {
			t.Fatalf("%s must carry a plan modifier", attribute)
		}
		if len(required.Validators) == 0 {
			t.Fatalf("%s must carry a validator", attribute)
		}
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
		t.Fatal("version must be a computed float")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed {
		t.Fatal("description must be optional and computed")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Required || displayName.Computed {
		t.Fatal("display_name must be optional only")
	}

	fields, ok := resp.Schema.Attributes["fields"].(rschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("fields attribute is %T, want ListNestedAttribute", resp.Schema.Attributes["fields"])
	}
	if !fields.Required {
		t.Fatal("fields must be required")
	}
	if len(fields.Validators) == 0 {
		t.Fatal("fields must carry a size validator")
	}

	for _, attribute := range []string{"name", "data_type"} {
		required, ok := fields.NestedObject.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("fields.%s must be a required string", attribute)
		}
		if len(required.Validators) == 0 {
			t.Fatalf("fields.%s must carry a validator", attribute)
		}
	}

	for _, attribute := range []string{"description", "display_name"} {
		optional, ok := fields.NestedObject.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !optional.Optional || optional.Required || optional.Computed {
			t.Fatalf("fields.%s must be optional only", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "elastic"
	serviceFQN := "elastic"
	fqn := "elastic.messages"
	displayName := "Messages"
	description := "Message search index."
	fieldDescription := "Message body."
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model searchIndexModel
	apply(&ometa.SearchIndex{
		Id:                 id,
		Name:               "messages",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: ometa.EntityReference{
			Name:               &serviceName,
			FullyQualifiedName: &serviceFQN,
			Type:               "searchService",
		},
		Fields: []ometa.SearchIndexField{
			{
				Name:        "message",
				DataType:    ometa.SearchIndexFieldDataTypeTEXT,
				Description: &fieldDescription,
			},
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "messages" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
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
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
	if model.UpdatedBy.ValueString() != updatedBy {
		t.Fatalf("UpdatedBy = %q", model.UpdatedBy.ValueString())
	}

	fields, diags := fieldsToAPI(context.Background(), model.Fields)
	if diags.HasError() {
		t.Fatalf("fieldsToAPI() diagnostics = %v", diags.Errors())
	}
	if len(fields) != 1 {
		t.Fatalf("len(fields) = %d, want 1", len(fields))
	}
	if fields[0].Name != "message" {
		t.Fatalf("fields[0].Name = %q", fields[0].Name)
	}
	if fields[0].DataType != ometa.SearchIndexFieldDataTypeTEXT {
		t.Fatalf("fields[0].DataType = %q", fields[0].DataType)
	}
	if fields[0].Description == nil || *fields[0].Description != fieldDescription {
		t.Fatalf("fields[0].Description = %v", fields[0].Description)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "elastic"

	model := searchIndexModel{Service: types.StringValue("kept")}
	apply(&ometa.SearchIndex{
		Id:      id,
		Name:    "bare",
		Service: ometa.EntityReference{Name: &serviceName, Type: "searchService"},
	}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.FullyQualifiedName.IsNull() {
		t.Fatalf("FullyQualifiedName = %v, want null", model.FullyQualifiedName)
	}
	if model.Service.ValueString() != serviceName {
		t.Fatalf("Service = %q, want the server reference", model.Service.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestApplyKeepsServiceWhenOmitted(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := searchIndexModel{Service: types.StringValue("kept")}
	apply(&ometa.SearchIndex{Id: id, Name: "bare"}, &model)

	if model.Service.ValueString() != "kept" {
		t.Fatalf("Service = %q, want the prior value preserved", model.Service.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := searchIndexModel{
		Name:        types.StringValue("messages"),
		Service:     types.StringValue("elastic"),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Fields: fieldsValue([]ometa.SearchIndexField{
			{Name: "message", DataType: ometa.SearchIndexFieldDataTypeTEXT},
		}),
	}
	changedFields := fieldsValue([]ometa.SearchIndexField{
		{Name: "message", DataType: ometa.SearchIndexFieldDataTypeTEXT},
		{Name: "level", DataType: ometa.SearchIndexFieldDataTypeKEYWORD},
	})

	tests := []struct {
		name      string
		plan      searchIndexModel
		state     searchIndexModel
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
			plan: searchIndexModel{
				Name:        base.Name,
				Service:     base.Service,
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Fields:      base.Fields,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: searchIndexModel{
				Name:        base.Name,
				Service:     base.Service,
				DisplayName: types.StringNull(),
				Description: base.Description,
				Fields:      base.Fields,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: searchIndexModel{
				Name:        base.Name,
				Service:     base.Service,
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Fields:      base.Fields,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "fields changed",
			plan: searchIndexModel{
				Name:        base.Name,
				Service:     base.Service,
				DisplayName: base.DisplayName,
				Description: base.Description,
				Fields:      changedFields,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/fields",
			wantOp:   "add",
		},
		{
			name: "name and service cannot be patched",
			plan: searchIndexModel{
				Name:        types.StringValue("renamed"),
				Service:     types.StringValue("other"),
				DisplayName: base.DisplayName,
				Description: base.Description,
				Fields:      base.Fields,
			},
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

func TestFieldsToAPIIsNeverNil(t *testing.T) {
	fields, diags := fieldsToAPI(context.Background(), types.ListNull(searchIndexFieldObjectType))
	if diags.HasError() {
		t.Fatalf("fieldsToAPI() diagnostics = %v", diags.Errors())
	}
	if fields == nil {
		t.Fatal("fieldsToAPI() = nil, want an empty slice because the API requires the member")
	}
	if len(fields) != 0 {
		t.Fatalf("len(fields) = %d, want 0", len(fields))
	}
}
