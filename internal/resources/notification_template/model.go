package notification_template

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// notificationTemplateModel is the Terraform representation of an OpenMetadata
// notification template. Subject and body are Jinja templates rendered by the
// server when an alert fires.
type notificationTemplateModel struct {
	ID                    types.String  `tfsdk:"id"`
	Name                  types.String  `tfsdk:"name"`
	FullyQualifiedName    types.String  `tfsdk:"fully_qualified_name"`
	DisplayName           types.String  `tfsdk:"display_name"`
	Description           types.String  `tfsdk:"description"`
	TemplateSubject       types.String  `tfsdk:"template_subject"`
	TemplateBody          types.String  `tfsdk:"template_body"`
	IsModifiedFromDefault types.Bool    `tfsdk:"is_modified_from_default"`
	Version               types.Float64 `tfsdk:"version"`
	UpdatedAt             types.String  `tfsdk:"updated_at"`
	UpdatedBy             types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(template *ometa.NotificationTemplate, model *notificationTemplateModel) {
	model.ID = types.StringValue(template.Id.String())
	model.Name = types.StringValue(template.Name)
	model.FullyQualifiedName = convert.StringValue(template.FullyQualifiedName)
	model.DisplayName = convert.StringValue(template.DisplayName)
	model.Description = convert.StringValueOr(template.Description, "")
	model.TemplateSubject = types.StringValue(template.TemplateSubject)
	model.TemplateBody = types.StringValue(template.TemplateBody)
	model.IsModifiedFromDefault = convert.BoolValueOr(template.IsModifiedFromDefault, false)
	model.Version = convert.Float64Value(template.Version)
	model.UpdatedAt = convert.TimestampValue(template.UpdatedAt)
	model.UpdatedBy = convert.StringValue(template.UpdatedBy)
}
