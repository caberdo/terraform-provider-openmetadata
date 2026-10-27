package event_subscription

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// destinationObjectType describes one nested subscription destination returned
// by the lookup.
var destinationObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"category": types.StringType,
	"type":     types.StringType,
	"enabled":  types.BoolType,
	"config":   types.StringType,
}}

// eventSubscriptionModel is the Terraform representation of an event
// subscription lookup. The id and fully_qualified_name attributes are both
// selectors and results.
type eventSubscriptionModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	AlertType          types.String  `tfsdk:"alert_type"`
	Resources          types.List    `tfsdk:"resources"`
	Destinations       types.List    `tfsdk:"destinations"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	PollInterval       types.Int64   `tfsdk:"poll_interval"`
	Retries            types.Int64   `tfsdk:"retries"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(subscription *ometa.EventSubscription, model *eventSubscriptionModel) {
	model.ID = types.StringValue(subscription.Id.String())
	model.Name = types.StringValue(subscription.Name)
	model.FullyQualifiedName = convert.StringValue(subscription.FullyQualifiedName)
	model.DisplayName = convert.StringValue(subscription.DisplayName)
	model.Description = convert.StringValueOr(subscription.Description, "")
	model.AlertType = types.StringValue(string(subscription.AlertType))
	model.Resources = types.ListNull(types.StringType)
	if subscription.FilteringRules != nil && len(subscription.FilteringRules.Resources) > 0 {
		model.Resources = stringListValue(subscription.FilteringRules.Resources)
	}
	model.Destinations = destinationsValue(subscription.Destinations)
	model.Enabled = convert.BoolValueOr(subscription.Enabled, true)
	model.PollInterval = convert.Int32Value(subscription.PollInterval)
	model.Retries = convert.Int32Value(subscription.Retries)
	model.Version = convert.Float64Value(subscription.Version)
	model.UpdatedAt = convert.TimestampValue(subscription.UpdatedAt)
	model.UpdatedBy = convert.StringValue(subscription.UpdatedBy)
}

// stringListValue maps a string slice to a Terraform list.
func stringListValue(values []string) types.List {
	elements := make([]attr.Value, len(values))
	for index, value := range values {
		elements[index] = types.StringValue(value)
	}
	return types.ListValueMust(types.StringType, elements)
}

// destinationsValue maps API destinations to a Terraform list in the server's
// order.
func destinationsValue(destinations []ometa.SubscriptionDestination) types.List {
	elements := make([]attr.Value, 0, len(destinations))
	for _, destination := range destinations {
		elements = append(elements, types.ObjectValueMust(destinationObjectType.AttrTypes, map[string]attr.Value{
			"category": types.StringValue(string(destination.Category)),
			"type":     types.StringValue(string(destination.Type)),
			"enabled":  convert.BoolValueOr(destination.Enabled, true),
			"config":   types.StringValue(configJSON(destination.Config)),
		}))
	}
	return types.ListValueMust(destinationObjectType, elements)
}

// configJSON renders a destination configuration, using an empty object when
// the server omits it.
func configJSON(config any) string {
	if config == nil {
		return "{}"
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}
