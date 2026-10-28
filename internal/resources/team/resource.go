package team

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
	_ resource.Resource                = (*teamResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamResource)(nil)
	_ resource.ResourceWithImportState = (*teamResource)(nil)
)

// Reference types used by team relationships. They match the values the server
// expects on an EntityReference.
const (
	referenceTeam    = "team"
	referencePolicy  = "policy"
	referenceRole    = "role"
	referencePersona = "persona"
	referenceUser    = "user"
)

// teamFields lists the relationship fields the read endpoints omit unless
// requested. Without them Read would zero parents, policies, roles, persona and
// users after every refresh.
const teamFields = "parents,policies,defaultRoles,defaultPersona,users,teamType,email"

var teamTypes = []string{
	string(ometa.CreateTeamTeamTypeDepartment),
	string(ometa.CreateTeamTeamTypeDivision),
	string(ometa.CreateTeamTeamTypeBusinessUnit),
	string(ometa.CreateTeamTeamTypeGroup),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *client.Client
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata team. Teams are the organisational units of OpenMetadata: " +
			"they own assets, group users and carry policies and default roles. A team without parents " +
			"is placed under the Organization team by the server.",
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
				Description: "Team name. Changing it replaces the team, because the name is part of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a team this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the team.",
				Default:     stringdefault.StaticString(""),
			},
			"team_type": rschema.StringAttribute{
				Required: true,
				Description: "Kind of team: Department, Division, BusinessUnit or Group. The server validates " +
					"the hierarchy, so a Department may sit under a Division, BusinessUnit or Organization. " +
					"Patchable in place.",
				Validators: []validator.String{
					stringvalidator.OneOf(teamTypes...),
				},
			},
			"email": rschema.StringAttribute{
				Optional:    true,
				Description: "Contact email of the team.",
			},
			"is_joinable": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether users may join the team themselves. The server defaults it to true.",
			},
			"parents": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of the parent teams. An omitted set places the team under " +
					"Organization. When set, at least one parent is required.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"policies": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of policies assigned to the team.",
			},
			"default_roles": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of roles granted to the team's users by default.",
			},
			"default_persona": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Fully qualified name of the persona the team's users land on by default.",
			},
			"users": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the users that belong to the team.",
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

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	parents, diags := setIDs(ctx, sdk, referenceTeam, plan.Parents, path.Root("parents"))
	resp.Diagnostics.Append(diags...)
	policies, diags := setIDs(ctx, sdk, referencePolicy, plan.Policies, path.Root("policies"))
	resp.Diagnostics.Append(diags...)
	defaultRoles, diags := setIDs(ctx, sdk, referenceRole, plan.DefaultRoles, path.Root("default_roles"))
	resp.Diagnostics.Append(diags...)
	users, diags := setIDs(ctx, sdk, referenceUser, plan.Users, path.Root("users"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreateTeam{
		Name:        plan.Name.ValueString(),
		TeamType:    ometa.CreateTeamTeamType(plan.TeamType.ValueString()),
		Description: convert.StringPointer(plan.Description),
		DisplayName: convert.StringPointer(plan.DisplayName),
		Email:       convert.StringPointer(plan.Email),
		IsJoinable:  convert.BoolPointer(plan.IsJoinable),
	}
	if len(parents) > 0 {
		body.Parents = &parents
	}
	if len(policies) > 0 {
		body.Policies = &policies
	}
	if len(defaultRoles) > 0 {
		body.DefaultRoles = &defaultRoles
	}
	if len(users) > 0 {
		body.Users = &users
	}

	if !plan.DefaultPersona.IsNull() && !plan.DefaultPersona.IsUnknown() {
		persona, err := resolveID(ctx, sdk, referencePersona, plan.DefaultPersona.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("default_persona"),
				"Failed to resolve persona reference",
				err.Error(),
			)
			return
		}
		body.DefaultPersona = &persona
	}

	created, err := sdk.Teams.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata team",
			fmt.Sprintf("Could not create team %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	// Create returns the entity without its relationship fields, so read it
	// back through the field-expanding lookup before writing state.
	found, err := lookupTeam(ctx, sdk, created.Id.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata team after create",
			fmt.Sprintf("Created team %q but could not read it back: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamModel
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

	found, err := lookupTeam(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata team",
			fmt.Sprintf("Could not read team %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state teamModel
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

	lookup := func(ctx context.Context, kind string, names []string) ([]ometa.EntityReference, error) {
		return resolveReferences(ctx, sdk, kind, names)
	}

	ops, err := patchOps(ctx, plan, state, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve team references", err.Error())
		return
	}

	if len(ops) > 0 {
		if _, err := sdk.Teams.Patch(ctx, state.ID.ValueString(), ops); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata team",
				fmt.Sprintf("Could not update team %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
	}

	// Patch returns the entity without its relationship fields, so always read
	// the team back with the expanded field list before writing state.
	found, err := lookupTeam(ctx, sdk, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read OpenMetadata team after update", client.ErrorDetail(err))
		return
	}

	apply(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamModel
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
	err = sdk.Teams.Delete(ctx, state.ID.ValueString(), &ometa.DeleteTeamParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata team",
			fmt.Sprintf("Could not delete team %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the team UUID or its fully qualified name.
func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupTeam resolves a UUID or a fully qualified name. Import and Read both
// go through it so either identifier works.
func lookupTeam(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Team, error) {
	fields := teamFields
	found, err := sdk.Teams.GetByID(ctx, reference, &ometa.GetTeamByIDParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Teams.GetByName(ctx, reference, &ometa.GetTeamByFQNParams{Fields: &fields})
}

// setStrings renders a Terraform string set as a Go slice.
func setStrings(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	var out []string
	diags := value.ElementsAs(ctx, &out, false)
	return out, diags
}

// referenceLookup resolves fully qualified names to the entity references a
// JSON Patch operation carries. Update binds it to the SDK; tests substitute a
// fake so the patch document can be asserted without a server.
type referenceLookup func(ctx context.Context, kind string, names []string) ([]ometa.EntityReference, error)

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, tags and followers managed outside
// Terraform are left alone. name carries RequiresReplace and is never patched.
func patchOps(ctx context.Context, plan, state teamModel, lookup referenceLookup) ([]ometa.JSONPatchOp, error) {
	var ops []ometa.JSONPatchOp

	if !plan.TeamType.Equal(state.TeamType) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/teamType",
			Value: plan.TeamType.ValueString(),
		})
	}

	switch {
	case plan.Email.Equal(state.Email):
	case plan.Email.IsNull():
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/email"})
	default:
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/email",
			Value: plan.Email.ValueString(),
		})
	}

	if !plan.IsJoinable.IsUnknown() && !plan.IsJoinable.Equal(state.IsJoinable) && !plan.IsJoinable.IsNull() {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/isJoinable",
			Value: plan.IsJoinable.ValueBool(),
		})
	}

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

	for _, set := range []struct {
		attribute string
		kind      string
		plan      types.Set
		state     types.Set
	}{
		{"/parents", referenceTeam, plan.Parents, state.Parents},
		{"/policies", referencePolicy, plan.Policies, state.Policies},
		{"/defaultRoles", referenceRole, plan.DefaultRoles, state.DefaultRoles},
		{"/users", referenceUser, plan.Users, state.Users},
	} {
		op, ok, err := listPatchOp(ctx, set.attribute, set.kind, set.plan, set.state, lookup)
		if err != nil {
			return nil, err
		}
		if ok {
			ops = append(ops, op)
		}
	}

	switch {
	case plan.DefaultPersona.IsUnknown() || plan.DefaultPersona.Equal(state.DefaultPersona):
	case plan.DefaultPersona.IsNull():
		ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/defaultPersona"})
	default:
		references, err := lookup(ctx, referencePersona, []string{plan.DefaultPersona.ValueString()})
		if err != nil {
			return nil, err
		}
		if len(references) == 0 {
			return nil, fmt.Errorf("persona %q resolved to no reference", plan.DefaultPersona.ValueString())
		}
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/defaultPersona",
			Value: references[0],
		})
	}

	return ops, nil
}

// listPatchOp builds one replace-style operation for a reference set. Unknown
// plans and unchanged sets produce no operation.
func listPatchOp(
	ctx context.Context,
	attribute string,
	kind string,
	plan types.Set,
	state types.Set,
	lookup referenceLookup,
) (ometa.JSONPatchOp, bool, error) {
	if plan.IsUnknown() || plan.Equal(state) {
		return ometa.JSONPatchOp{}, false, nil
	}
	if plan.IsNull() {
		return ometa.JSONPatchOp{Op: "remove", Path: attribute}, true, nil
	}

	names, diags := setStrings(ctx, plan)
	if diags.HasError() {
		return ometa.JSONPatchOp{}, false, fmt.Errorf("%s: %v", attribute, diags.Errors())
	}
	references, err := lookup(ctx, kind, names)
	if err != nil {
		return ometa.JSONPatchOp{}, false, err
	}
	return ometa.JSONPatchOp{Op: "add", Path: attribute, Value: references}, true, nil
}

// setIDs resolves a Terraform set of fully qualified names to UUIDs. Null and
// unknown values resolve to nothing so the field is omitted from the request.
func setIDs(
	ctx context.Context,
	sdk *ometa.Client,
	kind string,
	value types.Set,
	attribute path.Path,
) ([]uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	names, listDiags := setStrings(ctx, value)
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil, diags
	}

	resolved, err := resolveIDs(ctx, sdk, kind, names)
	if err != nil {
		diags.AddAttributeError(attribute, "Failed to resolve "+kind+" reference", err.Error())
		return nil, diags
	}
	return resolved, diags
}

// resolveIDs resolves fully qualified names to entity UUIDs.
func resolveIDs(ctx context.Context, sdk *ometa.Client, kind string, names []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		id, err := resolveID(ctx, sdk, kind, name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// resolveID resolves one fully qualified name to the UUID of an entity.
func resolveID(ctx context.Context, sdk *ometa.Client, kind, name string) (uuid.UUID, error) {
	switch kind {
	case referenceTeam:
		found, err := sdk.Teams.GetByName(ctx, name, nil)
		if err != nil {
			return uuid.Nil, fmt.Errorf("team %q: %s", name, client.ErrorDetail(err))
		}
		return found.Id, nil
	case referencePolicy:
		found, err := sdk.Policies.GetByName(ctx, name, nil)
		if err != nil {
			return uuid.Nil, fmt.Errorf("policy %q: %s", name, client.ErrorDetail(err))
		}
		return found.Id, nil
	case referenceRole:
		found, err := sdk.Roles.GetByName(ctx, name, nil)
		if err != nil {
			return uuid.Nil, fmt.Errorf("role %q: %s", name, client.ErrorDetail(err))
		}
		return found.Id, nil
	case referencePersona:
		found, err := sdk.Personas.GetByName(ctx, name, nil)
		if err != nil {
			return uuid.Nil, fmt.Errorf("persona %q: %s", name, client.ErrorDetail(err))
		}
		return found.Id, nil
	case referenceUser:
		found, err := sdk.Users.GetByName(ctx, name, nil)
		if err != nil {
			return uuid.Nil, fmt.Errorf("user %q: %s", name, client.ErrorDetail(err))
		}
		return found.Id, nil
	default:
		return uuid.Nil, fmt.Errorf("unsupported reference type %q", kind)
	}
}

// resolveReferences resolves fully qualified names to the entity references a
// JSON Patch operation expects.
func resolveReferences(ctx context.Context, sdk *ometa.Client, kind string, names []string) ([]ometa.EntityReference, error) {
	ids, err := resolveIDs(ctx, sdk, kind, names)
	if err != nil {
		return nil, err
	}
	references := make([]ometa.EntityReference, len(ids))
	for index, id := range ids {
		references[index] = ometa.EntityReference{Id: id, Type: kind}
	}
	return references, nil
}
