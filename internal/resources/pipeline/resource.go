package pipeline

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	_ resource.Resource                = (*pipelineResource)(nil)
	_ resource.ResourceWithConfigure   = (*pipelineResource)(nil)
	_ resource.ResourceWithImportState = (*pipelineResource)(nil)
)

var pipelineStates = []string{
	string(ometa.CreatePipelineStateActive),
	string(ometa.CreatePipelineStateInactive),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &pipelineResource{}
}

type pipelineResource struct {
	client *client.Client
}

func (r *pipelineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (r *pipelineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata pipeline. A pipeline belongs to a pipeline service and " +
			"owns the list of tasks that describe its stages.",
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
				Description: "Pipeline name. Changing it replaces the pipeline, because the name is part of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"service": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the parent pipeline service. Changing it replaces the pipeline.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name, in the form service.pipeline.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the pipeline.",
				Default:     stringdefault.StaticString(""),
			},
			"concurrency": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of parallel task runs. Filled by the server when omitted.",
			},
			"pipeline_location": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Location of the pipeline definition, for example a Git repository path.",
			},
			"schedule_interval": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cron expression that schedules the pipeline. The server keeps it at creation, so changing it replaces the pipeline.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_url": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Link to the pipeline definition in the orchestrator.",
			},
			"start_date": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Pipeline start time as epoch milliseconds. The server keeps it at creation, so changing it replaces the pipeline.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"state": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Scheduling state, either Active or Inactive.",
				Default:     stringdefault.StaticString(string(ometa.CreatePipelineStateActive)),
				Validators: []validator.String{
					stringvalidator.OneOf(pipelineStates...),
				},
			},
			"tasks": rschema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tasks that make up the pipeline. Changing the list replaces the pipeline, because the server only applies task membership and description changes.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"name": rschema.StringAttribute{
							Required:    true,
							Description: "Task name.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"display_name": rschema.StringAttribute{
							Optional:    true,
							Description: "Human-readable task name.",
						},
						"description": rschema.StringAttribute{
							Optional:    true,
							Description: "Markdown description of the task.",
						},
						"task_type": rschema.StringAttribute{
							Optional:    true,
							Description: "Task type reported by the orchestrator.",
						},
						"task_sql": rschema.StringAttribute{
							Optional:    true,
							Description: "SQL statement executed by the task.",
						},
						"source_url": rschema.StringAttribute{
							Optional:    true,
							Description: "Link to the task definition in the orchestrator.",
						},
						"start_date": rschema.StringAttribute{
							Optional:    true,
							Description: "Task start time reported by the orchestrator.",
						},
						"end_date": rschema.StringAttribute{
							Optional:    true,
							Description: "Task end time reported by the orchestrator.",
						},
						"downstream_tasks": rschema.ListAttribute{
							Optional:    true,
							Description: "Names of tasks that run after this one.",
							ElementType: types.StringType,
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

func (r *pipelineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pipelineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	tasks, diags := tasksToAPI(ctx, plan.Tasks)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreatePipeline{
		Name:             plan.Name.ValueString(),
		Service:          plan.Service.ValueString(),
		Description:      convert.StringPointer(plan.Description),
		DisplayName:      convert.StringPointer(plan.DisplayName),
		Concurrency:      int32Pointer(plan.Concurrency),
		PipelineLocation: convert.StringPointer(plan.PipelineLocation),
		ScheduleInterval: convert.StringPointer(plan.ScheduleInterval),
		SourceUrl:        convert.StringPointer(plan.SourceURL),
		StartDate:        int64Pointer(plan.StartDate),
		State:            statePointer(plan.State),
	}
	if tasks != nil {
		body.Tasks = &tasks
	}

	created, err := sdk.Pipelines.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata pipeline",
			fmt.Sprintf("Could not create pipeline %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pipelineModel
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

	found, err := lookupPipeline(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata pipeline",
			fmt.Sprintf("Could not read pipeline %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state pipelineModel
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

	ops, diags := patchOps(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(ops) > 0 {
		updated, err := sdk.Pipelines.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata pipeline",
				fmt.Sprintf("Could not update pipeline %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		resp.Diagnostics.Append(apply(ctx, updated, &plan)...)
	} else {
		found, err := lookupPipeline(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata pipeline after update", client.ErrorDetail(err))
			return
		}
		resp.Diagnostics.Append(apply(ctx, found, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pipelineModel
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
	err = sdk.Pipelines.Delete(ctx, state.ID.ValueString(), &ometa.DeletePipelineParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata pipeline",
			fmt.Sprintf("Could not delete pipeline %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the pipeline UUID or its fully qualified name.
func (r *pipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupPipeline resolves a UUID or a fully qualified name. Read and Import
// both go through it so either identifier works. Tasks are requested
// explicitly, because the API omits them by default.
func lookupPipeline(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Pipeline, error) {
	fields := "tasks"
	found, err := sdk.Pipelines.GetByID(ctx, reference, &ometa.GetPipelineWithIDParams{
		Fields: &fields,
	})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Pipelines.GetByName(ctx, reference, &ometa.GetPipelineByFQNParams{
		Fields: &fields,
	})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners and lineage managed outside
// Terraform are left alone. name and service carry RequiresReplace and are
// never patched.
func patchOps(ctx context.Context, plan, state pipelineModel) ([]ometa.JSONPatchOp, diag.Diagnostics) {
	var ops []ometa.JSONPatchOp
	var diags diag.Diagnostics

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

	if !plan.Concurrency.IsUnknown() && !plan.Concurrency.Equal(state.Concurrency) {
		if plan.Concurrency.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/concurrency"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/concurrency",
				Value: plan.Concurrency.ValueInt64(),
			})
		}
	}

	ops = appendOptionalString(ops, "/pipelineLocation", plan.PipelineLocation, state.PipelineLocation)
	ops = appendOptionalString(ops, "/sourceUrl", plan.SourceURL, state.SourceURL)
	ops = appendOptionalString(ops, "/state", plan.State, state.State)

	// schedule_interval, start_date and tasks are not fully updated by the
	// server. They carry RequiresReplace instead.

	return ops, diags
}

// appendOptionalString adds a patch operation for an optional string attribute.
// Unknown plan values are skipped.
func appendOptionalString(ops []ometa.JSONPatchOp, path string, plan, state types.String) []ometa.JSONPatchOp {
	if plan.IsUnknown() || plan.Equal(state) {
		return ops
	}
	if plan.IsNull() {
		return append(ops, ometa.JSONPatchOp{Op: "remove", Path: path})
	}
	return append(ops, ometa.JSONPatchOp{Op: "add", Path: path, Value: plan.ValueString()})
}

// statePointer maps the scheduling state to the SDK enum.
func statePointer(value types.String) *ometa.CreatePipelineState {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	state := ometa.CreatePipelineState(value.ValueString())
	return &state
}
