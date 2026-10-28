package bot

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// botModel is the Terraform representation of an OpenMetadata bot.
type botModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	BotUser            types.String  `tfsdk:"bot_user"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. BotUser keeps
// the planned value when the server omits the reference, so state never loses
// the configured user.
func apply(bot *ometa.Bot, model *botModel) {
	model.ID = types.StringValue(bot.Id.String())
	model.Name = types.StringValue(bot.Name)
	model.FullyQualifiedName = convert.StringValue(bot.FullyQualifiedName)
	model.DisplayName = convert.StringValue(bot.DisplayName)
	model.Description = convert.StringValueOr(bot.Description, "")
	switch {
	case bot.BotUser.FullyQualifiedName != nil:
		model.BotUser = types.StringValue(*bot.BotUser.FullyQualifiedName)
	case bot.BotUser.Name != nil:
		model.BotUser = types.StringValue(*bot.BotUser.Name)
	}
	model.Version = convert.Float64Value(bot.Version)
	model.UpdatedAt = convert.TimestampValue(bot.UpdatedAt)
	model.UpdatedBy = convert.StringValue(bot.UpdatedBy)
}
