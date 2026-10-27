package workflow_definition

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*workflowDefinitionResource)(nil)
	_ resource.ResourceWithConfigure   = (*workflowDefinitionResource)(nil)
	_ resource.ResourceWithImportState = (*workflowDefinitionResource)(nil)
)

// workflowDefinitionTypes are the workflow definition kinds accepted by the
// API. Automation workflows are created by OpenMetadata itself, so only the
// two authorable kinds are exposed.
var workflowDefinitionTypes = []string{
	string(ometa.EventBasedEntity),
	string(ometa.NoOp),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &workflowDefinitionResource{}
}

type workflowDefinitionResource struct {
	client *client.Client
}

func (r *workflowDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_definition"
}

func (r *workflowDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata workflow definition. A workflow definition wires " +
			"nodes and edges to a trigger and drives automations such as approval flows.",
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
				Description: "Workflow definition name. Changing it replaces the definition, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a workflow definition this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Required:    true,
				Description: "Markdown description of the workflow definition. The API requires it, so it has no default.",
			},
			"type": rschema.StringAttribute{
				Optional:    true,
				Description: "Workflow kind: eventBasedEntity or noOp. The API does not return it on read, so it is not computed.",
				Validators: []validator.String{
					stringvalidator.OneOf(workflowDefinitionTypes...),
				},
			},
			"store_stage_status": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the workflow stores the status of every stage it runs.",
				Default:     booldefault.StaticBool(false),
			},
			"nodes": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Workflow nodes as a JSON array. Each node carries at least name and type; the accepted properties depend on the node type.",
				Validators: []validator.String{
					jsonAttributeValidator{kind: "array"},
				},
				PlanModifiers: []planmodifier.String{
					semanticJSONPlanModifier(),
				},
			},
			"edges": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Workflow edges as a JSON array of objects with from and to nodes, plus an optional condition.",
				Validators: []validator.String{
					jsonAttributeValidator{kind: "array"},
				},
				PlanModifiers: []planmodifier.String{
					semanticJSONPlanModifier(),
				},
			},
			"trigger": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Workflow trigger as a JSON object with a type and a type-specific config.",
				Validators: []validator.String{
					jsonAttributeValidator{kind: "object"},
				},
				PlanModifiers: []planmodifier.String{
					semanticJSONPlanModifier(),
				},
			},
			"deployed": rschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the workflow definition is deployed on the server.",
			},
			"suspended": rschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the workflow definition is suspended on the server.",
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

func (r *workflowDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workflowDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	nodes, err := decodeNodes(plan.Nodes)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("nodes"), "Invalid nodes JSON", err.Error())
		return
	}

	edges, err := decodeEdges(plan.Edges)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("edges"), "Invalid edges JSON", err.Error())
		return
	}

	trigger, err := decodeTrigger(plan.Trigger)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("trigger"), "Invalid trigger JSON", err.Error())
		return
	}
	if trigger == nil {
		workflowType := plan.Type.ValueString()
		if workflowType != "" && workflowType != "noOp" {
			resp.Diagnostics.AddAttributeError(
				path.Root("trigger"),
				"Missing workflow trigger",
				fmt.Sprintf("trigger is required for workflow type %q.", workflowType),
			)
			return
		}
		noOp := "noOp"
		trigger = &ometa.WorkflowTriggerInterface{Type: &noOp}
	}

	body := &ometa.CreateWorkflowDefinition{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Config: &ometa.WorkflowConfiguration{
			StoreStageStatus: plan.StoreStageStatus.ValueBool(),
		},
		Nodes:   nodes,
		Edges:   edges,
		Trigger: trigger,
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() {
		definitionType := ometa.CreateWorkflowDefinitionType(plan.Type.ValueString())
		body.Type = &definitionType
	}

	created, err := sdk.WorkflowDefinitions.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata workflow definition",
			fmt.Sprintf("Could not create workflow definition %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowDefinitionModel
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

	found, err := lookupWorkflowDefinition(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata workflow definition",
			fmt.Sprintf("Could not read workflow definition %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workflowDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state workflowDefinitionModel
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
		updated, err := sdk.WorkflowDefinitions.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata workflow definition",
				fmt.Sprintf("Could not update workflow definition %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupWorkflowDefinition(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata workflow definition after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workflowDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowDefinitionModel
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
	recursive := true
	err = sdk.WorkflowDefinitions.Delete(ctx, state.ID.ValueString(), &ometa.DeleteWorkflowDefinitionParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata workflow definition",
			fmt.Sprintf("Could not delete workflow definition %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the workflow definition UUID or its fully
// qualified name.
func (r *workflowDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupWorkflowDefinition resolves a UUID or a fully qualified name. Import
// and Read both go through it so either identifier works.
func lookupWorkflowDefinition(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.WorkflowDefinition, error) {
	found, err := sdk.WorkflowDefinitions.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.WorkflowDefinitions.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners and reviews managed outside
// Terraform are left alone.
func patchOps(plan, state workflowDefinitionModel) []ometa.JSONPatchOp {
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

	if !plan.Type.Equal(state.Type) {
		if plan.Type.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/type"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/type",
				Value: ometa.CreateWorkflowDefinitionType(plan.Type.ValueString()),
			})
		}
	}

	if !plan.StoreStageStatus.IsUnknown() && !state.StoreStageStatus.IsUnknown() &&
		!plan.StoreStageStatus.Equal(state.StoreStageStatus) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:   "add",
			Path: "/config",
			Value: &ometa.WorkflowConfiguration{
				StoreStageStatus: plan.StoreStageStatus.ValueBool(),
			},
		})
	}

	ops = appendJSONOp(ops, "/nodes", plan.Nodes, state.Nodes, func(value types.String) (any, bool) {
		decoded, err := decodeNodes(value)
		return decoded, err == nil
	})
	ops = appendJSONOp(ops, "/edges", plan.Edges, state.Edges, func(value types.String) (any, bool) {
		decoded, err := decodeEdges(value)
		return decoded, err == nil
	})
	ops = appendJSONOp(ops, "/trigger", plan.Trigger, state.Trigger, func(value types.String) (any, bool) {
		decoded, err := decodeTrigger(value)
		return decoded, err == nil
	})

	// name is not patchable: it is the fully qualified name, so the attribute
	// carries RequiresReplace instead.

	return ops
}

// appendJSONOp adds a patch operation for a JSON attribute, keeping state
// churn out when the two documents are semantically equal.
func appendJSONOp(ops []ometa.JSONPatchOp, path string, plan, state types.String, decode func(types.String) (any, bool)) []ometa.JSONPatchOp {
	if plan.IsUnknown() || state.IsUnknown() {
		return ops
	}

	switch {
	case plan.IsNull() && state.IsNull():
		return ops
	case plan.IsNull():
		return append(ops, ometa.JSONPatchOp{Op: "remove", Path: path})
	case state.IsNull() || !jsonEqual(plan.ValueString(), state.ValueString()):
		value, ok := decode(plan)
		if !ok {
			return ops
		}
		return append(ops, ometa.JSONPatchOp{Op: "add", Path: path, Value: value})
	}

	return ops
}

// semanticJSONPlanModifier keeps the prior value when the planned document is
// semantically equal to state, so a reformatted but equivalent JSON attribute
// does not produce a perpetual diff.
func semanticJSONPlanModifier() planmodifier.String {
	return jsonEquivalentModifier{}
}

type jsonEquivalentModifier struct{}

func (m jsonEquivalentModifier) Description(_ context.Context) string {
	return "Keeps the prior value when the planned value is JSON-equal to state."
}

func (m jsonEquivalentModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m jsonEquivalentModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	if jsonEqual(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// jsonAttributeValidator rejects values that are not a JSON document of the
// expected shape, so a typo fails during plan instead of during apply.
type jsonAttributeValidator struct {
	kind string // "array" or "object"
}

func (v jsonAttributeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be a JSON %s", v.kind)
}

func (v jsonAttributeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonAttributeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := strings.TrimSpace(req.ConfigValue.ValueString())
	if value == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", "The value must not be empty.")
		return
	}
	if !json.Valid([]byte(value)) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", "The value must be a valid JSON document.")
		return
	}

	switch v.kind {
	case "array":
		if !strings.HasPrefix(value, "[") {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", "The value must be a JSON array.")
		}
	case "object":
		if !strings.HasPrefix(value, "{") {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON", "The value must be a JSON object.")
		}
	}
}
