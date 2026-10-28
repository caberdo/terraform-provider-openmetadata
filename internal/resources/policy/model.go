package policy

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// ruleObjectType describes one nested policy rule. The resource and the data
// source share this shape so a rule read from the API can be written straight
// back into state. Operations and resources are sets because OpenMetadata sorts
// and normalizes them.
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

// policyModel is the Terraform representation of an OpenMetadata policy.
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

// apply copies the API representation into the Terraform model. Omitted rule
// fields become null.
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

// rulesToAPI maps planned rules to the SDK shape. Rules and their members are
// sorted so the patch payload is deterministic; the server sorts them anyway.
// The rule index is part of the error so a malformed configuration is easy to
// find.
func rulesToAPI(ctx context.Context, value types.Set) ([]ometa.Rule, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var models []ruleModel
	if diags := value.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, fmt.Errorf("could not decode rules: %v", diags.Errors())
	}

	rules := make([]ometa.Rule, 0, len(models))
	for index, model := range models {
		operations, err := toRuleOperations(model.Operations)
		if err != nil {
			return nil, fmt.Errorf("rule %d (%s): %w", index, model.Name.ValueString(), err)
		}

		resources, err := stringList(model.Resources)
		if err != nil {
			return nil, fmt.Errorf("rule %d (%s): %w", index, model.Name.ValueString(), err)
		}

		sort.Slice(operations, func(i, j int) bool { return operations[i] < operations[j] })
		sort.Strings(resources)

		rules = append(rules, ometa.Rule{
			Name:        model.Name.ValueString(),
			Effect:      ometa.RuleEffect(model.Effect.ValueString()),
			Operations:  operations,
			Resources:   resources,
			Condition:   convert.StringPointer(model.Condition),
			Description: convert.StringPointer(model.Description),
		})
	}

	sort.Slice(rules, func(i, j int) bool { return rules[i].Name < rules[j].Name })

	return rules, nil
}

// toRuleOperations renders a Terraform string set as SDK rule operations.
func toRuleOperations(value types.Set) ([]ometa.RuleOperations, error) {
	values, err := stringList(value)
	if err != nil {
		return nil, err
	}

	operations := make([]ometa.RuleOperations, 0, len(values))
	for _, operation := range values {
		operations = append(operations, ometa.RuleOperations(operation))
	}
	return operations, nil
}

// stringList renders a Terraform string set as a Go slice.
func stringList(value types.Set) ([]string, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	elements := value.Elements()
	values := make([]string, 0, len(elements))
	for _, element := range elements {
		text, ok := element.(types.String)
		if !ok {
			return nil, fmt.Errorf("expected a string element, got %T", element)
		}
		values = append(values, text.ValueString())
	}
	return values, nil
}
