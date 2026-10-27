package ai_governance_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

// policyTypes are the AI governance policy categories OpenMetadata accepts.
var policyTypes = []string{
	string(ometa.CreateAIGovernancePolicyPolicyTypeBiasThreshold),
	string(ometa.CreateAIGovernancePolicyPolicyTypeComplianceCheck),
	string(ometa.CreateAIGovernancePolicyPolicyTypeCostControl),
	string(ometa.CreateAIGovernancePolicyPolicyTypeDataAccess),
	string(ometa.CreateAIGovernancePolicyPolicyTypeModelApproval),
	string(ometa.CreateAIGovernancePolicyPolicyTypePerformanceStandard),
	string(ometa.CreateAIGovernancePolicyPolicyTypeSecurityControl),
}

// enforcementLevels are the severities an AI governance policy can apply.
var enforcementLevels = []string{
	string(ometa.CreateAIGovernancePolicyEnforcementLevelAdvisory),
	string(ometa.CreateAIGovernancePolicyEnforcementLevelBlocking),
	string(ometa.CreateAIGovernancePolicyEnforcementLevelWarning),
}

var (
	_ resource.Resource                = (*aiGovernancePolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*aiGovernancePolicyResource)(nil)
	_ resource.ResourceWithImportState = (*aiGovernancePolicyResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &aiGovernancePolicyResource{}
}

type aiGovernancePolicyResource struct {
	client *client.Client
}

func (r *aiGovernancePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_governance_policy"
}

func (r *aiGovernancePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata AI governance policy. A policy declares the category " +
			"of AI governance rule, how strictly it is enforced, and whether it is active.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Policy name. Changing it replaces the policy, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"policy_type": rschema.StringAttribute{
				Required:    true,
				Description: "Policy category such as BiasThreshold, ComplianceCheck, CostControl, DataAccess, ModelApproval, PerformanceStandard or SecurityControl.",
				Validators: []validator.String{
					stringvalidator.OneOf(policyTypes...),
				},
			},
			"enforcement_level": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How strictly the policy is enforced: Advisory, Warning or Blocking. The server default applies when omitted.",
				Validators: []validator.String{
					stringvalidator.OneOf(enforcementLevels...),
				},
			},
			"enabled": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the policy is active. Defaults to false.",
				Default:     booldefault.StaticBool(false),
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For an AI governance policy this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the policy.",
				Default:     stringdefault.StaticString(""),
			},
			"version": rschema.Float64Attribute{
				Computed:    true,
				Description: "Server-side entity version.",
			},
			"updated_at": rschema.StringAttribute{
				Computed:    true,
				Description: "Last update time in RFC 3339 format.",
			},
			"updated_by": rschema.StringAttribute{
				Computed:    true,
				Description: "User that performed the last update.",
			},
		},
	}
}

func (r *aiGovernancePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*providerdata.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *providerdata.ProviderData, got %T. This is a bug in the provider.", req.ProviderData),
		)
		return
	}

	r.client = data.Client
}

func (r *aiGovernancePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aiGovernancePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateAIGovernancePolicy{
		Name:        plan.Name.ValueString(),
		PolicyType:  ometa.CreateAIGovernancePolicyPolicyType(plan.PolicyType.ValueString()),
		Enabled:     convert.BoolPointer(plan.Enabled),
		Description: convert.StringPointer(plan.Description),
	}
	if level := convert.StringPointer(plan.EnforcementLevel); level != nil {
		enforcementLevel := ometa.CreateAIGovernancePolicyEnforcementLevel(*level)
		body.EnforcementLevel = &enforcementLevel
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.AIGovernancePolicies.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata AI governance policy",
			fmt.Sprintf("Could not create AI governance policy %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiGovernancePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aiGovernancePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	reference := state.ID.ValueString()
	if reference == "" {
		reference = state.FullyQualifiedName.ValueString()
	}

	found, err := lookupAIGovernancePolicy(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata AI governance policy",
			fmt.Sprintf("Could not read AI governance policy %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aiGovernancePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state aiGovernancePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	ops := patchOps(plan, state)
	if len(ops) > 0 {
		updated, err := sdk.AIGovernancePolicies.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata AI governance policy",
				fmt.Sprintf("Could not update AI governance policy %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupAIGovernancePolicy(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata AI governance policy after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiGovernancePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aiGovernancePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	hardDelete := true
	err = sdk.AIGovernancePolicies.Delete(ctx, state.ID.ValueString(), &ometa.DeleteAIGovernancePolicyParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata AI governance policy",
			fmt.Sprintf("Could not delete AI governance policy %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the AI governance policy UUID or its fully
// qualified name.
func (r *aiGovernancePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupAIGovernancePolicy resolves a UUID or a fully qualified name. Import
// and Read both go through it so either identifier works.
func lookupAIGovernancePolicy(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.AIGovernancePolicy, error) {
	found, err := sdk.AIGovernancePolicies.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.AIGovernancePolicies.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners and rules managed outside
// Terraform are left alone. name is fixed at creation and carries
// RequiresReplace instead.
func patchOps(plan, state aiGovernancePolicyModel) []ometa.JSONPatchOp {
	var ops []ometa.JSONPatchOp

	switch {
	case plan.DisplayName.Equal(state.DisplayName):
	case plan.DisplayName.IsNull():
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/displayName"})
	default:
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/displayName",
			Value: plan.DisplayName.ValueString(),
		})
	}

	if !plan.Description.Equal(state.Description) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/description",
			Value: plan.Description.ValueString(),
		})
	}

	if !plan.PolicyType.Equal(state.PolicyType) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/policyType",
			Value: plan.PolicyType.ValueString(),
		})
	}

	if !plan.EnforcementLevel.Equal(state.EnforcementLevel) {
		if plan.EnforcementLevel.IsNull() || plan.EnforcementLevel.IsUnknown() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/enforcementLevel"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/enforcementLevel",
				Value: plan.EnforcementLevel.ValueString(),
			})
		}
	}

	// enabled carries a schema default, so the planned value is always known
	// and non-null.
	if !plan.Enabled.Equal(state.Enabled) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/enabled",
			Value: plan.Enabled.ValueBool(),
		})
	}

	return ops
}
