package pipeline_service

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// pipelineServiceModel is the Terraform representation of an OpenMetadata
// pipeline service. The connector configuration is kept as a JSON object
// string, because its shape depends on the service type.
type pipelineServiceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	ServiceType        types.String  `tfsdk:"service_type"`
	Connection         types.String  `tfsdk:"connection_config"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(service *ometa.PipelineService, model *pipelineServiceModel) {
	model.ID = types.StringValue(service.Id.String())
	model.Name = types.StringValue(service.Name)
	model.FullyQualifiedName = convert.StringValue(service.FullyQualifiedName)
	model.DisplayName = convert.StringValue(service.DisplayName)
	model.Description = convert.StringValueOr(service.Description, "")
	model.ServiceType = types.StringValue(string(service.ServiceType))
	model.Connection = connectionValue(service.Connection, model.Connection)
	model.Version = convert.Float64Value(service.Version)
	model.UpdatedAt = convert.TimestampValue(service.UpdatedAt)
	model.UpdatedBy = convert.StringValue(service.UpdatedBy)
}

// connectionValue renders the connector configuration. When the existing value
// is known and semantically equal it is kept verbatim, so formatting coming
// from the configuration does not churn state. A missing configuration never
// zeroes a value that is already in state.
func connectionValue(connection *ometa.PipelineConnection, existing types.String) types.String {
	if connection == nil || connection.Config == nil {
		if existing.IsUnknown() {
			return types.StringNull()
		}
		return existing
	}

	encoded, err := json.Marshal(connection.Config)
	if err != nil {
		if existing.IsUnknown() {
			return types.StringNull()
		}
		return existing
	}

	value := string(encoded)
	if !existing.IsUnknown() && !existing.IsNull() && jsonEqual(existing.ValueString(), value) {
		return existing
	}
	return types.StringValue(value)
}

// parseConnectionJSON decodes the connection attribute into a generic JSON
// value for an API payload.
func parseConnectionJSON(value string) (any, error) {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

// jsonEqual reports whether two JSON documents are semantically equal,
// ignoring key order and whitespace.
func jsonEqual(left, right string) bool {
	var decodedLeft any
	if err := json.Unmarshal([]byte(left), &decodedLeft); err != nil {
		return false
	}
	var decodedRight any
	if err := json.Unmarshal([]byte(right), &decodedRight); err != nil {
		return false
	}
	return reflect.DeepEqual(decodedLeft, decodedRight)
}

// connectionConfigured reports whether the attribute holds a value to send.
func connectionConfigured(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && strings.TrimSpace(value.ValueString()) != ""
}
