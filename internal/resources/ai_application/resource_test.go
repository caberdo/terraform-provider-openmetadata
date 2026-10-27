package ai_application

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

	if resp.TypeName != "openmetadata_ai_application" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ai_application")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	applicationType, ok := resp.Schema.Attributes["application_type"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("application_type attribute is %T, want StringAttribute", resp.Schema.Attributes["application_type"])
	}
	if !applicationType.Required {
		t.Fatal("application_type must be required")
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatal("name must be a required string")
	}

	for _, attribute := range []string{"deployment_url", "documentation", "source_code", "description"} {
		optional, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !optional.Optional || !optional.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if optional.Default == nil {
			t.Fatalf("%s must carry a default", attribute)
		}
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
	fqn := "support-copilot"
	displayName := "Support Copilot"
	deploymentURL := "https://copilot.example.com"
	developmentStage := ometa.AIApplicationDevelopmentStageProduction
	documentation := "https://docs.example.com/support-copilot"
	sourceCode := "https://github.com/example/support-copilot"
	description := "Answers support questions."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model aiApplicationModel
	apply(&ometa.AIApplication{
		Id:                 id,
		Name:               "support-copilot",
		ApplicationType:    ometa.AIApplicationApplicationTypeCopilot,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		DeploymentUrl:      &deploymentURL,
		DevelopmentStage:   &developmentStage,
		Documentation:      &documentation,
		SourceCode:         &sourceCode,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "support-copilot" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.ApplicationType.ValueString() != "Copilot" {
		t.Fatalf("ApplicationType = %q", model.ApplicationType.ValueString())
	}
	if model.DeploymentURL.ValueString() != deploymentURL {
		t.Fatalf("DeploymentURL = %q", model.DeploymentURL.ValueString())
	}
	if model.DevelopmentStage.ValueString() != "Production" {
		t.Fatalf("DevelopmentStage = %q", model.DevelopmentStage.ValueString())
	}
	if model.Documentation.ValueString() != documentation {
		t.Fatalf("Documentation = %q", model.Documentation.ValueString())
	}
	if model.SourceCode.ValueString() != sourceCode {
		t.Fatalf("SourceCode = %q", model.SourceCode.ValueString())
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

	var model aiApplicationModel
	apply(&ometa.AIApplication{
		Id:              id,
		Name:            "bare",
		ApplicationType: ometa.AIApplicationApplicationTypeChatbot,
	}, &model)

	for attribute, value := range map[string]types.String{
		"deployment_url": model.DeploymentURL,
		"documentation":  model.Documentation,
		"source_code":    model.SourceCode,
		"description":    model.Description,
	} {
		if value.ValueString() != "" {
			t.Fatalf("%s = %q, want the empty default", attribute, value.ValueString())
		}
	}
	if !model.DevelopmentStage.IsNull() {
		t.Fatalf("DevelopmentStage = %v, want null", model.DevelopmentStage)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := aiApplicationModel{
		ApplicationType:  types.StringValue("Chatbot"),
		DisplayName:      types.StringValue("Display"),
		Description:      types.StringValue("before"),
		DeploymentURL:    types.StringValue("https://old.example.com"),
		DevelopmentStage: types.StringValue("Staging"),
		Documentation:    types.StringValue(""),
		SourceCode:       types.StringValue(""),
	}

	tests := []struct {
		name      string
		plan      aiApplicationModel
		state     aiApplicationModel
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
			plan: aiApplicationModel{
				ApplicationType:  base.ApplicationType,
				DisplayName:      types.StringValue("New display"),
				Description:      base.Description,
				DeploymentURL:    base.DeploymentURL,
				DevelopmentStage: base.DevelopmentStage,
				Documentation:    base.Documentation,
				SourceCode:       base.SourceCode,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: aiApplicationModel{
				ApplicationType:  base.ApplicationType,
				DisplayName:      types.StringNull(),
				Description:      base.Description,
				DeploymentURL:    base.DeploymentURL,
				DevelopmentStage: base.DevelopmentStage,
				Documentation:    base.Documentation,
				SourceCode:       base.SourceCode,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: aiApplicationModel{
				ApplicationType:  base.ApplicationType,
				DisplayName:      base.DisplayName,
				Description:      types.StringValue("after"),
				DeploymentURL:    base.DeploymentURL,
				DevelopmentStage: base.DevelopmentStage,
				Documentation:    base.Documentation,
				SourceCode:       base.SourceCode,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "deployment url changed",
			plan: aiApplicationModel{
				ApplicationType:  base.ApplicationType,
				DisplayName:      base.DisplayName,
				Description:      base.Description,
				DeploymentURL:    types.StringValue("https://new.example.com"),
				DevelopmentStage: base.DevelopmentStage,
				Documentation:    base.Documentation,
				SourceCode:       base.SourceCode,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/deploymentUrl",
			wantOp:    "add",
			wantValue: "https://new.example.com",
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
