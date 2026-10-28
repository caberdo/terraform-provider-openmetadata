package ai_governance_policy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// aiGovernancePolicyModel is the Terraform representation of an OpenMetadata
// AI governance policy.
type aiGovernancePolicyModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	PolicyType         types.String  `tfsdk:"policy_type"`
	EnforcementLevel   types.String  `tfsdk:"enforcement_level"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. enforcement
// level is left untouched when the server omits it, so a response without the
// field never zeroes state.
func apply(policy *ometa.AIGovernancePolicy, model *aiGovernancePolicyModel) {
	model.ID = types.StringValue(policy.Id.String())
	model.Name = types.StringValue(policy.Name)
	model.PolicyType = types.StringValue(string(policy.PolicyType))
	if policy.EnforcementLevel != nil && *policy.EnforcementLevel != "" {
		model.EnforcementLevel = types.StringValue(string(*policy.EnforcementLevel))
	} else if model.EnforcementLevel.IsUnknown() {
		model.EnforcementLevel = types.StringNull()
	}
	model.Enabled = convert.BoolValueOr(policy.Enabled, false)
	model.FullyQualifiedName = convert.StringValue(policy.FullyQualifiedName)
	model.DisplayName = convert.StringValue(policy.DisplayName)
	model.Description = convert.StringValueOr(policy.Description, "")
	model.Version = convert.Float64Value(policy.Version)
	model.UpdatedAt = convert.TimestampValue(policy.UpdatedAt)
	model.UpdatedBy = convert.StringValue(policy.UpdatedBy)
}
