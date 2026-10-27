package api_service

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

// apiServiceTypes are the connector types the OpenMetadata API service
// supports.
var apiServiceTypes = []string{
	string(ometa.CreateApiServiceServiceTypeRest),
	string(ometa.CreateApiServiceServiceTypeWEBHOOK),
}

var (
	_ resource.Resource                = (*apiServiceResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiServiceResource)(nil)
	_ resource.ResourceWithImportState = (*apiServiceResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &apiServiceResource{}
}

type apiServiceResource struct {
	client *client.Client
}

func (r *apiServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_service"
}

func (r *apiServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata API service. An API service registers a connection " +
			"to a REST or webhook source so its API collections and endpoints can be ingested.",
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
				Description: "API service name. Changing it replaces the service, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"service_type": rschema.StringAttribute{
				Required:    true,
				Description: "OpenMetadata API service type: Rest or WEBHOOK. The server fixes it at creation, so changing it replaces the service.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(apiServiceTypes...),
				},
			},
			"connection_config": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Connector configuration as a JSON object string. Its shape depends on service_type and it may contain credentials, so it is marked sensitive.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					semanticJSONPlanModifier(),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For an API service this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the API service.",
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

func (r *apiServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateApiService{
		Name:        plan.Name.ValueString(),
		ServiceType: ometa.CreateApiServiceServiceType(plan.ServiceType.ValueString()),
		Description: convert.StringPointer(plan.Description),
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	if connectionConfigured(plan.Connection) {
		decoded, err := parseConnectionJSON(plan.Connection.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("connection_config"),
				"Invalid connection JSON",
				fmt.Sprintf("The connection attribute must be a JSON object: %s", err),
			)
			return
		}
		body.Connection = &ometa.ApiConnection{Config: decoded}
	}

	created, err := sdk.ApiServices.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata API service",
			fmt.Sprintf("Could not create API service %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiServiceModel
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

	found, err := lookupAPIService(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata API service",
			fmt.Sprintf("Could not read API service %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state apiServiceModel
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
		updated, err := sdk.ApiServices.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata API service",
				fmt.Sprintf("Could not update API service %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupAPIService(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata API service after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiServiceModel
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
	err = sdk.ApiServices.Delete(ctx, state.ID.ValueString(), &ometa.DeleteAPIServiceParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata API service",
			fmt.Sprintf("Could not delete API service %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the API service UUID or its fully qualified name.
func (r *apiServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupAPIService resolves a UUID or a fully qualified name. The connection
// field is requested explicitly, because the API omits it by default.
func lookupAPIService(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.ApiService, error) {
	fields := "connection"
	found, err := sdk.ApiServices.GetByID(ctx, reference, &ometa.GetAPIServiceByIDParams{
		Fields: &fields,
	})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.ApiServices.GetByName(ctx, reference, &ometa.GetAPIServiceByFQNParams{
		Fields: &fields,
	})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// name and service_type are fixed at creation and carry RequiresReplace
// instead.
func patchOps(plan, state apiServiceModel) []ometa.JSONPatchOp {
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

	if !plan.Connection.IsUnknown() && !state.Connection.IsUnknown() {
		switch {
		case plan.Connection.IsNull() && state.Connection.IsNull():
		case plan.Connection.IsNull():
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/connection"})
		case state.Connection.IsNull() || !jsonEqual(plan.Connection.ValueString(), state.Connection.ValueString()):
			if decoded, err := parseConnectionJSON(plan.Connection.ValueString()); err == nil {
				ops = append(ops, ometa.JSONPatchOp{
					Op:   "add",
					Path: "/connection",
					Value: &ometa.ApiConnection{
						Config: decoded,
					},
				})
			}
		}
	}

	return ops
}

// semanticJSONPlanModifier keeps the prior connection value when the planned
// document is semantically equal to state, so a reformatted but equivalent
// JSON attribute does not produce a perpetual diff.
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
