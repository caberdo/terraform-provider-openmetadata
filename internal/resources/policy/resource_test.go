package policy

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_policy" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_policy")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok {
		t.Fatalf("name attribute is %T, want StringAttribute", resp.Schema.Attributes["name"])
	}
	if !name.Required {
		t.Fatal("name must be required")
	}
	if len(name.PlanModifiers) == 0 {
		t.Fatal("name must carry a plan modifier so a rename replaces the policy")
	}

	rules, ok := resp.Schema.Attributes["rules"].(rschema.SetNestedAttribute)
	if !ok {
		t.Fatalf("rules attribute is %T, want SetNestedAttribute", resp.Schema.Attributes["rules"])
	}
	if !rules.Required {
		t.Fatal("rules must be required")
	}
	if len(rules.Validators) == 0 {
		t.Fatal("rules must require at least one entry")
	}

	operations, ok := rules.NestedObject.Attributes["operations"].(rschema.SetAttribute)
	if !ok || !operations.Required {
		t.Fatal("rules.operations must be a required set")
	}
	if len(operations.Validators) == 0 {
		t.Fatal("rules.operations must restrict its values")
	}

	effect, ok := rules.NestedObject.Attributes["effect"].(rschema.StringAttribute)
	if !ok || !effect.Required || len(effect.Validators) == 0 {
		t.Fatal("rules.effect must be a required, validated string")
	}

	resources, ok := rules.NestedObject.Attributes["resources"].(rschema.SetAttribute)
	if !ok || !resources.Required || len(resources.Validators) == 0 {
		t.Fatal("rules.resources must be a required, non-empty set")
	}

	enabled, ok := resp.Schema.Attributes["enabled"].(rschema.BoolAttribute)
	if !ok {
		t.Fatalf("enabled attribute is %T, want BoolAttribute", resp.Schema.Attributes["enabled"])
	}
	if !enabled.Optional || !enabled.Computed {
		t.Fatal("enabled must be optional and computed so the server default applies")
	}
	if enabled.Default != nil {
		t.Fatal("enabled must not carry a client-side default")
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "version", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if attribute == "version" {
			floatAttribute, floatOK := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !floatOK || !floatAttribute.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	displayName := "Viewer policy"
	fqn := "viewer"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	enabled := false
	condition := "isOwner"

	var model policyModel
	apply(&ometa.Policy{
		Id:                 id,
		Name:               "viewer",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        stringPointer("Read-only access"),
		Enabled:            &enabled,
		Rules: []ometa.Rule{
			{
				Name:        "view",
				Effect:      ometa.RuleEffectAllow,
				Operations:  []ometa.RuleOperations{ometa.RuleOperationsViewAll},
				Resources:   []string{"table"},
				Condition:   &condition,
				Description: stringPointer("View tables"),
			},
			{
				Name:       "deny-delete",
				Effect:     ometa.RuleEffectDeny,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsDelete},
				Resources:  []string{"table", "dashboard"},
			},
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "viewer" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Read-only access" {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Enabled.ValueBool() {
		t.Fatal("Enabled = true")
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
	if model.UpdatedBy.ValueString() != updatedBy {
		t.Fatalf("UpdatedBy = %q", model.UpdatedBy.ValueString())
	}

	ruleModels := decodeRules(t, model.Rules)
	if len(ruleModels) != 2 {
		t.Fatalf("rules length = %d, want 2", len(ruleModels))
	}

	view := ruleByName(t, ruleModels, "view")
	if view.Effect.ValueString() != "allow" {
		t.Fatalf("view rule = %+v", view)
	}
	if view.Condition.ValueString() != condition {
		t.Fatalf("view rule condition = %q", view.Condition.ValueString())
	}
	if view.Description.ValueString() != "View tables" {
		t.Fatalf("view rule description = %q", view.Description.ValueString())
	}

	deny := ruleByName(t, ruleModels, "deny-delete")
	if deny.Effect.ValueString() != "deny" {
		t.Fatalf("deny rule effect = %q", deny.Effect.ValueString())
	}
	if !deny.Condition.IsNull() || !deny.Description.IsNull() {
		t.Fatal("omitted rule fields must be null")
	}
	if got := attrStrings(t, deny.Resources); !reflect.DeepEqual(got, []string{"dashboard", "table"}) {
		t.Fatalf("deny rule resources = %v", got)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model policyModel
	apply(&ometa.Policy{
		Id:   id,
		Name: "bare",
		Rules: []ometa.Rule{
			{
				Name:       "view",
				Effect:     ometa.RuleEffectAllow,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsViewAll},
				Resources:  []string{"table"},
			},
		},
	}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Fatal("Enabled = false, want the true default")
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}

	ruleModels := decodeRules(t, model.Rules)
	if len(ruleModels) != 1 {
		t.Fatalf("rules length = %d, want 1", len(ruleModels))
	}
	view := ruleByName(t, ruleModels, "view")
	if !view.Condition.IsNull() || !view.Description.IsNull() {
		t.Fatal("omitted rule fields must be null")
	}
	if got := attrStrings(t, view.Operations); !reflect.DeepEqual(got, []string{"ViewAll"}) {
		t.Fatalf("view rule operations = %v", got)
	}
}

func TestPatchOps(t *testing.T) {
	ctx := context.Background()
	base := policyModel{
		Name:        types.StringValue("viewer"),
		DisplayName: types.StringValue("Viewer"),
		Description: types.StringValue("before"),
		Enabled:     types.BoolValue(true),
		Rules: rulesValue([]ometa.Rule{
			{
				Name:       "view",
				Effect:     ometa.RuleEffectAllow,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsViewAll},
				Resources:  []string{"table"},
			},
		}),
	}

	t.Run("no changes", func(t *testing.T) {
		ops, err := patchOps(ctx, base, base)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		if len(ops) != 0 {
			t.Fatalf("ops = %v, want none", ops)
		}
	})

	t.Run("server ordering is irrelevant", func(t *testing.T) {
		plan := base
		plan.Rules = rulesValue([]ometa.Rule{
			{
				Name:       "view",
				Effect:     ometa.RuleEffectAllow,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsViewAll, ometa.RuleOperationsViewBasic},
				Resources:  []string{"table", "dashboard"},
			},
		})
		state := base
		state.Rules = rulesValue([]ometa.Rule{
			{
				Name:       "view",
				Effect:     ometa.RuleEffectAllow,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsViewBasic, ometa.RuleOperationsViewAll},
				Resources:  []string{"dashboard", "table"},
			},
		})

		ops, err := patchOps(ctx, plan, state)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		if len(ops) != 0 {
			t.Fatalf("ops = %v, want none for reordered set members", ops)
		}
	})

	t.Run("name is never patched", func(t *testing.T) {
		plan := base
		plan.Name = types.StringValue("renamed")
		ops, err := patchOps(ctx, plan, base)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		if len(ops) != 0 {
			t.Fatalf("ops = %v, want none for an immutable name", ops)
		}
	})

	t.Run("display name replaced", func(t *testing.T) {
		plan := base
		plan.DisplayName = types.StringValue("New display")
		ops, err := patchOps(ctx, plan, base)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		requireSingleOp(t, ops, "add", "/displayName", "New display")
	})

	t.Run("display name removed", func(t *testing.T) {
		plan := base
		plan.DisplayName = types.StringNull()
		ops, err := patchOps(ctx, plan, base)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		requireSingleOp(t, ops, "remove", "/displayName", nil)
	})

	t.Run("description changed", func(t *testing.T) {
		plan := base
		plan.Description = types.StringValue("after")
		ops, err := patchOps(ctx, plan, base)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		requireSingleOp(t, ops, "add", "/description", "after")
	})

	t.Run("enabled changed", func(t *testing.T) {
		plan := base
		plan.Enabled = types.BoolValue(false)
		ops, err := patchOps(ctx, plan, base)
		if err != nil {
			t.Fatalf("patchOps() error = %v", err)
		}
		requireSingleOp(t, ops, "add", "/enabled", false)
	})
}

func TestPatchOpsReplacesRules(t *testing.T) {
	ctx := context.Background()
	state := policyModel{
		Name:    types.StringValue("viewer"),
		Enabled: types.BoolValue(true),
		Rules: rulesValue([]ometa.Rule{
			{
				Name:       "view",
				Effect:     ometa.RuleEffectAllow,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsViewAll},
				Resources:  []string{"table"},
			},
		}),
	}

	condition := "isOwner"
	description := "Edit owned tables"
	plan := state
	plan.Rules = rulesValue([]ometa.Rule{
		{
			Name:       "view",
			Effect:     ometa.RuleEffectAllow,
			Operations: []ometa.RuleOperations{ometa.RuleOperationsViewBasic},
			Resources:  []string{"table", "dashboard"},
		},
		{
			Name:        "edit",
			Effect:      ometa.RuleEffectDeny,
			Operations:  []ometa.RuleOperations{ometa.RuleOperationsEditDescription},
			Resources:   []string{"table"},
			Condition:   &condition,
			Description: &description,
		},
	})

	ops, err := patchOps(ctx, plan, state)
	if err != nil {
		t.Fatalf("patchOps() error = %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops = %v, want one /rules operation", ops)
	}
	if ops[0].Op != "add" || ops[0].Path != "/rules" {
		t.Fatalf("op = %+v, want add /rules", ops[0])
	}

	rules, ok := ops[0].Value.([]ometa.Rule)
	if !ok {
		t.Fatalf("value is %T, want []ometa.Rule", ops[0].Value)
	}
	if len(rules) != 2 {
		t.Fatalf("rules length = %d, want 2", len(rules))
	}
	// The payload is sorted by rule name so it is deterministic.
	if rules[0].Name != "edit" || rules[1].Name != "view" {
		t.Fatalf("rule order = %q, %q, want edit, view", rules[0].Name, rules[1].Name)
	}
	if rules[0].Condition == nil || *rules[0].Condition != condition {
		t.Fatalf("edit rule condition = %v", rules[0].Condition)
	}
	if rules[0].Description == nil || *rules[0].Description != description {
		t.Fatalf("edit rule description = %v", rules[0].Description)
	}
	if !reflect.DeepEqual(rules[1].Operations, []ometa.RuleOperations{ometa.RuleOperationsViewBasic}) {
		t.Fatalf("view rule operations = %v", rules[1].Operations)
	}
	if !reflect.DeepEqual(rules[1].Resources, []string{"dashboard", "table"}) {
		t.Fatalf("view rule resources = %v", rules[1].Resources)
	}
}

func decodeRules(t *testing.T, set types.Set) []ruleModel {
	t.Helper()

	var models []ruleModel
	if diags := set.ElementsAs(context.Background(), &models, false); diags.HasError() {
		t.Fatalf("decoding rules: %v", diags.Errors())
	}
	return models
}

func ruleByName(t *testing.T, models []ruleModel, name string) ruleModel {
	t.Helper()

	for _, model := range models {
		if model.Name.ValueString() == name {
			return model
		}
	}
	t.Fatalf("rule %q not found in %v", name, models)
	return ruleModel{}
}

func attrStrings(t *testing.T, set types.Set) []string {
	t.Helper()

	values := make([]string, 0, len(set.Elements()))
	for _, element := range set.Elements() {
		text, ok := element.(types.String)
		if !ok {
			t.Fatalf("element is %T, want types.String", element)
		}
		values = append(values, text.ValueString())
	}
	sort.Strings(values)
	return values
}

func stringPointer(value string) *string {
	return &value
}

func requireSingleOp(t *testing.T, ops []ometa.JSONPatchOp, wantOp, wantPath string, wantValue any) {
	t.Helper()

	if len(ops) != 1 {
		t.Fatalf("ops = %v, want exactly one", ops)
	}
	if ops[0].Op != wantOp || ops[0].Path != wantPath {
		t.Fatalf("op = %+v, want %s %s", ops[0], wantOp, wantPath)
	}
	if wantValue == nil {
		if ops[0].Value != nil {
			t.Fatalf("value = %v, want nil", ops[0].Value)
		}
		return
	}
	if !reflect.DeepEqual(ops[0].Value, wantValue) {
		t.Fatalf("value = %v, want %v", ops[0].Value, wantValue)
	}
}

func TestResourcesValidator(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		values  []string
		wantErr bool
	}{
		{name: "plain resources", values: []string{"table", "dashboard"}},
		{name: "canonical all", values: []string{"All"}},
		{name: "lowercase all", values: []string{"all"}, wantErr: true},
		{name: "all combined with another resource", values: []string{"All", "table"}, wantErr: true},
		{name: "lowercase all combined with another resource", values: []string{"all", "dashboard"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, diags := types.SetValueFrom(ctx, types.StringType, tt.values)
			if diags.HasError() {
				t.Fatalf("building set: %v", diags.Errors())
			}

			var resp validator.SetResponse
			policyResourcesValidator{}.ValidateSet(ctx, validator.SetRequest{
				Path:        path.Root("rules").AtListIndex(0).AtName("resources"),
				ConfigValue: value,
			}, &resp)

			if got := resp.Diagnostics.HasError(); got != tt.wantErr {
				t.Fatalf("HasError() = %v, want %v (diagnostics: %v)", got, tt.wantErr, resp.Diagnostics.Errors())
			}
		})
	}
}
