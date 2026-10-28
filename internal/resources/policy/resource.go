package policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*policyResource)(nil)
	_ resource.ResourceWithConfigure   = (*policyResource)(nil)
	_ resource.ResourceWithImportState = (*policyResource)(nil)
)

// ruleEffects lists the effects a policy rule accepts.
var ruleEffects = []string{
	string(ometa.RuleEffectAllow),
	string(ometa.RuleEffectDeny),
}

// ruleOperations lists every operation an OpenMetadata rule can grant or deny.
var ruleOperations = []string{
	string(ometa.RuleOperationsAll),
	string(ometa.RuleOperationsAuditLogs),
	string(ometa.RuleOperationsBulkCreate),
	string(ometa.RuleOperationsBulkUpdate),
	string(ometa.RuleOperationsCreate),
	string(ometa.RuleOperationsCreateIngestionPipelineAutomator),
	string(ometa.RuleOperationsCreateScim),
	string(ometa.RuleOperationsCreateTests),
	string(ometa.RuleOperationsDelete),
	string(ometa.RuleOperationsDeleteScim),
	string(ometa.RuleOperationsDeleteTestCaseFailedRowsSample),
	string(ometa.RuleOperationsDeploy),
	string(ometa.RuleOperationsEditAll),
	string(ometa.RuleOperationsEditCertification),
	string(ometa.RuleOperationsEditCustomFields),
	string(ometa.RuleOperationsEditDataProfile),
	string(ometa.RuleOperationsEditDescription),
	string(ometa.RuleOperationsEditDisplayName),
	string(ometa.RuleOperationsEditEntityRelationship),
	string(ometa.RuleOperationsEditGlossaryTerms),
	string(ometa.RuleOperationsEditIngestionPipelineStatus),
	string(ometa.RuleOperationsEditKnowledgePanel),
	string(ometa.RuleOperationsEditLifeCycle),
	string(ometa.RuleOperationsEditLineage),
	string(ometa.RuleOperationsEditOwners),
	string(ometa.RuleOperationsEditPage),
	string(ometa.RuleOperationsEditPolicy),
	string(ometa.RuleOperationsEditQueries),
	string(ometa.RuleOperationsEditReviewers),
	string(ometa.RuleOperationsEditRole),
	string(ometa.RuleOperationsEditSampleData),
	string(ometa.RuleOperationsEditScim),
	string(ometa.RuleOperationsEditStatus),
	string(ometa.RuleOperationsEditTags),
	string(ometa.RuleOperationsEditTeams),
	string(ometa.RuleOperationsEditTestDefinitionLibrary),
	string(ometa.RuleOperationsEditTests),
	string(ometa.RuleOperationsEditTier),
	string(ometa.RuleOperationsEditUsage),
	string(ometa.RuleOperationsEditUserNotificationTemplate),
	string(ometa.RuleOperationsEditUsers),
	string(ometa.RuleOperationsGenerateToken),
	string(ometa.RuleOperationsImpersonate),
	string(ometa.RuleOperationsKill),
	string(ometa.RuleOperationsTrigger),
	string(ometa.RuleOperationsViewAll),
	string(ometa.RuleOperationsViewBasic),
	string(ometa.RuleOperationsViewCustomFields),
	string(ometa.RuleOperationsViewDataProfile),
	string(ometa.RuleOperationsViewProfilerGlobalConfiguration),
	string(ometa.RuleOperationsViewQueries),
	string(ometa.RuleOperationsViewSampleData),
	string(ometa.RuleOperationsViewScim),
	string(ometa.RuleOperationsViewTestCaseFailedRowsSample),
	string(ometa.RuleOperationsViewTestDefinitionLibrary),
	string(ometa.RuleOperationsViewTests),
	string(ometa.RuleOperationsViewUsage),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &policyResource{}
}

type policyResource struct {
	client *client.Client
}

func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata policy. A policy grants or denies a set of operations on " +
			"entity resources and is attached to roles.",
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
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a policy this equals name.",
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
			"enabled": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the policy is active. The server defaults this to true.",
			},
			"rules": rschema.SetNestedAttribute{
				Required:    true,
				Description: "Rules granted by the policy. OpenMetadata orders rules and their operations itself, so this is a set. Updating the set replaces the whole rule set.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"name": rschema.StringAttribute{
							Required:    true,
							Description: "Rule name, unique within the policy.",
						},
						"effect": rschema.StringAttribute{
							Required:    true,
							Description: "Whether the rule allows or denies the operations.",
							Validators: []validator.String{
								stringvalidator.OneOf(ruleEffects...),
							},
						},
						"operations": rschema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "Operations covered by the rule, for example ViewAll or EditDescription. The server sorts them.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								setvalidator.ValueStringsAre(stringvalidator.OneOf(ruleOperations...)),
							},
						},
						"resources": rschema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "Entity types the rule applies to, for example table, dashboard or pipeline. Use the canonical \"All\" instead of lowercase \"all\"; \"All\" cannot be combined with other resources.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
								policyResourcesValidator{},
							},
						},
						"condition": rschema.StringAttribute{
							Optional:    true,
							Description: "Optional condition expression that narrows when the rule applies.",
						},
						"description": rschema.StringAttribute{
							Optional:    true,
							Description: "Optional description of the rule.",
						},
					},
				},
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

func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	rules, err := rulesToAPI(ctx, plan.Rules)
	if err != nil {
		resp.Diagnostics.AddError("Invalid policy rules", err.Error())
		return
	}

	body := &ometa.CreatePolicy{
		Name:        plan.Name.ValueString(),
		Description: convert.StringPointer(plan.Description),
		DisplayName: convert.StringPointer(plan.DisplayName),
		Enabled:     convert.BoolPointer(plan.Enabled),
		Rules:       rules,
	}

	created, err := sdk.Policies.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata policy",
			fmt.Sprintf("Could not create policy %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyModel
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

	found, err := lookupPolicy(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata policy",
			fmt.Sprintf("Could not read policy %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state policyModel
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

	ops, err := patchOps(ctx, plan, state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid policy rules", err.Error())
		return
	}

	if len(ops) > 0 {
		updated, err := sdk.Policies.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata policy",
				fmt.Sprintf("Could not update policy %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupPolicy(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata policy after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyModel
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
	err = sdk.Policies.Delete(ctx, state.ID.ValueString(), &ometa.DeletePolicyParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata policy",
			fmt.Sprintf("Could not delete policy %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the policy UUID or its fully qualified name.
func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupPolicy resolves a UUID or a fully qualified name. Import and Read both
// go through it so either identifier works.
func lookupPolicy(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Policy, error) {
	found, err := sdk.Policies.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Policies.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, tags, and roles managed outside
// Terraform are left alone.
func patchOps(ctx context.Context, plan, state policyModel) ([]ometa.JSONPatchOp, error) {
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

	if !plan.Enabled.Equal(state.Enabled) {
		if plan.Enabled.IsNull() || plan.Enabled.IsUnknown() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/enabled"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/enabled",
				Value: plan.Enabled.ValueBool(),
			})
		}
	}

	// name cannot be patched: the server treats it as the fully qualified name,
	// so the attribute carries RequiresReplace instead. Rules are a set because
	// the server orders rules and operations itself.
	if !plan.Rules.IsUnknown() && !plan.Rules.Equal(state.Rules) {
		rules, err := rulesToAPI(ctx, plan.Rules)
		if err != nil {
			return nil, err
		}
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/rules",
			Value: rules,
		})
	}

	return ops, nil
}

// policyResourcesValidator rejects resource sets that OpenMetadata would
// silently rewrite: a lowercase "all" is normalized to "All", and "All"
// together with any other resource collapses to just "All". Either rewrite
// would leave state different from the configuration.
type policyResourcesValidator struct{}

var _ validator.Set = policyResourcesValidator{}

func (v policyResourcesValidator) Description(_ context.Context) string {
	return `Resource sets must not contain lowercase "all", and "All" cannot be combined with other resources.`
}

func (v policyResourcesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v policyResourcesValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var values []string
	if diags := req.ConfigValue.ElementsAs(ctx, &values, false); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	hasAll := false
	for _, value := range values {
		switch value {
		case "all":
			resp.Diagnostics.AddAttributeError(
				req.Path.AtSetValue(types.StringValue(value)),
				"Invalid policy resource",
				`Use the canonical "All" instead of lowercase "all"; the server would rewrite it.`,
			)
		case "All":
			hasAll = true
		}
	}

	if hasAll && len(values) > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid policy resource",
			`"All" cannot be combined with other resources; the server would drop them.`,
		)
	}
}
