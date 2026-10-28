package ai_application

import (
	"context"
	"fmt"

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

// applicationTypes are the AI application kinds OpenMetadata accepts.
var applicationTypes = []string{
	string(ometa.CreateAIApplicationApplicationTypeAgent),
	string(ometa.CreateAIApplicationApplicationTypeAssistant),
	string(ometa.CreateAIApplicationApplicationTypeAutomationBot),
	string(ometa.CreateAIApplicationApplicationTypeChatbot),
	string(ometa.CreateAIApplicationApplicationTypeCodeGenerator),
	string(ometa.CreateAIApplicationApplicationTypeCopilot),
	string(ometa.CreateAIApplicationApplicationTypeCustom),
	string(ometa.CreateAIApplicationApplicationTypeDataAnalyst),
	string(ometa.CreateAIApplicationApplicationTypeMultiAgent),
	string(ometa.CreateAIApplicationApplicationTypeRAG),
}

// developmentStages are the lifecycle stages OpenMetadata accepts.
var developmentStages = []string{
	string(ometa.CreateAIApplicationDevelopmentStageDeprecated),
	string(ometa.CreateAIApplicationDevelopmentStageDevelopment),
	string(ometa.CreateAIApplicationDevelopmentStageProduction),
	string(ometa.CreateAIApplicationDevelopmentStageProposal),
	string(ometa.CreateAIApplicationDevelopmentStageStaging),
	string(ometa.CreateAIApplicationDevelopmentStageTesting),
	string(ometa.CreateAIApplicationDevelopmentStageUnauthorized),
}

var (
	_ resource.Resource                = (*aiApplicationResource)(nil)
	_ resource.ResourceWithConfigure   = (*aiApplicationResource)(nil)
	_ resource.ResourceWithImportState = (*aiApplicationResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &aiApplicationResource{}
}

type aiApplicationResource struct {
	client *client.Client
}

func (r *aiApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ai_application"
}

func (r *aiApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata AI application. An AI application records a deployed " +
			"assistant, chatbot or copilot together with its development stage and documentation.",
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
				Description: "AI application name. Changing it replaces the application, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"application_type": rschema.StringAttribute{
				Required:    true,
				Description: "Kind of AI application, for example Chatbot or Copilot.",
				Validators: []validator.String{
					stringvalidator.OneOf(applicationTypes...),
				},
			},
			"deployment_url": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL where the application is deployed.",
				Default:     stringdefault.StaticString(""),
			},
			"development_stage": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Lifecycle stage of the application, for example Production or Staging. Unset when the server holds no stage.",
				Validators: []validator.String{
					stringvalidator.OneOf(developmentStages...),
				},
			},
			"documentation": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown documentation of the application.",
				Default:     stringdefault.StaticString(""),
			},
			"source_code": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL or reference to the application source code.",
				Default:     stringdefault.StaticString(""),
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For an AI application this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the application.",
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

func (r *aiApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *aiApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aiApplicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateAIApplication{
		Name:             plan.Name.ValueString(),
		ApplicationType:  ometa.CreateAIApplicationApplicationType(plan.ApplicationType.ValueString()),
		DeploymentUrl:    convert.StringPointer(plan.DeploymentURL),
		DevelopmentStage: developmentStagePointer(plan.DevelopmentStage),
		Documentation:    convert.StringPointer(plan.Documentation),
		SourceCode:       convert.StringPointer(plan.SourceCode),
		Description:      convert.StringPointer(plan.Description),
		DisplayName:      convert.StringPointer(plan.DisplayName),
	}

	created, err := sdk.AIApplications.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata AI application",
			fmt.Sprintf("Could not create AI application %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aiApplicationModel
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

	found, err := lookupAIApplication(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata AI application",
			fmt.Sprintf("Could not read AI application %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aiApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state aiApplicationModel
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
		updated, err := sdk.AIApplications.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata AI application",
				fmt.Sprintf("Could not update AI application %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupAIApplication(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata AI application after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aiApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aiApplicationModel
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
	err = sdk.AIApplications.Delete(ctx, state.ID.ValueString(), &ometa.DeleteAIApplicationParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata AI application",
			fmt.Sprintf("Could not delete AI application %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the AI application UUID or its fully qualified name.
func (r *aiApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupAIApplication resolves a UUID or a fully qualified name. Import and
// Read both go through it so either identifier works.
func lookupAIApplication(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.AIApplication, error) {
	found, err := sdk.AIApplications.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.AIApplications.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners, and reviewers managed outside
// Terraform are left alone. name is fixed at creation and carries
// RequiresReplace instead.
func patchOps(plan, state aiApplicationModel) []ometa.JSONPatchOp {
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

	ops = appendOptionalString(ops, "/description", plan.Description, state.Description)
	ops = appendOptionalString(ops, "/deploymentUrl", plan.DeploymentURL, state.DeploymentURL)
	ops = appendOptionalString(ops, "/developmentStage", plan.DevelopmentStage, state.DevelopmentStage)
	ops = appendOptionalString(ops, "/documentation", plan.Documentation, state.Documentation)
	ops = appendOptionalString(ops, "/sourceCode", plan.SourceCode, state.SourceCode)
	ops = appendOptionalString(ops, "/applicationType", plan.ApplicationType, state.ApplicationType)

	return ops
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

// developmentStagePointer maps the lifecycle stage to the SDK enum, omitting
// the field when Terraform has no value.
func developmentStagePointer(value types.String) *ometa.CreateAIApplicationDevelopmentStage {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	stage := ometa.CreateAIApplicationDevelopmentStage(value.ValueString())
	return &stage
}
