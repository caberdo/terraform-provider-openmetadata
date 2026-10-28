package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithConfigure   = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

const defaultJWTTokenExpiry = "7"

// setStrings maps a Terraform set to a slice. Null and unknown sets yield nil.
func setStrings(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var out []string
	diags := value.ElementsAs(ctx, &out, false)
	return out, diags
}

var jwtTokenExpiries = []string{"1", "7", "30", "60", "90", "OneHour", "Unlimited"}

// referenceResolver resolves an entity fully qualified name to its UUID.
type referenceResolver func(ctx context.Context, kind, name string) (uuid.UUID, error)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *client.Client
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata user. Name, email, bot status and JWT expiry are fixed at creation " +
			"and replace the user when changed.",
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
				Description: "User name and fully qualified name. It is immutable, so changing it replaces the user. Under basic authentication the server rewrites the name to the email local part, so configure both to agree.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a user this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the user.",
				Default:     stringdefault.StaticString(""),
			},
			"email": rschema.StringAttribute{
				Required:    true,
				Description: "Email address. The server restores the original value on update, so changing it replaces the user. Under basic authentication the email local part must equal name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_admin": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the user is an administrator. The server defaults this to false.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_bot": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the user backs an OpenMetadata bot. The server fixes this at creation, so changing it replaces the user.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"jwt_token_expiry": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JWT token expiry for a bot user, one of 1, 7, 30, 60, 90, OneHour or Unlimited. It is ignored for normal users and defaults to 7 for bots. The server fixes it at creation, so changing it replaces the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(jwtTokenExpiries...),
				},
			},
			"teams": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of the teams the user belongs to. The server returns teams sorted, so this is a set. Creating a user without teams assigns the Organization team; configuring teams replaces that assignment.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"roles": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of the roles assigned to the user. The server returns roles sorted, so this is a set.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"default_persona": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Fully qualified name of the persona shown by default for the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IANA timezone of the user, for example Europe/Amsterdam.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	resolve := resolverFor(sdk)

	body := &ometa.CreateUser{
		Name:        plan.Name.ValueString(),
		Email:       plan.Email.ValueString(),
		Description: convert.StringPointer(plan.Description),
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	body.IsAdmin = convert.BoolPointer(plan.IsAdmin)
	body.IsBot = convert.BoolPointer(plan.IsBot)
	body.Timezone = convert.StringPointer(plan.Timezone)

	teams, diags := resolveReferences(ctx, "team", plan.Teams, resolve)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if teams != nil {
		body.Teams = &teams
	}

	roles, diags := resolveReferences(ctx, "role", plan.Roles, resolve)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if roles != nil {
		body.Roles = &roles
	}

	if !plan.DefaultPersona.IsNull() && !plan.DefaultPersona.IsUnknown() {
		defaultPersona, err := resolve(ctx, "persona", plan.DefaultPersona.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve default persona",
				fmt.Sprintf("Could not resolve persona %q: %s", plan.DefaultPersona.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		body.DefaultPersona = &ometa.EntityReference{Id: defaultPersona, Type: "persona"}
	}

	isBot := !plan.IsBot.IsNull() && !plan.IsBot.IsUnknown() && plan.IsBot.ValueBool()
	if isBot {
		expiry := defaultJWTTokenExpiry
		if !plan.JWTTokenExpiry.IsNull() && !plan.JWTTokenExpiry.IsUnknown() && plan.JWTTokenExpiry.ValueString() != "" {
			expiry = plan.JWTTokenExpiry.ValueString()
		}
		plan.JWTTokenExpiry = types.StringValue(expiry)

		authType := ometa.JWT
		body.AuthenticationMechanism = &ometa.AuthenticationMechanism{
			AuthType: &authType,
			Config:   map[string]any{"JWTTokenExpiry": expiry},
		}
	} else if plan.JWTTokenExpiry.IsUnknown() {
		plan.JWTTokenExpiry = types.StringNull()
	}

	created, err := sdk.Users.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata user",
			fmt.Sprintf("Could not create user %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	// The create response omits the relationship lists, so re-read with fields
	// before writing state.
	found, err := lookupUser(ctx, sdk, created.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata user after create",
			fmt.Sprintf("Could not read user %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel
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

	found, err := lookupUser(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata user",
			fmt.Sprintf("Could not read user %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userModel
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

	ops, diags := patchOps(ctx, plan, state, resolverFor(sdk))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(ops) > 0 {
		_, err := sdk.Users.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata user",
				fmt.Sprintf("Could not update user %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		found, err := lookupUser(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata user after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	} else {
		found, err := lookupUser(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata user after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel
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
	err = sdk.Users.Delete(ctx, state.ID.ValueString(), &ometa.DeleteUserParams{HardDelete: &hardDelete})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata user",
			fmt.Sprintf("Could not delete user %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the user UUID or its fully qualified name.
func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// userFields requests the relationship lists that the user endpoints omit by
// default.
const userFields = "teams,roles,defaultPersona"

// lookupUser resolves a UUID or a fully qualified name. Import and Read both
// go through it so either identifier works, and the relationship lists are
// always requested so state is not zeroed.
func lookupUser(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.User, error) {
	fields := userFields
	found, err := sdk.Users.GetByID(ctx, reference, &ometa.GetUserByIDParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Users.GetByName(ctx, reference, &ometa.GetUserByFQNParams{Fields: &fields})
}

// resolverFor turns fully qualified names into UUIDs through the SDK.
func resolverFor(sdk *ometa.Client) referenceResolver {
	return func(ctx context.Context, kind, name string) (uuid.UUID, error) {
		switch kind {
		case "team":
			team, err := sdk.Teams.GetByName(ctx, name, nil)
			if err != nil {
				return uuid.Nil, err
			}
			return team.Id, nil
		case "role":
			role, err := sdk.Roles.GetByName(ctx, name, nil)
			if err != nil {
				return uuid.Nil, err
			}
			return role.Id, nil
		case "persona":
			persona, err := sdk.Personas.GetByName(ctx, name, nil)
			if err != nil {
				return uuid.Nil, err
			}
			return persona.Id, nil
		default:
			return uuid.Nil, fmt.Errorf("unsupported entity type %q", kind)
		}
	}
}

// resolveReferences maps a planned list of fully qualified names to UUIDs.
// Null and unknown lists yield nil so the field is omitted from the request.
func resolveReferences(ctx context.Context, kind string, value types.Set, resolve referenceResolver) ([]uuid.UUID, diag.Diagnostics) {
	names, diags := setStrings(ctx, value)
	if diags.HasError() || len(names) == 0 {
		return nil, diags
	}

	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		id, err := resolve(ctx, kind, name)
		if err != nil {
			diags.AddError(
				"Failed to resolve "+kind,
				fmt.Sprintf("Could not resolve %s %q: %s", kind, name, client.ErrorDetail(err)),
			)
			return nil, diags
		}
		ids = append(ids, id)
	}
	return ids, diags
}

// referenceList maps names to the reference list shape used by JSON Patch.
func referenceList(ctx context.Context, kind string, value types.Set, resolve referenceResolver, diags *diag.Diagnostics) []ometa.EntityReference {
	names, listDiags := setStrings(ctx, value)
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil
	}

	references := make([]ometa.EntityReference, 0, len(names))
	for _, name := range names {
		id, err := resolve(ctx, kind, name)
		if err != nil {
			diags.AddError(
				"Failed to resolve "+kind,
				fmt.Sprintf("Could not resolve %s %q: %s", kind, name, client.ErrorDetail(err)),
			)
			return nil
		}
		references = append(references, ometa.EntityReference{Id: id, Type: kind})
	}
	return references
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Name, email, is_bot and jwt_token_expiry carry RequiresReplace instead: the
// server rewrites or fixes them at creation, so patching them is pointless.
func patchOps(ctx context.Context, plan, state userModel, resolve referenceResolver) ([]ometa.JSONPatchOp, diag.Diagnostics) {
	var ops []ometa.JSONPatchOp
	var diags diag.Diagnostics

	switch {
	case plan.DisplayName.Equal(state.DisplayName), plan.DisplayName.IsUnknown():
	case plan.DisplayName.IsNull():
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/displayName"})
	default:
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/displayName", Value: plan.DisplayName.ValueString()})
	}

	if !plan.Description.Equal(state.Description) && !plan.Description.IsUnknown() {
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/description", Value: plan.Description.ValueString()})
	}

	if !plan.IsAdmin.Equal(state.IsAdmin) && !plan.IsAdmin.IsUnknown() {
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/isAdmin", Value: plan.IsAdmin.ValueBool()})
	}

	switch {
	case plan.Timezone.Equal(state.Timezone), plan.Timezone.IsUnknown():
	case plan.Timezone.IsNull():
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/timezone"})
	default:
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/timezone", Value: plan.Timezone.ValueString()})
	}

	if !plan.Teams.Equal(state.Teams) && !plan.Teams.IsUnknown() {
		references := referenceList(ctx, "team", plan.Teams, resolve, &diags)
		if diags.HasError() {
			return nil, diags
		}
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/teams", Value: references})
	}

	if !plan.Roles.Equal(state.Roles) && !plan.Roles.IsUnknown() {
		references := referenceList(ctx, "role", plan.Roles, resolve, &diags)
		if diags.HasError() {
			return nil, diags
		}
		ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/roles", Value: references})
	}

	switch {
	case plan.DefaultPersona.Equal(state.DefaultPersona), plan.DefaultPersona.IsUnknown():
	case plan.DefaultPersona.IsNull():
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/defaultPersona"})
	default:
		id, err := resolve(ctx, "persona", plan.DefaultPersona.ValueString())
		if err != nil {
			diags.AddError(
				"Failed to resolve default persona",
				fmt.Sprintf("Could not resolve persona %q: %s", plan.DefaultPersona.ValueString(), client.ErrorDetail(err)),
			)
			return nil, diags
		}
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/defaultPersona",
			Value: ometa.EntityReference{Id: id, Type: "persona"},
		})
	}

	return ops, diags
}
