package ai_governance_policy

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_ai_governance_policy" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ai_governance_policy")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatal("name must be a required string")
	}
	if len(name.PlanModifiers) == 0 {
		t.Fatal("name must carry a plan modifier")
	}

	policyType, ok := resp.Schema.Attributes["policy_type"].(rschema.StringAttribute)
	if !ok || !policyType.Required || policyType.Computed || policyType.Optional {
		t.Fatal("policy_type must be required only")
	}
	if len(policyType.Validators) == 0 {
		t.Fatal("policy_type must validate the policy category")
	}

	enforcementLevel, ok := resp.Schema.Attributes["enforcement_level"].(rschema.StringAttribute)
	if !ok || !enforcementLevel.Optional || !enforcementLevel.Computed {
		t.Fatal("enforcement_level must be optional and computed")
	}
	if len(enforcementLevel.Validators) == 0 {
		t.Fatal("enforcement_level must validate the enforcement level")
	}

	enabled, ok := resp.Schema.Attributes["enabled"].(rschema.BoolAttribute)
	if !ok || !enabled.Optional || !enabled.Computed || enabled.Default == nil {
		t.Fatal("enabled must be optional and computed with a false default")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed || description.Default == nil {
		t.Fatal("description must be optional and computed with an empty default")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Computed || displayName.Required {
		t.Fatal("display_name must be optional only")
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
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
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

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
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "bias-review" {
		t.Fatalf("Name = %q", model.Name.ValueString())
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
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
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
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model aiGovernancePolicyModel
	apply(&ometa.AIGovernancePolicy{
		Id:         id,
		Name:       "bare",
		PolicyType: ometa.AIGovernancePolicyPolicyTypeComplianceCheck,
	}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.Enabled.ValueBool() {
		t.Fatal("Enabled = true, want the false default")
	}
	if !model.EnforcementLevel.IsNull() {
		t.Fatalf("EnforcementLevel = %v, want null", model.EnforcementLevel)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := aiGovernancePolicyModel{
		Name:             types.StringValue("bias-review"),
		PolicyType:       types.StringValue("BiasThreshold"),
		EnforcementLevel: types.StringValue("Advisory"),
		Enabled:          types.BoolValue(false),
		DisplayName:      types.StringValue("Display"),
		Description:      types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      aiGovernancePolicyModel
		state     aiGovernancePolicyModel
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name added",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       base.PolicyType,
				EnforcementLevel: base.EnforcementLevel,
				Enabled:          base.Enabled,
				DisplayName:      types.StringValue("New display"),
				Description:      base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       base.PolicyType,
				EnforcementLevel: base.EnforcementLevel,
				Enabled:          base.Enabled,
				DisplayName:      types.StringNull(),
				Description:      base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       base.PolicyType,
				EnforcementLevel: base.EnforcementLevel,
				Enabled:          base.Enabled,
				DisplayName:      base.DisplayName,
				Description:      types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "policy type changed",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       types.StringValue("CostControl"),
				EnforcementLevel: base.EnforcementLevel,
				Enabled:          base.Enabled,
				DisplayName:      base.DisplayName,
				Description:      base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/policyType",
			wantOp:    "add",
			wantValue: "CostControl",
		},
		{
			name: "enabled changed",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       base.PolicyType,
				EnforcementLevel: base.EnforcementLevel,
				Enabled:          types.BoolValue(true),
				DisplayName:      base.DisplayName,
				Description:      base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/enabled",
			wantOp:    "add",
			wantValue: true,
		},
		{
			name: "enforcement level changed",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       base.PolicyType,
				EnforcementLevel: types.StringValue("Blocking"),
				Enabled:          base.Enabled,
				DisplayName:      base.DisplayName,
				Description:      base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/enforcementLevel",
			wantOp:    "add",
			wantValue: "Blocking",
		},
		{
			name: "enforcement level removed",
			plan: aiGovernancePolicyModel{
				Name:             base.Name,
				PolicyType:       base.PolicyType,
				EnforcementLevel: types.StringNull(),
				Enabled:          base.Enabled,
				DisplayName:      base.DisplayName,
				Description:      base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/enforcementLevel",
			wantOp:   "remove",
		},
		{
			name: "name cannot be patched",
			plan: aiGovernancePolicyModel{
				Name:             types.StringValue("renamed"),
				PolicyType:       base.PolicyType,
				EnforcementLevel: base.EnforcementLevel,
				Enabled:          base.Enabled,
				DisplayName:      base.DisplayName,
				Description:      base.Description,
			},
			state:   base,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
			if len(ops) != tt.wantOps {
				t.Fatalf("len(ops) = %d, want %d: %#v", len(ops), tt.wantOps, ops)
			}
			if tt.wantOps == 0 {
				return
			}
			if ops[0].Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", ops[0].Path, tt.wantPath)
			}
			if ops[0].Op != tt.wantOp {
				t.Fatalf("op = %q, want %q", ops[0].Op, tt.wantOp)
			}
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
