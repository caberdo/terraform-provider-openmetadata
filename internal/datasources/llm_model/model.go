package llm_model

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// llmModelModel is the Terraform representation of an LLM model lookup. The id
// and fully_qualified_name attributes are both selectors and results.
type llmModelModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	BaseModel          types.String  `tfsdk:"base_model"`
	ModelProvider      types.String  `tfsdk:"model_provider"`
	ModelVersion       types.String  `tfsdk:"model_version"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The service
// selector is left untouched when the server omits the reference.
func apply(llmModel *ometa.LLMModel, model *llmModelModel) {
	model.ID = types.StringValue(llmModel.Id.String())
	model.Name = types.StringValue(llmModel.Name)
	model.BaseModel = types.StringValue(llmModel.BaseModel)
	if llmModel.ModelProvider != nil {
		model.ModelProvider = types.StringValue(*llmModel.ModelProvider)
	} else if model.ModelProvider.IsUnknown() {
		model.ModelProvider = types.StringNull()
	}
	if llmModel.ModelVersion != nil {
		model.ModelVersion = types.StringValue(*llmModel.ModelVersion)
	} else if model.ModelVersion.IsUnknown() {
		model.ModelVersion = types.StringNull()
	}
	model.Service = types.StringNull()
	if llmModel.Service != nil {
		if fqn := llmModel.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
			model.Service = types.StringValue(*fqn)
		} else if name := llmModel.Service.Name; name != nil && *name != "" {
			model.Service = types.StringValue(*name)
		}
	}
	model.FullyQualifiedName = convert.StringValue(llmModel.FullyQualifiedName)
	model.DisplayName = convert.StringValue(llmModel.DisplayName)
	model.Description = convert.StringValueOr(llmModel.Description, "")
	model.Version = convert.Float64Value(llmModel.Version)
	model.UpdatedAt = convert.TimestampValue(llmModel.UpdatedAt)
	model.UpdatedBy = convert.StringValue(llmModel.UpdatedBy)
}
