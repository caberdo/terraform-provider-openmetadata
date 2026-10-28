package llm_model

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

	if resp.TypeName != "openmetadata_llm_model" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_llm_model")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "service", "base_model"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !required.Required {
			t.Fatalf("%s must be required", attribute)
		}
	}

	for _, attribute := range []string{"name", "service"} {
		required := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if len(required.PlanModifiers) == 0 {
			t.Fatalf("%s must carry a plan modifier", attribute)
		}
	}

	for _, attribute := range []string{"model_provider", "model_version"} {
		optional, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !optional.Optional || !optional.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
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
	if !ok || !description.Optional || !description.Computed || description.Default == nil {
		t.Fatal("description must be optional and computed with an empty default")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "llm-service.my-model"
	displayName := "Support model"
	description := "Answers support questions."
	modelProvider := "openai"
	modelVersion := "2024-08-06"
	serviceName := "llm-service"
	serviceFQN := "llm-service"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model llmModelModel
	apply(&ometa.LLMModel{
		Id:                 id,
		Name:               "my-model",
		BaseModel:          "gpt-4o",
		ModelProvider:      &modelProvider,
		ModelVersion:       &modelVersion,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Service: &ometa.EntityReference{
			Name:               &serviceName,
			FullyQualifiedName: &serviceFQN,
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "my-model" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.BaseModel.ValueString() != "gpt-4o" {
		t.Fatalf("BaseModel = %q", model.BaseModel.ValueString())
	}
	if model.ModelProvider.ValueString() != modelProvider {
		t.Fatalf("ModelProvider = %q", model.ModelProvider.ValueString())
	}
	if model.ModelVersion.ValueString() != modelVersion {
		t.Fatalf("ModelVersion = %q", model.ModelVersion.ValueString())
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
}

func TestApplyServiceFallsBackToName(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "llm-service"

	var model llmModelModel
	apply(&ometa.LLMModel{
		Id:        id,
		Name:      "my-model",
		BaseModel: "gpt-4o",
		Service:   &ometa.EntityReference{Name: &serviceName},
	}, &model)

	if model.Service.ValueString() != serviceName {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
}

func TestApplyMissingOptionalKeepsPrior(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := llmModelModel{
		Service:       types.StringValue("llm-service"),
		ModelProvider: types.StringValue("openai"),
		ModelVersion:  types.StringValue("2024-08-06"),
	}
	apply(&ometa.LLMModel{Id: id, Name: "my-model", BaseModel: "gpt-4o"}, &model)

	if model.Service.ValueString() != "llm-service" {
		t.Fatalf("Service = %q, want the prior value", model.Service.ValueString())
	}
	if model.ModelProvider.ValueString() != "openai" {
		t.Fatalf("ModelProvider = %q, want the prior value", model.ModelProvider.ValueString())
	}
	if model.ModelVersion.ValueString() != "2024-08-06" {
		t.Fatalf("ModelVersion = %q, want the prior value", model.ModelVersion.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := llmModelModel{
		Name:          types.StringValue("my-model"),
		Service:       types.StringValue("llm-service"),
		BaseModel:     types.StringValue("gpt-4o"),
		ModelProvider: types.StringValue("openai"),
		ModelVersion:  types.StringValue("2024-08-06"),
		DisplayName:   types.StringValue("Display"),
		Description:   types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      llmModelModel
		state     llmModelModel
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
			name: "name and service are not patched",
			plan: llmModelModel{
				Name:          types.StringValue("other"),
				Service:       types.StringValue("other-service"),
				BaseModel:     base.BaseModel,
				ModelProvider: base.ModelProvider,
				ModelVersion:  base.ModelVersion,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name changed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     base.BaseModel,
				ModelProvider: base.ModelProvider,
				ModelVersion:  base.ModelVersion,
				DisplayName:   types.StringValue("New display"),
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     base.BaseModel,
				ModelProvider: base.ModelProvider,
				ModelVersion:  base.ModelVersion,
				DisplayName:   types.StringNull(),
				Description:   base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     base.BaseModel,
				ModelProvider: base.ModelProvider,
				ModelVersion:  base.ModelVersion,
				DisplayName:   base.DisplayName,
				Description:   types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "base model changed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     types.StringValue("gpt-4o-mini"),
				ModelProvider: base.ModelProvider,
				ModelVersion:  base.ModelVersion,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/baseModel",
			wantOp:    "add",
			wantValue: "gpt-4o-mini",
		},
		{
			name: "model provider changed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     base.BaseModel,
				ModelProvider: types.StringValue("azure"),
				ModelVersion:  base.ModelVersion,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/modelProvider",
			wantOp:    "add",
			wantValue: "azure",
		},
		{
			name: "model provider removed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     base.BaseModel,
				ModelProvider: types.StringNull(),
				ModelVersion:  base.ModelVersion,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/modelProvider",
			wantOp:   "remove",
		},
		{
			name: "model version changed",
			plan: llmModelModel{
				Name:          base.Name,
				Service:       base.Service,
				BaseModel:     base.BaseModel,
				ModelProvider: base.ModelProvider,
				ModelVersion:  types.StringValue("2025-01-01"),
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/modelVersion",
			wantOp:    "add",
			wantValue: "2025-01-01",
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
