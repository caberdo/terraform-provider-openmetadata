package prompt_template

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// promptTemplateModel is the Terraform representation of a prompt template
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type promptTemplateModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	TemplateContent    types.String  `tfsdk:"template_content"`
	TemplateType       types.String  `tfsdk:"template_type"`
	TemplateVersion    types.String  `tfsdk:"template_version"`
	SystemPrompt       types.String  `tfsdk:"system_prompt"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Optional
// fields the server omits keep their current value.
func apply(template *ometa.PromptTemplate, model *promptTemplateModel) {
	model.ID = types.StringValue(template.Id.String())
	model.Name = types.StringValue(template.Name)
	model.FullyQualifiedName = convert.StringValue(template.FullyQualifiedName)
	model.DisplayName = convert.StringValue(template.DisplayName)
	model.Description = convert.StringValue(template.Description)
	model.TemplateContent = types.StringValue(template.TemplateContent)
	if template.TemplateType != nil {
		model.TemplateType = types.StringValue(string(*template.TemplateType))
	} else if model.TemplateType.IsUnknown() {
		model.TemplateType = types.StringNull()
	}
	if template.TemplateVersion != nil {
		model.TemplateVersion = types.StringValue(*template.TemplateVersion)
	} else if model.TemplateVersion.IsUnknown() {
		model.TemplateVersion = types.StringNull()
	}
	model.SystemPrompt = convert.StringValueOr(template.SystemPrompt, "")
	model.Version = convert.Float64Value(template.Version)
	model.UpdatedAt = convert.TimestampValue(template.UpdatedAt)
	model.UpdatedBy = convert.StringValue(template.UpdatedBy)
}
