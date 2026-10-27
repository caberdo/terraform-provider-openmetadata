package api_collection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// apiCollectionModel is the Terraform representation of an API collection
// lookup. The id and fully_qualified_name attributes are both selectors and
// results.
type apiCollectionModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	ServiceType        types.String  `tfsdk:"service_type"`
	EndpointURL        types.String  `tfsdk:"endpoint_url"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Missing
// service and service type references leave the prior model value untouched.
func apply(collection *ometa.APICollection, model *apiCollectionModel) {
	model.ID = types.StringValue(collection.Id.String())
	model.Name = types.StringValue(collection.Name)
	model.FullyQualifiedName = convert.StringValue(collection.FullyQualifiedName)
	model.DisplayName = convert.StringValue(collection.DisplayName)
	model.Description = convert.StringValueOr(collection.Description, "")
	model.EndpointURL = convert.StringValueOr(collection.EndpointURL, "")
	if fqn := collection.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := collection.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	if collection.ServiceType != nil && *collection.ServiceType != "" {
		model.ServiceType = types.StringValue(string(*collection.ServiceType))
	}
	model.Version = convert.Float64Value(collection.Version)
	model.UpdatedAt = convert.TimestampValue(collection.UpdatedAt)
	model.UpdatedBy = convert.StringValue(collection.UpdatedBy)
}
