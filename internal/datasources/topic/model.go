package topic

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// topicModel is the Terraform representation of a topic lookup. The id and
// fully_qualified_name attributes are both selectors and results.
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

// apply copies the API representation into the Terraform model.
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
	model.MessageSchema = messageSchemaValue(topic.MessageSchema)
	model.TopicConfig = topicConfigValue(topic.TopicConfig)
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
	model.CleanupPolicies = types.ListNull(types.StringType)
	return nil
}

// messageSchemaValue renders the message schema as a compact JSON object
// string.
func messageSchemaValue(schema *ometa.MessageSchema) types.String {
	if schema == nil {
		return types.StringNull()
	}

	encoded, err := json.Marshal(schema)
	if err != nil {
		return types.StringNull()
	}

	return types.StringValue(string(encoded))
}

// topicConfigValue renders the broker configuration as a compact JSON object
// string.
func topicConfigValue(config *ometa.TopicConfig) types.String {
	if config == nil {
		return types.StringNull()
	}

	encoded, err := json.Marshal(*config)
	if err != nil {
		return types.StringNull()
	}

	return types.StringValue(string(encoded))
}
