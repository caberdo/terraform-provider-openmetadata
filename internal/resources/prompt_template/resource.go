package prompt_template

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
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

// promptTemplateTypes are the template types OpenMetadata accepts.
var promptTemplateTypes = []string{
	string(ometa.CreatePromptTemplateTemplateTypeChatCompletion),
	string(ometa.CreatePromptTemplateTemplateTypeClassification),
	string(ometa.CreatePromptTemplateTemplateTypeCodeGeneration),
	string(ometa.CreatePromptTemplateTemplateTypeCustom),
	string(ometa.CreatePromptTemplateTemplateTypeEmbedding),
	string(ometa.CreatePromptTemplateTemplateTypeExtraction),
	string(ometa.CreatePromptTemplateTemplateTypeTextGeneration),
}

var (
	_ resource.Resource                = (*promptTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*promptTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*promptTemplateResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &promptTemplateResource{}
}

type promptTemplateResource struct {
	client *client.Client
}

func (r *promptTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt_template"
}

func (r *promptTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata prompt template. A prompt template stores the text and " +
			"system prompt used by an AI application or LLM model.",
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
				Description: "Prompt template name. Changing it replaces the template, because the name is the basis of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"template_content": rschema.StringAttribute{
				Required:    true,
				Description: "Template body. Placeholders use the OpenMetadata {{variable}} syntax.",
			},
			"template_type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Template category, for example Classification or ChatCompletion.",
				Validators: []validator.String{
					stringvalidator.OneOf(promptTemplateTypes...),
				},
			},
			"template_version": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Caller-managed template version string.",
			},
			"system_prompt": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "System prompt sent ahead of the template content.",
				Default:     stringdefault.StaticString(""),
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a prompt template this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the prompt template.",
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

func (r *promptTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *promptTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan promptTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreatePromptTemplate{
		Name:            plan.Name.ValueString(),
		TemplateContent: plan.TemplateContent.ValueString(),
		TemplateVersion: convert.StringPointer(plan.TemplateVersion),
		SystemPrompt:    convert.StringPointer(plan.SystemPrompt),
		Description:     convert.StringPointer(plan.Description),
	}
	if !plan.TemplateType.IsNull() && plan.TemplateType.ValueString() != "" {
		templateType := ometa.CreatePromptTemplateTemplateType(plan.TemplateType.ValueString())
		body.TemplateType = &templateType
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.PromptTemplates.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata prompt template",
			fmt.Sprintf("Could not create prompt template %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *promptTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state promptTemplateModel
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

	found, err := lookupPromptTemplate(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata prompt template",
			fmt.Sprintf("Could not read prompt template %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *promptTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state promptTemplateModel
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
		updated, err := sdk.PromptTemplates.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata prompt template",
				fmt.Sprintf("Could not update prompt template %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupPromptTemplate(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata prompt template after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *promptTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state promptTemplateModel
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
	err = sdk.PromptTemplates.Delete(ctx, state.ID.ValueString(), &ometa.DeletePromptTemplateParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata prompt template",
			fmt.Sprintf("Could not delete prompt template %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the prompt template UUID or its fully qualified name.
func (r *promptTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupPromptTemplate resolves a UUID or a fully qualified name. Read and
// Update both go through it so either identifier works.
func lookupPromptTemplate(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.PromptTemplate, error) {
	found, err := sdk.PromptTemplates.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.PromptTemplates.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners, and reviewers managed outside
// Terraform are left alone.
func patchOps(plan, state promptTemplateModel) []ometa.JSONPatchOp {
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

	switch {
	case plan.SystemPrompt.Equal(state.SystemPrompt):
	case plan.SystemPrompt.ValueString() == "":
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/systemPrompt"})
	default:
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/systemPrompt",
			Value: plan.SystemPrompt.ValueString(),
		})
	}

	if !plan.TemplateContent.Equal(state.TemplateContent) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/templateContent",
			Value: plan.TemplateContent.ValueString(),
		})
	}

	switch {
	case plan.TemplateType.Equal(state.TemplateType):
	case plan.TemplateType.IsUnknown():
	case plan.TemplateType.IsNull(), plan.TemplateType.ValueString() == "":
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/templateType"})
	default:
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/templateType",
			Value: plan.TemplateType.ValueString(),
		})
	}

	switch {
	case plan.TemplateVersion.Equal(state.TemplateVersion):
	case plan.TemplateVersion.IsUnknown():
	case plan.TemplateVersion.IsNull(), plan.TemplateVersion.ValueString() == "":
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/templateVersion"})
	default:
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/templateVersion",
			Value: plan.TemplateVersion.ValueString(),
		})
	}

	return ops
}
