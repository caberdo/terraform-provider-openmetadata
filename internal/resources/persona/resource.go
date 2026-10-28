package persona

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
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
	_ resource.Resource                = (*personaResource)(nil)
	_ resource.ResourceWithConfigure   = (*personaResource)(nil)
	_ resource.ResourceWithImportState = (*personaResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &personaResource{}
}

type personaResource struct {
	client *client.Client
}

func (r *personaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_persona"
}

func (r *personaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata persona. A persona represents a role a user assumes in " +
			"the organisation and can carry a customised UI and a set of assigned users.",
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
				Description: "Persona name. Changing it replaces the persona, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a persona this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the persona.",
				Default:     stringdefault.StaticString(""),
			},
			"default": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this is the default persona for the organisation. The server defaults it to false.",
			},
			"users": rschema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Fully qualified names of the users assigned to this persona. When set it replaces the whole set; when omitted the server assigns none.",
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

func (r *personaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *personaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan personaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	users, diags := userIDs(ctx, sdk, plan.Users)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreatePersona{
		Name:        plan.Name.ValueString(),
		Description: convert.StringPointer(plan.Description),
		Default:     convert.BoolPointer(plan.Default),
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	if len(users) > 0 {
		body.Users = &users
	}

	created, err := sdk.Personas.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata persona",
			fmt.Sprintf("Could not create persona %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	// The create response does not reliably include the assigned users, so read
	// the entity back with the fields the resource owns.
	found, err := lookupPersona(ctx, sdk, created.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata persona after create",
			fmt.Sprintf("Could not read persona %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *personaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state personaModel
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

	found, err := lookupPersona(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata persona",
			fmt.Sprintf("Could not read persona %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *personaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state personaModel
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

	var refs []ometa.EntityReference
	if usersChanged(plan, state) {
		var diags diag.Diagnostics
		refs, diags = userReferences(ctx, sdk, plan.Users)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	ops := patchOps(plan, state, refs)
	if len(ops) > 0 {
		if _, err := sdk.Personas.Patch(ctx, state.ID.ValueString(), ops); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata persona",
				fmt.Sprintf("Could not update persona %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
	}

	// Read back with the fields the resource owns: the patch response does not
	// reliably include the assigned users.
	found, err := lookupPersona(ctx, sdk, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read OpenMetadata persona after update", client.ErrorDetail(err))
		return
	}
	resp.Diagnostics.Append(apply(ctx, found, &plan)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *personaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state personaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK exposes no Delete for personas, but the REST API supports a
	// permanent delete at DELETE /personas/{id}.
	err := r.client.Delete(ctx, "personas/"+url.PathEscape(state.ID.ValueString()), url.Values{
		"hardDelete": {"true"},
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata persona",
			fmt.Sprintf("Could not delete persona %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the persona UUID or its fully qualified name.
func (r *personaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupPersona resolves a UUID or a fully qualified name. Assigned users are
// requested explicitly because the default read omits them.
func lookupPersona(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Persona, error) {
	fields := "users"
	found, err := sdk.Personas.GetByID(ctx, reference, &ometa.GetPersonaByIDParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Personas.GetByName(ctx, reference, &ometa.GetPersonaByFQNParams{Fields: &fields})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, experts and other fields managed
// outside Terraform are left alone. name is fixed at creation and carries
// RequiresReplace instead.
func patchOps(plan, state personaModel, users []ometa.EntityReference) []ometa.JSONPatchOp {
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

	if !plan.Default.IsNull() && !plan.Default.IsUnknown() && !plan.Default.Equal(state.Default) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/default",
			Value: plan.Default.ValueBool(),
		})
	}

	if usersChanged(plan, state) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/users",
			Value: users,
		})
	}

	return ops
}

// usersChanged reports whether the planned user list is known and differs from
// state. Unknown lists belong to attributes the configuration omits.
func usersChanged(plan, state personaModel) bool {
	return !plan.Users.IsNull() && !plan.Users.IsUnknown() && !plan.Users.Equal(state.Users)
}

// userNames renders a configured set of user names as a Go slice. Null and
// unknown sets yield nil so the field is omitted from a request.
func userNames(ctx context.Context, values types.Set) ([]string, diag.Diagnostics) {
	if values.IsNull() || values.IsUnknown() {
		return nil, nil
	}

	var names []string
	diags := values.ElementsAs(ctx, &names, false)
	return names, diags
}

// userIDs resolves user fully qualified names to the UUIDs persona creation
// expects. A null or unknown set yields nil so the field is omitted.
func userIDs(ctx context.Context, sdk *ometa.Client, values types.Set) ([]uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, nameDiags := userNames(ctx, values)
	diags.Append(nameDiags...)
	if diags.HasError() {
		return nil, diags
	}

	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		user, err := sdk.Users.GetByName(ctx, name, nil)
		if err != nil {
			diags.AddError(
				"Failed to resolve OpenMetadata user",
				fmt.Sprintf("Could not resolve user %q: %s", name, client.ErrorDetail(err)),
			)
			continue
		}
		ids = append(ids, user.Id)
	}

	return ids, diags
}

// userReferences turns the fully qualified names in the plan into the entity
// references the persona update endpoint expects.
func userReferences(ctx context.Context, sdk *ometa.Client, values types.Set) ([]ometa.EntityReference, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, nameDiags := userNames(ctx, values)
	diags.Append(nameDiags...)
	if diags.HasError() {
		return nil, diags
	}

	references := make([]ometa.EntityReference, 0, len(names))
	for _, name := range names {
		user, err := sdk.Users.GetByName(ctx, name, nil)
		if err != nil {
			diags.AddError(
				"Failed to resolve OpenMetadata user",
				fmt.Sprintf("Could not resolve user %q: %s", name, client.ErrorDetail(err)),
			)
			continue
		}

		reference := ometa.EntityReference{
			Id:                 user.Id,
			Type:               "user",
			FullyQualifiedName: user.FullyQualifiedName,
		}
		reference.Name = &user.Name
		references = append(references, reference)
	}

	return references, diags
}
