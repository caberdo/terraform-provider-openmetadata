package learning_resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*learningResourceResource)(nil)
	_ resource.ResourceWithConfigure   = (*learningResourceResource)(nil)
	_ resource.ResourceWithImportState = (*learningResourceResource)(nil)
)

var (
	learningResourceTypes = []string{
		string(ometa.CreateLearningResourceResourceTypeArticle),
		string(ometa.CreateLearningResourceResourceTypeVideo),
		string(ometa.CreateLearningResourceResourceTypeStorylane),
	}
	learningResourceCategories = []string{
		string(ometa.CreateLearningResourceCategoriesAI),
		string(ometa.CreateLearningResourceCategoriesAdministration),
		string(ometa.CreateLearningResourceCategoriesDataGovernance),
		string(ometa.CreateLearningResourceCategoriesDataQuality),
		string(ometa.CreateLearningResourceCategoriesDiscovery),
		string(ometa.CreateLearningResourceCategoriesObservability),
	}
	learningResourceDifficulties = []string{
		string(ometa.CreateLearningResourceDifficultyIntro),
		string(ometa.CreateLearningResourceDifficultyIntermediate),
		string(ometa.CreateLearningResourceDifficultyAdvanced),
	}
	learningResourceStatuses = []string{
		string(ometa.CreateLearningResourceStatusActive),
		string(ometa.CreateLearningResourceStatusDraft),
		string(ometa.CreateLearningResourceStatusDeprecated),
	}
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &learningResourceResource{}
}

type learningResourceResource struct {
	client *client.Client
}

func (r *learningResourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_learning_resource"
}

func (r *learningResourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata learning resource. A learning resource is a piece of " +
			"academy content, such as an article or a video, tagged with the categories and UI " +
			"contexts where it should appear.",
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
				Description: "Learning resource name. Changing it replaces the resource, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"resource_type": rschema.StringAttribute{
				Required:    true,
				Description: "Content type, one of Article, Video or Storylane.",
				Validators: []validator.String{
					stringvalidator.OneOf(learningResourceTypes...),
				},
			},
			"categories": rschema.ListAttribute{
				Required:    true,
				Description: "Academy categories the resource belongs to.",
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.OneOf(learningResourceCategories...)),
				},
			},
			"contexts": rschema.ListNestedAttribute{
				Required:    true,
				Description: "UI contexts the resource appears in.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"page_id": rschema.StringAttribute{
							Required:    true,
							Description: "Page the resource is shown on.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"component_id": rschema.StringAttribute{
							Optional:    true,
							Description: "Component on the page the resource is attached to.",
						},
						"priority": rschema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "Ordering priority inside the context. Filled by the server when omitted.",
						},
					},
				},
			},
			"source": rschema.ListNestedAttribute{
				Required:    true,
				Description: "Location of the content. Exactly one source must be set.",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"url": rschema.StringAttribute{
							Required:    true,
							Description: "URL of the content.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"provider": rschema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "System that provides the content, for example user or system. Filled by the server when omitted.",
						},
					},
				},
			},
			"difficulty": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Difficulty level, one of Intro, Intermediate or Advanced.",
				Validators: []validator.String{
					stringvalidator.OneOf(learningResourceDifficulties...),
				},
			},
			"estimated_duration": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Estimated time to finish the resource, in minutes.",
			},
			"completion_threshold": rschema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Share of the resource that must be completed to mark it done.",
			},
			"status": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Publication status, one of Active, Draft or Deprecated.",
				Validators: []validator.String{
					stringvalidator.OneOf(learningResourceStatuses...),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a learning resource this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the learning resource.",
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

func (r *learningResourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *learningResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan learningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body, diags := createBody(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := sdk.LearningResources.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata learning resource",
			fmt.Sprintf("Could not create learning resource %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *learningResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state learningResourceModel
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

	found, err := lookupLearningResource(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata learning resource",
			fmt.Sprintf("Could not read learning resource %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *learningResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state learningResourceModel
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
		updated, err := sdk.LearningResources.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata learning resource",
				fmt.Sprintf("Could not update learning resource %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		resp.Diagnostics.Append(apply(ctx, updated, &plan)...)
	} else {
		found, err := lookupLearningResource(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata learning resource after update", client.ErrorDetail(err))
			return
		}
		resp.Diagnostics.Append(apply(ctx, found, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *learningResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state learningResourceModel
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
	err = sdk.LearningResources.Delete(ctx, state.ID.ValueString(), &ometa.DeleteLearningResourceParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata learning resource",
			fmt.Sprintf("Could not delete learning resource %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the learning resource UUID or its fully qualified
// name.
func (r *learningResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupLearningResource resolves a UUID or a fully qualified name. Read and
// Import both go through it so either identifier works.
func lookupLearningResource(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.LearningResource, error) {
	found, err := sdk.LearningResources.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.LearningResources.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners and reviewers managed outside
// Terraform are left alone. name carries RequiresReplace and is never patched.
func patchOps(ctx context.Context, plan, state learningResourceModel) ([]ometa.JSONPatchOp, diag.Diagnostics) {
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

	if !plan.ResourceType.Equal(state.ResourceType) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/resourceType",
			Value: plan.ResourceType.ValueString(),
		})
	}

	if !plan.Categories.IsUnknown() && !plan.Categories.Equal(state.Categories) {
		categories, categoryDiags := categoriesToAPI(ctx, plan.Categories)
		diags.Append(categoryDiags...)
		if diags.HasError() {
			return nil, diags
		}
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/categories", Value: categories})
	}

	if !plan.Contexts.IsUnknown() && !plan.Contexts.Equal(state.Contexts) {
		contexts, contextDiags := contextsToAPI(ctx, plan.Contexts)
		diags.Append(contextDiags...)
		if diags.HasError() {
			return nil, diags
		}
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/contexts", Value: contexts})
	}

	if !plan.Source.IsUnknown() && !plan.Source.Equal(state.Source) {
		source, sourceDiags := sourceToAPI(ctx, plan.Source)
		diags.Append(sourceDiags...)
		if diags.HasError() {
			return nil, diags
		}
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/source", Value: source})
	}

	ops = appendOptionalString(ops, "/difficulty", plan.Difficulty, state.Difficulty)
	ops = appendOptionalString(ops, "/status", plan.Status, state.Status)
	ops = appendIntIfChanged(ops, "/estimatedDuration", plan.EstimatedDuration, state.EstimatedDuration)

	if !plan.CompletionThreshold.IsUnknown() && !plan.CompletionThreshold.Equal(state.CompletionThreshold) {
		if plan.CompletionThreshold.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/completionThreshold"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/completionThreshold",
				Value: plan.CompletionThreshold.ValueFloat64(),
			})
		}
	}

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

// appendIntIfChanged adds a patch operation for an optional number attribute.
func appendIntIfChanged(ops []ometa.JSONPatchOp, path string, plan, state types.Int64) []ometa.JSONPatchOp {
	if plan.IsUnknown() || plan.Equal(state) {
		return ops
	}
	if plan.IsNull() {
		return append(ops, ometa.JSONPatchOp{Op: "remove", Path: path})
	}
	return append(ops, ometa.JSONPatchOp{Op: "add", Path: path, Value: plan.ValueInt64()})
}
