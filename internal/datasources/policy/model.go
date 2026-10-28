package policy

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// ruleObjectType describes one nested policy rule returned by the lookup.
// Operations and resources are sets because OpenMetadata sorts and normalizes
// them.
var ruleObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"name":        types.StringType,
	"effect":      types.StringType,
	"operations":  types.SetType{ElemType: types.StringType},
	"resources":   types.SetType{ElemType: types.StringType},
	"condition":   types.StringType,
	"description": types.StringType,
}}

// ruleModel is the Terraform representation of one policy rule.
type ruleModel struct {
	Name        types.String `tfsdk:"name"`
	Effect      types.String `tfsdk:"effect"`
	Operations  types.Set    `tfsdk:"operations"`
	Resources   types.Set    `tfsdk:"resources"`
	Condition   types.String `tfsdk:"condition"`
	Description types.String `tfsdk:"description"`
}

// policyModel is the Terraform representation of a policy lookup. The id and
// fully_qualified_name attributes are both selectors and results.
type policyModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	Rules              types.Set     `tfsdk:"rules"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(policy *ometa.Policy, model *policyModel) {
	model.ID = types.StringValue(policy.Id.String())
	model.Name = types.StringValue(policy.Name)
	model.FullyQualifiedName = convert.StringValue(policy.FullyQualifiedName)
	model.DisplayName = convert.StringValue(policy.DisplayName)
	model.Description = convert.StringValueOr(policy.Description, "")
	model.Enabled = convert.BoolValueOr(policy.Enabled, true)
	model.Rules = rulesValue(policy.Rules)
	model.Version = convert.Float64Value(policy.Version)
	model.UpdatedAt = convert.TimestampValue(policy.UpdatedAt)
	model.UpdatedBy = convert.StringValue(policy.UpdatedBy)
}

// rulesValue maps API rules to a Terraform set.
func rulesValue(rules []ometa.Rule) types.Set {
	elements := make([]attr.Value, 0, len(rules))
	for _, rule := range rules {
		operations := make([]attr.Value, 0, len(rule.Operations))
		for _, operation := range rule.Operations {
			operations = append(operations, types.StringValue(string(operation)))
		}

		resources := make([]attr.Value, 0, len(rule.Resources))
		for _, resource := range rule.Resources {
			resources = append(resources, types.StringValue(resource))
		}

		elements = append(elements, types.ObjectValueMust(ruleObjectType.AttrTypes, map[string]attr.Value{
			"name":        types.StringValue(rule.Name),
			"effect":      types.StringValue(string(rule.Effect)),
			"operations":  types.SetValueMust(types.StringType, operations),
			"resources":   types.SetValueMust(types.StringType, resources),
			"condition":   convert.StringValue(rule.Condition),
			"description": convert.StringValue(rule.Description),
		}))
	}

	return types.SetValueMust(ruleObjectType, elements)
}
