package policy

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_policy" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_policy")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be an optional, computed selector", attribute)
		}
		if len(selector.Validators) == 0 {
			t.Fatalf("%s must require exactly one identifier", attribute)
		}
	}

	for _, attribute := range []string{"name", "display_name", "description", "version", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if attribute == "version" {
			floatAttribute, floatOK := resp.Schema.Attributes[attribute].(dschema.Float64Attribute)
			if !floatOK || !floatAttribute.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	enabled, ok := resp.Schema.Attributes["enabled"].(dschema.BoolAttribute)
	if !ok || !enabled.Computed {
		t.Fatal("enabled must be a computed bool")
	}

	rules, ok := resp.Schema.Attributes["rules"].(dschema.SetNestedAttribute)
	if !ok || !rules.Computed {
		t.Fatal("rules must be a computed nested set")
	}
	for _, attribute := range []string{"name", "effect", "operations", "resources", "condition", "description"} {
		if _, ok := rules.NestedObject.Attributes[attribute]; !ok {
			t.Fatalf("rules is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "viewer"
	displayName := "Viewer policy"
	condition := "isOwner"

	var model policyModel
	apply(&ometa.Policy{
		Id:                 id,
		Name:               "viewer",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        stringPointer("Read-only access"),
		Enabled:            boolPointer(false),
		Rules: []ometa.Rule{
			{
				Name:       "view",
				Effect:     ometa.RuleEffectAllow,
				Operations: []ometa.RuleOperations{ometa.RuleOperationsViewAll},
				Resources:  []string{"table"},
				Condition:  &condition,
			},
		},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "viewer" {
		t.Fatalf("Name = %q", model.Name.ValueString())
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

	var rules []ruleModel
	if diags := model.Rules.ElementsAs(context.Background(), &rules, false); diags.HasError() {
		t.Fatalf("decoding rules: %v", diags.Errors())
	}
	if len(rules) != 1 {
		t.Fatalf("rules length = %d, want 1", len(rules))
	}
	if rules[0].Name.ValueString() != "view" || rules[0].Effect.ValueString() != "allow" {
		t.Fatalf("rule = %+v", rules[0])
	}
	if rules[0].Condition.ValueString() != condition {
		t.Fatalf("rule condition = %q", rules[0].Condition.ValueString())
	}
	if !rules[0].Description.IsNull() {
		t.Fatal("omitted rule description must be null")
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	var model policyModel
	apply(&ometa.Policy{
		Id:   uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"),
		Name: "bare",
	}, &model)

	if !model.Enabled.ValueBool() {
		t.Fatal("Enabled = false, want the true default")
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Rules.IsNull() || len(model.Rules.Elements()) != 0 {
		t.Fatalf("Rules = %v, want an empty set", model.Rules)
	}
}

func stringPointer(value string) *string {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
