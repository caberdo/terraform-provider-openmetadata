package ai_governance_policy

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

	if resp.TypeName != "openmetadata_ai_governance_policy" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ai_governance_policy")
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
		if !ok || !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if len(selector.Validators) == 0 {
			t.Fatalf("%s must validate the selector pair", attribute)
		}
	}

	for _, attribute := range []string{"name", "policy_type", "enforcement_level", "display_name", "description", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed || computed.Optional || computed.Required {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	enabled, ok := resp.Schema.Attributes["enabled"].(dschema.BoolAttribute)
	if !ok || !enabled.Computed {
		t.Fatal("enabled must be a computed bool")
	}

	version, ok := resp.Schema.Attributes["version"].(dschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "bias-review"
	displayName := "Bias review"
	description := "Flags model bias above the threshold."
	enforcementLevel := ometa.AIGovernancePolicyEnforcementLevelAdvisory
	enabled := true

	var model aiGovernancePolicyModel
	apply(&ometa.AIGovernancePolicy{
		Id:                 id,
		Name:               "bias-review",
		PolicyType:         ometa.AIGovernancePolicyPolicyTypeBiasThreshold,
		EnforcementLevel:   &enforcementLevel,
		Enabled:            &enabled,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "bias-review" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.PolicyType.ValueString() != "BiasThreshold" {
		t.Fatalf("PolicyType = %q", model.PolicyType.ValueString())
	}
	if model.EnforcementLevel.ValueString() != "Advisory" {
		t.Fatalf("EnforcementLevel = %q", model.EnforcementLevel.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Fatal("Enabled = false")
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model aiGovernancePolicyModel
	apply(&ometa.AIGovernancePolicy{
		Id:         id,
		Name:       "bare",
		PolicyType: ometa.AIGovernancePolicyPolicyTypeComplianceCheck,
	}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.Enabled.ValueBool() {
		t.Fatal("Enabled = true, want the false default")
	}
	if !model.EnforcementLevel.IsNull() {
		t.Fatalf("EnforcementLevel = %v, want null", model.EnforcementLevel)
	}
}
