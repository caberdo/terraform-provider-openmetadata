package workflow_definition

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// workflowDefinitionModel is the Terraform representation of a workflow
// definition lookup. The id and fully_qualified_name attributes are both
// selectors and results.
type workflowDefinitionModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
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

// apply copies the API representation into the Terraform model.
func apply(definition *ometa.WorkflowDefinition, model *workflowDefinitionModel) {
	if definition.Id != nil {
		model.ID = types.StringValue(definition.Id.String())
	}
	model.Name = types.StringValue(definition.Name)
	model.FullyQualifiedName = convert.StringValue(definition.FullyQualifiedName)
	model.DisplayName = convert.StringValue(definition.DisplayName)
	model.Description = types.StringValue(definition.Description)
	model.StoreStageStatus = types.BoolValue(definition.Config != nil && definition.Config.StoreStageStatus)
	model.Nodes = canonicalJSON(definition.Nodes)
	model.Edges = canonicalJSON(definition.Edges)
	model.Trigger = canonicalJSON(definition.Trigger)
	model.Deployed = convert.BoolValueOr(definition.Deployed, false)
	model.Suspended = convert.BoolValueOr(definition.Suspended, false)
	model.Version = convert.Float64Value(definition.Version)
	model.UpdatedAt = convert.TimestampValue(definition.UpdatedAt)
	model.UpdatedBy = convert.StringValue(definition.UpdatedBy)
}

// canonicalJSON renders an API value as JSON. Missing pointers become null so
// the computed attribute is always known after a read.
func canonicalJSON(value any) types.String {
	switch typed := value.(type) {
	case *[]ometa.WorkflowNodeDefinitionInterface:
		if typed == nil {
			return types.StringNull()
		}
		return marshalJSON(*typed)
	case *[]ometa.EdgeDefinition:
		if typed == nil {
			return types.StringNull()
		}
		return marshalJSON(*typed)
	case *ometa.WorkflowTriggerInterface:
		if typed == nil {
			return types.StringNull()
		}
		return marshalJSON(*typed)
	default:
		return types.StringNull()
	}
}

func marshalJSON(value any) types.String {
	encoded, err := json.Marshal(value)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(encoded))
}
