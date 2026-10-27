package event_subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// destinationObjectType describes one nested subscription destination. The
// resource and the data source share this shape so a destination read from the
// API can be written straight back into state.
var destinationObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"category": types.StringType,
	"type":     types.StringType,
	"enabled":  types.BoolType,
	"config":   types.StringType,
}}

// destinationModel is the Terraform representation of one destination.
type destinationModel struct {
	Category types.String `tfsdk:"category"`
	Type     types.String `tfsdk:"type"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Config   types.String `tfsdk:"config"`
}

// eventSubscriptionModel is the Terraform representation of an OpenMetadata
// event subscription. Destinations keep their backend configuration as a JSON
// object string, because its shape depends on the destination type.
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

// apply copies the API representation into the Terraform model. Fields the
// server omits keep their previous value instead of being zeroed.
func apply(subscription *ometa.EventSubscription, model *eventSubscriptionModel) {
	model.ID = types.StringValue(subscription.Id.String())
	model.Name = types.StringValue(subscription.Name)
	model.FullyQualifiedName = convert.StringValue(subscription.FullyQualifiedName)
	model.DisplayName = convert.StringValue(subscription.DisplayName)
	model.Description = convert.StringValueOr(subscription.Description, "")

	if subscription.AlertType != "" {
		model.AlertType = types.StringValue(string(subscription.AlertType))
	}

	if subscription.FilteringRules != nil && len(subscription.FilteringRules.Resources) > 0 {
		model.Resources = stringListValue(subscription.FilteringRules.Resources)
	}
	model.Destinations = destinationsValue(subscription.Destinations, model.Destinations)
	model.Enabled = convert.BoolValueOr(subscription.Enabled, true)
	model.PollInterval = optionalInt32Value(subscription.PollInterval, model.PollInterval)
	model.Retries = optionalInt32Value(subscription.Retries, model.Retries)
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

// destinationsValue maps API destinations to a Terraform list. A list whose
// content already matches state is kept verbatim, so JSON formatting coming
// from the configuration does not churn state.
func destinationsValue(destinations []ometa.SubscriptionDestination, existing types.List) types.List {
	if len(destinations) == 0 {
		if existing.IsNull() || existing.IsUnknown() {
			return types.ListValueMust(destinationObjectType, nil)
		}
		return existing
	}

	value := types.ListValueMust(destinationObjectType, destinationElements(destinations))
	if !existing.IsNull() && !existing.IsUnknown() &&
		reflect.DeepEqual(canonicalFromList(existing), canonicalFromAPI(destinations)) {
		return existing
	}
	return value
}

// destinationElements builds the nested objects for one API destination.
func destinationElements(destinations []ometa.SubscriptionDestination) []attr.Value {
	elements := make([]attr.Value, 0, len(destinations))
	for _, destination := range destinations {
		elements = append(elements, types.ObjectValueMust(destinationObjectType.AttrTypes, map[string]attr.Value{
			"category": types.StringValue(string(destination.Category)),
			"type":     types.StringValue(string(destination.Type)),
			"enabled":  convert.BoolValueOr(destination.Enabled, true),
			"config":   types.StringValue(configJSON(destination.Config)),
		}))
	}
	return elements
}

// destinationsToAPI maps planned destinations to the SDK shape. The
// destination index is part of the error so a malformed config is easy to find.
func destinationsToAPI(ctx context.Context, value types.List) ([]ometa.SubscriptionDestination, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var models []destinationModel
	if diags := value.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, fmt.Errorf("could not decode destinations: %v", diags.Errors())
	}

	destinations := make([]ometa.SubscriptionDestination, 0, len(models))
	for index, model := range models {
		config, err := parseConfigJSON(model.Config)
		if err != nil {
			return nil, fmt.Errorf("destination %d (%s) config is not valid JSON: %w", index, model.Type.ValueString(), err)
		}
		destinations = append(destinations, ometa.SubscriptionDestination{
			Category: ometa.SubscriptionDestinationCategory(model.Category.ValueString()),
			Type:     ometa.SubscriptionDestinationType(model.Type.ValueString()),
			Enabled:  convert.BoolPointer(model.Enabled),
			Config:   config,
		})
	}
	return destinations, nil
}

// destinationsEqual reports whether two destination lists carry the same
// content, ignoring JSON formatting.
func destinationsEqual(left, right types.List) bool {
	if left.IsNull() || left.IsUnknown() || right.IsNull() || right.IsUnknown() {
		return false
	}
	return reflect.DeepEqual(canonicalFromList(left), canonicalFromList(right))
}

// canonicalDestination is a comparable form of a destination. Configs are
// normalized to generic JSON so maps and SDK structs compare structurally.
type canonicalDestination struct {
	Category string
	Type     string
	Enabled  bool
	Config   any
}

// canonicalFromList renders a Terraform destination list in comparable form.
func canonicalFromList(list types.List) []canonicalDestination {
	elements := list.Elements()
	destinations := make([]canonicalDestination, 0, len(elements))
	for _, element := range elements {
		object, ok := element.(types.Object)
		if !ok {
			return nil
		}
		attributes := object.Attributes()
		category, _ := attributes["category"].(types.String)
		kind, _ := attributes["type"].(types.String)
		enabled, _ := attributes["enabled"].(types.Bool)
		config, _ := attributes["config"].(types.String)

		destinations = append(destinations, canonicalDestination{
			Category: category.ValueString(),
			Type:     kind.ValueString(),
			Enabled:  enabled.ValueBool(),
			Config:   normalizeJSON(decodeJSON(config.ValueString())),
		})
	}
	return destinations
}

// canonicalFromAPI renders API destinations in comparable form.
func canonicalFromAPI(destinations []ometa.SubscriptionDestination) []canonicalDestination {
	out := make([]canonicalDestination, 0, len(destinations))
	for _, destination := range destinations {
		out = append(out, canonicalDestination{
			Category: string(destination.Category),
			Type:     string(destination.Type),
			Enabled:  destination.Enabled == nil || *destination.Enabled,
			Config:   normalizeJSON(destination.Config),
		})
	}
	return out
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

// parseConfigJSON decodes a destination config string for an API payload.
// Blank values become an empty object, matching the attribute default.
func parseConfigJSON(value types.String) (any, error) {
	if value.IsNull() || value.IsUnknown() || strings.TrimSpace(value.ValueString()) == "" {
		return map[string]any{}, nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

// decodeJSON decodes a JSON document, returning the original string when it is
// not valid JSON so unequal documents never compare equal.
func decodeJSON(document string) any {
	if strings.TrimSpace(document) == "" {
		return map[string]any{}
	}
	var decoded any
	if err := json.Unmarshal([]byte(document), &decoded); err != nil {
		return document
	}
	return decoded
}

// normalizeJSON round-trips a value through encoding/json so maps, structs and
// a nil config all compare structurally.
func normalizeJSON(value any) any {
	if value == nil {
		return map[string]any{}
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return decodeJSON(string(encoded))
}

// optionalInt32Value keeps an unset optional number unset. Adopting a
// server-side default would show a permanent diff for an attribute the
// configuration deliberately omits.
func optionalInt32Value(value *int32, existing types.Int64) types.Int64 {
	if existing.IsNull() {
		return types.Int64Null()
	}
	if value == nil {
		if existing.IsUnknown() {
			return types.Int64Null()
		}
		return existing
	}
	return types.Int64Value(int64(*value))
}

// int32Pointer maps an optional Terraform number to the SDK's int32 pointer.
// Null and unknown values are omitted from the request.
func int32Pointer(value types.Int64) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	converted := int32(value.ValueInt64())
	return &converted
}
