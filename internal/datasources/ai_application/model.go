package ai_application

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// aiApplicationModel is the Terraform representation of an AI application
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type aiApplicationModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	ApplicationType    types.String  `tfsdk:"application_type"`
	DeploymentURL      types.String  `tfsdk:"deployment_url"`
	DevelopmentStage   types.String  `tfsdk:"development_stage"`
	Documentation      types.String  `tfsdk:"documentation"`
	SourceCode         types.String  `tfsdk:"source_code"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. development_stage
// is left untouched when the server omits it.
func apply(application *ometa.AIApplication, model *aiApplicationModel) {
	model.ID = types.StringValue(application.Id.String())
	model.Name = types.StringValue(application.Name)
	model.ApplicationType = types.StringValue(string(application.ApplicationType))
	model.DeploymentURL = convert.StringValueOr(application.DeploymentUrl, "")
	if application.DevelopmentStage != nil && *application.DevelopmentStage != "" {
		model.DevelopmentStage = types.StringValue(string(*application.DevelopmentStage))
	} else if model.DevelopmentStage.IsUnknown() {
		model.DevelopmentStage = types.StringNull()
	}
	model.Documentation = convert.StringValueOr(application.Documentation, "")
	model.SourceCode = convert.StringValueOr(application.SourceCode, "")
	model.FullyQualifiedName = convert.StringValue(application.FullyQualifiedName)
	model.DisplayName = convert.StringValue(application.DisplayName)
	model.Description = convert.StringValueOr(application.Description, "")
	model.Version = convert.Float64Value(application.Version)
	model.UpdatedAt = convert.TimestampValue(application.UpdatedAt)
	model.UpdatedBy = convert.StringValue(application.UpdatedBy)
}
