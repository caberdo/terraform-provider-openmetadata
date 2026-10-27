package notification_template

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

	if resp.TypeName != "openmetadata_notification_template" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_notification_template")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "template_subject", "template_body"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "version", "updated_at", "updated_by"} {
		if attribute == "version" {
			version, ok := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !ok || !version.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	modified, ok := resp.Schema.Attributes["is_modified_from_default"].(rschema.BoolAttribute)
	if !ok || !modified.Computed {
		t.Fatal("is_modified_from_default must be a computed bool")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "asset-created"
	displayName := "Asset created"
	description := "Sent when an asset is created."
	modified := true
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model notificationTemplateModel
	apply(&ometa.NotificationTemplate{
		Id:                    id,
		Name:                  "asset-created",
		FullyQualifiedName:    &fqn,
		DisplayName:           &displayName,
		Description:           &description,
		TemplateSubject:       "Asset {{ entity.name }} created",
		TemplateBody:          "{{ entity.name }} was created by {{ user.name }}.",
		IsModifiedFromDefault: &modified,
		Version:               &version,
		UpdatedAt:             &updatedAt,
		UpdatedBy:             &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "asset-created" {
		t.Fatalf("Name = %q", model.Name.ValueString())
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
	if model.TemplateSubject.ValueString() != "Asset {{ entity.name }} created" {
		t.Fatalf("TemplateSubject = %q", model.TemplateSubject.ValueString())
	}
	if model.TemplateBody.ValueString() != "{{ entity.name }} was created by {{ user.name }}." {
		t.Fatalf("TemplateBody = %q", model.TemplateBody.ValueString())
	}
	if !model.IsModifiedFromDefault.ValueBool() {
		t.Fatal("IsModifiedFromDefault = false")
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

	var model notificationTemplateModel
	apply(&ometa.NotificationTemplate{Id: id, Name: "bare"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.IsModifiedFromDefault.ValueBool() {
		t.Fatal("IsModifiedFromDefault = true, want the false default")
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := notificationTemplateModel{
		DisplayName:     types.StringValue("Display"),
		Description:     types.StringValue("before"),
		TemplateSubject: types.StringValue("Subject before"),
		TemplateBody:    types.StringValue("Body before"),
	}

	tests := []struct {
		name      string
		plan      notificationTemplateModel
		state     notificationTemplateModel
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
			plan: notificationTemplateModel{
				DisplayName:     types.StringValue("New display"),
				Description:     base.Description,
				TemplateSubject: base.TemplateSubject,
				TemplateBody:    base.TemplateBody,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: notificationTemplateModel{
				DisplayName:     types.StringNull(),
				Description:     base.Description,
				TemplateSubject: base.TemplateSubject,
				TemplateBody:    base.TemplateBody,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: notificationTemplateModel{
				DisplayName:     base.DisplayName,
				Description:     types.StringValue("after"),
				TemplateSubject: base.TemplateSubject,
				TemplateBody:    base.TemplateBody,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "template subject changed",
			plan: notificationTemplateModel{
				DisplayName:     base.DisplayName,
				Description:     base.Description,
				TemplateSubject: types.StringValue("Subject after"),
				TemplateBody:    base.TemplateBody,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/templateSubject",
			wantOp:    "add",
			wantValue: "Subject after",
		},
		{
			name: "template body changed",
			plan: notificationTemplateModel{
				DisplayName:     base.DisplayName,
				Description:     base.Description,
				TemplateSubject: base.TemplateSubject,
				TemplateBody:    types.StringValue("Body after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/templateBody",
			wantOp:    "add",
			wantValue: "Body after",
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
