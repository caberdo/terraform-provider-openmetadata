package api_service

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// apiServiceModel is the Terraform representation of an API service lookup.
// The id and fully_qualified_name attributes are both selectors and results.
type apiServiceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	ServiceType        types.String  `tfsdk:"service_type"`
	Connection         types.String  `tfsdk:"connection_config"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(service *ometa.ApiService, model *apiServiceModel) {
	model.ID = types.StringValue(service.Id.String())
	model.Name = types.StringValue(service.Name)
	if service.ServiceType != "" {
		model.ServiceType = types.StringValue(string(service.ServiceType))
	}
	model.FullyQualifiedName = convert.StringValue(service.FullyQualifiedName)
	model.DisplayName = convert.StringValue(service.DisplayName)
	model.Description = convert.StringValueOr(service.Description, "")
	model.Connection = connectionValue(service.Connection)
	model.Version = convert.Float64Value(service.Version)
	model.UpdatedAt = convert.TimestampValue(service.UpdatedAt)
	model.UpdatedBy = convert.StringValue(service.UpdatedBy)
}

// connectionValue renders the connector configuration as a compact JSON
// string. A missing configuration becomes null.
func connectionValue(connection *ometa.ApiConnection) types.String {
	if connection == nil || connection.Config == nil {
		return types.StringNull()
	}

	encoded, err := json.Marshal(connection.Config)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(encoded))
}
