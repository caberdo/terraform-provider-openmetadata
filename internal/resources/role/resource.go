package role

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*roleResource)(nil)
	_ resource.ResourceWithConfigure   = (*roleResource)(nil)
	_ resource.ResourceWithImportState = (*roleResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *client.Client
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata role. A role groups policies that grant access to " +
			"metadata assets and is assigned to users and teams.",
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
				Description: "Role name. Changing it replaces the role, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a role this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the role.",
				Default:     stringdefault.StaticString(""),
			},
			"policies": rschema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Fully qualified names of the policies granted to the role. The server sorts the list, so it is modelled as a set. At least one is required.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
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

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var policies []string
	resp.Diagnostics.Append(plan.Policies.ElementsAs(ctx, &policies, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreateRole{
		Name:     plan.Name.ValueString(),
		Policies: policies,
	}
	body.Description = convert.StringPointer(plan.Description)
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.Roles.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata role",
			fmt.Sprintf("Could not create role %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleModel
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

	found, err := lookupRole(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata role",
			fmt.Sprintf("Could not read role %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleModel
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

	var policyRefs []ometa.EntityReference
	if !plan.Policies.Equal(state.Policies) {
		var diags diag.Diagnostics
		policyRefs, diags = policyReferences(ctx, sdk, plan.Policies)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	ops := patchOps(plan, state, policyRefs)
	if len(ops) > 0 {
		if _, err := sdk.Roles.Patch(ctx, state.ID.ValueString(), ops); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata role",
				fmt.Sprintf("Could not update role %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
	}

	// Re-read after the patch so policies are always present: the patch
	// response does not include the relationship unless it was requested.
	found, err := lookupRole(ctx, sdk, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read OpenMetadata role after update", client.ErrorDetail(err))
		return
	}
	apply(found, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleModel
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
	err = sdk.Roles.Delete(ctx, state.ID.ValueString(), &ometa.DeleteRoleParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata role",
			fmt.Sprintf("Could not delete role %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the role UUID or its fully qualified name.
func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupRole resolves a UUID or a fully qualified name. Policies are requested
// explicitly because the default read omits them.
func lookupRole(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Role, error) {
	fields := "policies"
	found, err := sdk.Roles.GetByID(ctx, reference, &ometa.GetRoleByIDParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Roles.GetByName(ctx, reference, &ometa.GetRoleByFQNParams{Fields: &fields})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// The policies value must be entity references: unlike create, the update
// endpoint does not accept fully qualified name strings. Name is fixed at
// creation and never patched.
func patchOps(plan, state roleModel, policies []ometa.EntityReference) []ometa.JSONPatchOp {
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

	if !plan.Policies.Equal(state.Policies) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/policies",
			Value: policies,
		})
	}

	return ops
}

// policyReferences turns the policy names in the plan into the entity
// references the role update endpoint expects.
func policyReferences(ctx context.Context, sdk *ometa.Client, values types.Set) ([]ometa.EntityReference, diag.Diagnostics) {
	var diags diag.Diagnostics

	var names []string
	diags.Append(values.ElementsAs(ctx, &names, false)...)
	if diags.HasError() {
		return nil, diags
	}

	references := make([]ometa.EntityReference, 0, len(names))
	for _, name := range names {
		policy, err := sdk.Policies.GetByName(ctx, name, nil)
		if err != nil {
			diags.AddError(
				"Failed to resolve OpenMetadata policy",
				fmt.Sprintf("Could not resolve policy %q: %s", name, client.ErrorDetail(err)),
			)
			continue
		}

		reference := ometa.EntityReference{
			Id:                 policy.Id,
			Type:               "policy",
			FullyQualifiedName: policy.FullyQualifiedName,
		}
		reference.Name = &policy.Name
		references = append(references, reference)
	}

	return references, diags
}
