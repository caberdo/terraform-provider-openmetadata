package api_endpoint

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// apiEndpointModel is the Terraform representation of an OpenMetadata API
// endpoint. An endpoint lives inside an API collection and inherits its
// service.
type apiEndpointModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	APICollection      types.String  `tfsdk:"api_collection"`
	Service            types.String  `tfsdk:"service"`
	ServiceType        types.String  `tfsdk:"service_type"`
	EndpointURL        types.String  `tfsdk:"endpoint_url"`
	RequestMethod      types.String  `tfsdk:"request_method"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The collection,
// service and service type selectors are left untouched when the server omits
// the reference, so a lookup without those fields never zeroes state.
func apply(endpoint *ometa.APIEndpoint, model *apiEndpointModel) {
	model.ID = types.StringValue(endpoint.Id.String())
	model.Name = types.StringValue(endpoint.Name)
	model.FullyQualifiedName = convert.StringValue(endpoint.FullyQualifiedName)
	model.DisplayName = convert.StringValue(endpoint.DisplayName)
	model.Description = convert.StringValueOr(endpoint.Description, "")
	model.EndpointURL = convert.StringValueOr(endpoint.EndpointURL, "")
	if endpoint.RequestMethod != nil && *endpoint.RequestMethod != "" {
		model.RequestMethod = types.StringValue(string(*endpoint.RequestMethod))
	}
	if endpoint.ApiCollection != nil {
		if fqn := endpoint.ApiCollection.FullyQualifiedName; fqn != nil && *fqn != "" {
			model.APICollection = types.StringValue(*fqn)
		} else if name := endpoint.ApiCollection.Name; name != nil && *name != "" {
			model.APICollection = types.StringValue(*name)
		}
	}
	if fqn := endpoint.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := endpoint.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	if endpoint.ServiceType != nil && *endpoint.ServiceType != "" {
		model.ServiceType = types.StringValue(string(*endpoint.ServiceType))
	}
	model.Version = convert.Float64Value(endpoint.Version)
	model.UpdatedAt = convert.TimestampValue(endpoint.UpdatedAt)
	model.UpdatedBy = convert.StringValue(endpoint.UpdatedBy)
}
