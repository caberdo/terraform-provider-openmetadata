package prompt_template

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

	if resp.TypeName != "openmetadata_prompt_template" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_prompt_template")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want StringAttribute", resp.Schema.Attributes["name"])
	}
	if !name.Required {
		t.Fatal("name must be required")
	}

	content, ok := resp.Schema.Attributes["template_content"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("template_content attribute is %T, want StringAttribute", resp.Schema.Attributes["template_content"])
	}
	if !content.Required {
		t.Fatal("template_content must be required")
	}

	templateType, ok := resp.Schema.Attributes["template_type"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("template_type attribute is %T, want StringAttribute", resp.Schema.Attributes["template_type"])
	}
	if !templateType.Computed || !templateType.Optional {
		t.Fatal("template_type must be optional and computed")
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
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "ticket-classifier"
	displayName := "Ticket classifier"
	description := "Classifies support tickets."
	templateType := ometa.PromptTemplateTemplateTypeClassification
	templateVersion := "1.2"
	systemPrompt := "You are a support agent."
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model promptTemplateModel
	apply(&ometa.PromptTemplate{
		Id:                 id,
		Name:               "ticket-classifier",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		TemplateContent:    "Classify the ticket: {{ticket}}",
		TemplateType:       &templateType,
		TemplateVersion:    &templateVersion,
		SystemPrompt:       &systemPrompt,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "ticket-classifier" {
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
	if model.TemplateContent.ValueString() != "Classify the ticket: {{ticket}}" {
		t.Fatalf("TemplateContent = %q", model.TemplateContent.ValueString())
	}
	if model.TemplateType.ValueString() != string(templateType) {
		t.Fatalf("TemplateType = %q", model.TemplateType.ValueString())
	}
	if model.TemplateVersion.ValueString() != templateVersion {
		t.Fatalf("TemplateVersion = %q", model.TemplateVersion.ValueString())
	}
	if model.SystemPrompt.ValueString() != systemPrompt {
		t.Fatalf("SystemPrompt = %q", model.SystemPrompt.ValueString())
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

	var model promptTemplateModel
	apply(&ometa.PromptTemplate{Id: id, Name: "bare", TemplateContent: "body"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.TemplateType.IsNull() {
		t.Fatalf("TemplateType = %v, want null", model.TemplateType)
	}
	if !model.TemplateVersion.IsNull() {
		t.Fatalf("TemplateVersion = %v, want null", model.TemplateVersion)
	}
	if model.SystemPrompt.ValueString() != "" {
		t.Fatalf("SystemPrompt = %q, want the empty default", model.SystemPrompt.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := promptTemplateModel{
		DisplayName:     types.StringValue("Display"),
		Description:     types.StringValue("before"),
		TemplateContent: types.StringValue("Classify the ticket: {{ticket}}"),
		TemplateType:    types.StringValue("Classification"),
		TemplateVersion: types.StringValue("1"),
		SystemPrompt:    types.StringValue(""),
	}

	tests := []struct {
		name      string
		plan      promptTemplateModel
		state     promptTemplateModel
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
			name: "display name added",
			plan: func() promptTemplateModel {
				model := base
				model.DisplayName = types.StringValue("New display")
				return model
			}(),
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: func() promptTemplateModel {
				model := base
				model.DisplayName = types.StringNull()
				return model
			}(),
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: func() promptTemplateModel {
				model := base
				model.Description = types.StringValue("after")
				return model
			}(),
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "template content changed",
			plan: func() promptTemplateModel {
				model := base
				model.TemplateContent = types.StringValue("Summarise the ticket: {{ticket}}")
				return model
			}(),
			state:     base,
			wantOps:   1,
			wantPath:  "/templateContent",
			wantOp:    "add",
			wantValue: "Summarise the ticket: {{ticket}}",
		},
		{
			name: "template type removed",
			plan: func() promptTemplateModel {
				model := base
				model.TemplateType = types.StringNull()
				return model
			}(),
			state:    base,
			wantOps:  1,
			wantPath: "/templateType",
			wantOp:   "remove",
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
