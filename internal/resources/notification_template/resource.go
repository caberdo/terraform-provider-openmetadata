package notification_template

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

var (
	_ resource.Resource                = (*notificationTemplateResource)(nil)
	_ resource.ResourceWithConfigure   = (*notificationTemplateResource)(nil)
	_ resource.ResourceWithImportState = (*notificationTemplateResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &notificationTemplateResource{}
}

type notificationTemplateResource struct {
	client *client.Client
}

func (r *notificationTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_template"
}

func (r *notificationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata notification template. Templates hold the Jinja subject and body " +
			"rendered when an alert or event subscription fires.",
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
				Description: "Template name. Changing it replaces the template, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a notification template this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the template.",
				Default:     stringdefault.StaticString(""),
			},
			"template_subject": rschema.StringAttribute{
				Required:    true,
				Description: "Jinja template for the notification subject.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"template_body": rschema.StringAttribute{
				Required:    true,
				Description: "Jinja template for the notification body.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"is_modified_from_default": rschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the template differs from the OpenMetadata built-in default.",
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

func (r *notificationTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateNotificationTemplate{
		Name:            plan.Name.ValueString(),
		TemplateSubject: plan.TemplateSubject.ValueString(),
		TemplateBody:    plan.TemplateBody.ValueString(),
	}
	body.Description = convert.StringPointer(plan.Description)
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.NotificationTemplates.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata notification template",
			fmt.Sprintf("Could not create notification template %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationTemplateModel
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

	found, err := lookupNotificationTemplate(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata notification template",
			fmt.Sprintf("Could not read notification template %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state notificationTemplateModel
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
		updated, err := sdk.NotificationTemplates.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata notification template",
				fmt.Sprintf("Could not update notification template %q: %s", state.Name.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupNotificationTemplate(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata notification template after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationTemplateModel
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
	err = sdk.NotificationTemplates.Delete(ctx, state.ID.ValueString(), &ometa.DeleteNotificationTemplateParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata notification template",
			fmt.Sprintf("Could not delete notification template %q: %s", state.Name.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the notification template UUID or its fully
// qualified name.
func (r *notificationTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupNotificationTemplate resolves a UUID or a fully qualified name. Import
// and Read both go through it so either identifier works.
func lookupNotificationTemplate(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.NotificationTemplate, error) {
	found, err := sdk.NotificationTemplates.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.NotificationTemplates.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners and tags managed outside Terraform
// are left alone.
func patchOps(plan, state notificationTemplateModel) []ometa.JSONPatchOp {
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

	if !plan.TemplateSubject.Equal(state.TemplateSubject) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/templateSubject",
			Value: plan.TemplateSubject.ValueString(),
		})
	}

	if !plan.TemplateBody.Equal(state.TemplateBody) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/templateBody",
			Value: plan.TemplateBody.ValueString(),
		})
	}

	// name is not patchable: it is the fully qualified name, so the attribute
	// carries RequiresReplace instead.

	return ops
}
