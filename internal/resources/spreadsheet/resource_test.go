package spreadsheet

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

	if resp.TypeName != "openmetadata_spreadsheet" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_spreadsheet")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "service", "parent"} {
		attr, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if attribute == "parent" {
			if !attr.Optional || attr.Required {
				t.Fatalf("%s must be optional", attribute)
			}
		} else if !attr.Required {
			t.Fatalf("%s must be required", attribute)
		}
		if len(attr.PlanModifiers) == 0 {
			t.Fatalf("%s must carry a plan modifier", attribute)
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
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	directoryID := uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "google-drive"
	serviceFQN := "google-drive"
	directoryName := "reports"
	directoryFQN := "google-drive.reports"
	displayName := "Quarterly report"
	fqn := "google-drive.reports.quarterly"
	description := "Quarterly revenue workbook."
	version := 0.3
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model spreadsheetModel
	apply(&ometa.Spreadsheet{
		Id:                 id,
		Name:               "quarterly",
		Service:            ometa.EntityReference{Name: &serviceName, FullyQualifiedName: &serviceFQN},
		Directory:          &ometa.EntityReference{Id: directoryID, Name: &directoryName, FullyQualifiedName: &directoryFQN, Type: "directory"},
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "quarterly" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.Parent.ValueString() != directoryFQN {
		t.Fatalf("Parent = %q", model.Parent.ValueString())
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
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "google-drive"

	model := spreadsheetModel{
		Service: types.StringValue("google-drive"),
		Parent:  types.StringValue("google-drive.reports"),
	}
	apply(&ometa.Spreadsheet{
		Id:      id,
		Name:    "bare",
		Service: ometa.EntityReference{Name: &serviceName},
	}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.Service.ValueString() != "google-drive" {
		t.Fatalf("Service = %q, want the name fallback", model.Service.ValueString())
	}
	if model.Parent.ValueString() != "google-drive.reports" {
		t.Fatalf("Parent = %q, want the prior value preserved", model.Parent.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := spreadsheetModel{
		Name:        types.StringValue("quarterly"),
		Service:     types.StringValue("google-drive"),
		Parent:      types.StringValue("google-drive.reports"),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      spreadsheetModel
		state     spreadsheetModel
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
			plan: spreadsheetModel{
				Name:        base.Name,
				Service:     base.Service,
				Parent:      base.Parent,
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: spreadsheetModel{
				Name:        base.Name,
				Service:     base.Service,
				Parent:      base.Parent,
				DisplayName: types.StringNull(),
				Description: base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: spreadsheetModel{
				Name:        base.Name,
				Service:     base.Service,
				Parent:      base.Parent,
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "name service and parent cannot be patched",
			plan: spreadsheetModel{
				Name:        types.StringValue("renamed"),
				Service:     types.StringValue("other-drive"),
				Parent:      types.StringValue("other-drive.reports"),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state:   base,
			wantOps: 0,
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
