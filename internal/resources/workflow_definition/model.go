package workflow_definition

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// workflowDefinitionModel is the Terraform representation of an OpenMetadata
// workflow definition. Nodes, edges and trigger are kept as JSON strings,
// because their shapes depend on the node types a workflow uses.
type workflowDefinitionModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Type               types.String  `tfsdk:"type"`
	StoreStageStatus   types.Bool    `tfsdk:"store_stage_status"`
	Nodes              types.String  `tfsdk:"nodes"`
	Edges              types.String  `tfsdk:"edges"`
	Trigger            types.String  `tfsdk:"trigger"`
	Deployed           types.Bool    `tfsdk:"deployed"`
	Suspended          types.Bool    `tfsdk:"suspended"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Type is left
// untouched, because the workflow definition response does not carry it.
func apply(definition *ometa.WorkflowDefinition, model *workflowDefinitionModel) {
	if definition.Id != nil {
		model.ID = types.StringValue(definition.Id.String())
	}
	model.Name = types.StringValue(definition.Name)
	model.FullyQualifiedName = convert.StringValue(definition.FullyQualifiedName)
	model.DisplayName = convert.StringValue(definition.DisplayName)
	model.Description = types.StringValue(definition.Description)
	model.StoreStageStatus = storeStageStatusValue(definition.Config, model.StoreStageStatus)
	model.Nodes = nodesValue(definition.Nodes, model.Nodes)
	model.Edges = edgesValue(definition.Edges, model.Edges)
	model.Trigger = triggerValue(definition.Trigger, model.Trigger)
	model.Deployed = convert.BoolValueOr(definition.Deployed, false)
	model.Suspended = convert.BoolValueOr(definition.Suspended, false)
	model.Version = convert.Float64Value(definition.Version)
	model.UpdatedAt = convert.TimestampValue(definition.UpdatedAt)
	model.UpdatedBy = convert.StringValue(definition.UpdatedBy)
}

// storeStageStatusValue reads the workflow configuration. A missing
// configuration never zeroes a value that is already in state.
func storeStageStatusValue(config *ometa.WorkflowConfiguration, existing types.Bool) types.Bool {
	if config == nil {
		if existing.IsUnknown() || existing.IsNull() {
			return types.BoolValue(false)
		}
		return existing
	}
	return types.BoolValue(config.StoreStageStatus)
}

func nodesValue(nodes *[]ometa.WorkflowNodeDefinitionInterface, existing types.String) types.String {
	if nodes == nil {
		return preserveOrNull(existing)
	}
	return canonicalJSON(*nodes, existing)
}

func edgesValue(edges *[]ometa.EdgeDefinition, existing types.String) types.String {
	if edges == nil {
		return preserveOrNull(existing)
	}
	return canonicalJSON(*edges, existing)
}

func triggerValue(trigger *ometa.WorkflowTriggerInterface, existing types.String) types.String {
	if trigger == nil {
		return preserveOrNull(existing)
	}
	return canonicalJSON(*trigger, existing)
}

// preserveOrNull keeps a known value when the API omits a field, but turns an
// unknown computed value into null so create and import always end known.
func preserveOrNull(existing types.String) types.String {
	if existing.IsUnknown() {
		return types.StringNull()
	}
	return existing
}

// canonicalJSON renders an API value as JSON. When the existing value is
// semantically equal it is kept verbatim, so formatting coming from the
// configuration does not churn state.
func canonicalJSON(value any, existing types.String) types.String {
	encoded, err := json.Marshal(value)
	if err != nil {
		return existing
	}

	rendered := string(encoded)
	if !existing.IsUnknown() && !existing.IsNull() && jsonEqual(existing.ValueString(), rendered) {
		return existing
	}
	return types.StringValue(rendered)
}

// decodeNodes parses the nodes attribute for an API payload. A missing
// attribute yields nil so the field is omitted from the request.
func decodeNodes(value types.String) (*[]ometa.WorkflowNodeDefinitionInterface, error) {
	if !configured(value) {
		return nil, nil
	}
	var decoded []ometa.WorkflowNodeDefinitionInterface
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

// decodeEdges parses the edges attribute for an API payload.
func decodeEdges(value types.String) (*[]ometa.EdgeDefinition, error) {
	if !configured(value) {
		return nil, nil
	}
	var decoded []ometa.EdgeDefinition
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

// decodeTrigger parses the trigger attribute for an API payload.
func decodeTrigger(value types.String) (*ometa.WorkflowTriggerInterface, error) {
	if !configured(value) {
		return nil, nil
	}
	var decoded ometa.WorkflowTriggerInterface
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

// configured reports whether the attribute holds a value to send.
func configured(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown() && strings.TrimSpace(value.ValueString()) != ""
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
