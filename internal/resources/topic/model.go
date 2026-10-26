package topic

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// topicModel is the Terraform representation of an OpenMetadata topic. The
// message schema and topic configuration are kept as JSON object strings,
// because their shape depends on the broker and schema registry.
type topicModel struct {
	ID                    types.String  `tfsdk:"id"`
	Name                  types.String  `tfsdk:"name"`
	Service               types.String  `tfsdk:"service"`
	FullyQualifiedName    types.String  `tfsdk:"fully_qualified_name"`
	DisplayName           types.String  `tfsdk:"display_name"`
	Description           types.String  `tfsdk:"description"`
	Partitions            types.Int64   `tfsdk:"partitions"`
	MaximumMessageSize    types.Int64   `tfsdk:"maximum_message_size"`
	MinimumInSyncReplicas types.Int64   `tfsdk:"minimum_in_sync_replicas"`
	ReplicationFactor     types.Int64   `tfsdk:"replication_factor"`
	RetentionSize         types.Float64 `tfsdk:"retention_size"`
	RetentionTime         types.Float64 `tfsdk:"retention_time"`
	CleanupPolicies       types.List    `tfsdk:"cleanup_policies"`
	MessageSchema         types.String  `tfsdk:"message_schema"`
	TopicConfig           types.String  `tfsdk:"topic_config"`
	Version               types.Float64 `tfsdk:"version"`
	UpdatedAt             types.String  `tfsdk:"updated_at"`
	UpdatedBy             types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The parent
// service selector is left untouched when the server omits the reference, and
// JSON attributes already in state are kept when the API omits them.
func apply(ctx context.Context, topic *ometa.Topic, model *topicModel) diag.Diagnostics {
	model.ID = types.StringValue(topic.Id.String())
	model.Name = types.StringValue(topic.Name)
	model.FullyQualifiedName = convert.StringValue(topic.FullyQualifiedName)
	model.DisplayName = convert.StringValue(topic.DisplayName)
	model.Description = convert.StringValueOr(topic.Description, "")
	if fqn := topic.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := topic.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	model.Partitions = types.Int64Value(int64(topic.Partitions))
	model.MaximumMessageSize = convert.Int32Value(topic.MaximumMessageSize)
	model.MinimumInSyncReplicas = convert.Int32Value(topic.MinimumInSyncReplicas)
	model.ReplicationFactor = convert.Int32Value(topic.ReplicationFactor)
	model.RetentionSize = convert.Float64Value(topic.RetentionSize)
	model.RetentionTime = convert.Float64Value(topic.RetentionTime)
	model.MessageSchema = messageSchemaValue(topic.MessageSchema, model.MessageSchema)
	model.TopicConfig = topicConfigValue(topic.TopicConfig, model.TopicConfig)
	model.Version = convert.Float64Value(topic.Version)
	model.UpdatedAt = convert.TimestampValue(topic.UpdatedAt)
	model.UpdatedBy = convert.StringValue(topic.UpdatedBy)

	if topic.CleanupPolicies != nil {
		policies := make([]string, 0, len(*topic.CleanupPolicies))
		for _, policy := range *topic.CleanupPolicies {
			policies = append(policies, string(policy))
		}
		list, diags := types.ListValueFrom(ctx, types.StringType, policies)
		model.CleanupPolicies = list
		return diags
	}
	if model.CleanupPolicies.IsUnknown() {
		model.CleanupPolicies = types.ListNull(types.StringType)
	}
	return nil
}

// messageSchemaValue renders the message schema. When the existing value is
// known and semantically equal it is kept verbatim, so formatting coming from
// the configuration does not churn state. A missing schema never zeroes a value
// that is already in state.
func messageSchemaValue(schema *ometa.MessageSchema, existing types.String) types.String {
	if schema == nil {
		if existing.IsUnknown() {
			return types.StringNull()
		}
		return existing
	}

	encoded, err := json.Marshal(schema)
	if err != nil {
		if existing.IsUnknown() {
			return types.StringNull()
		}
		return existing
	}

	// A server that stores an empty schema field list returns it explicitly;
	// drop it so it matches a configuration that omits the key.
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err == nil {
		if fields, ok := document["schemaFields"].([]any); ok && len(fields) == 0 {
			delete(document, "schemaFields")
		}
		if normalized, err := json.Marshal(document); err == nil {
			encoded = normalized
		}
	}

	value := string(encoded)
	if !existing.IsUnknown() && !existing.IsNull() && jsonEqual(existing.ValueString(), value) {
		return existing
	}
	return types.StringValue(value)
}

// topicConfigValue renders the broker configuration. Missing configurations do
// not zero a value already in state.
func topicConfigValue(config *ometa.TopicConfig, existing types.String) types.String {
	if config == nil {
		if existing.IsUnknown() {
			return types.StringNull()
		}
		return existing
	}

	encoded, err := json.Marshal(*config)
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

// parseMessageSchemaJSON decodes the message schema attribute for an API
// payload.
func parseMessageSchemaJSON(value string) (*ometa.MessageSchema, error) {
	var decoded ometa.MessageSchema
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

// parseTopicConfigJSON decodes the topic configuration attribute for an API
// payload.
func parseTopicConfigJSON(value string) (*ometa.TopicConfig, error) {
	var decoded ometa.TopicConfig
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
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

// jsonConfigured reports whether the attribute holds a value to send.
func jsonConfigured(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && strings.TrimSpace(value.ValueString()) != ""
}

// int32Pointer maps an optional Terraform number to the SDK's int32 pointer.
func int32Pointer(value types.Int64) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := int32(value.ValueInt64())
	return &v
}

// float64Pointer maps an optional Terraform number to a float64 pointer.
func float64Pointer(value types.Float64) *float64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueFloat64()
	return &v
}

// cleanupPoliciesToAPI maps planned cleanup policy names to the create enum.
func cleanupPoliciesToAPI(ctx context.Context, value types.List) ([]ometa.CreateTopicCleanupPolicies, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var names []string
	diags := value.ElementsAs(ctx, &names, false)
	if diags.HasError() {
		return nil, diags
	}

	policies := make([]ometa.CreateTopicCleanupPolicies, 0, len(names))
	for _, name := range names {
		policies = append(policies, ometa.CreateTopicCleanupPolicies(name))
	}
	return policies, diags
}
