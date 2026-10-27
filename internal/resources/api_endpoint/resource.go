package api_endpoint

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

// apiRequestMethods are the HTTP methods OpenMetadata accepts for an endpoint.
var apiRequestMethods = []string{
	string(ometa.CreateAPIEndpointRequestMethodGET),
	string(ometa.CreateAPIEndpointRequestMethodPOST),
	string(ometa.CreateAPIEndpointRequestMethodPUT),
	string(ometa.CreateAPIEndpointRequestMethodPATCH),
	string(ometa.CreateAPIEndpointRequestMethodDELETE),
	string(ometa.CreateAPIEndpointRequestMethodHEAD),
	string(ometa.CreateAPIEndpointRequestMethodCONNECT),
	string(ometa.CreateAPIEndpointRequestMethodOPTIONS),
	string(ometa.CreateAPIEndpointRequestMethodTRACE),
}

var (
	_ resource.Resource                = (*apiEndpointResource)(nil)
	_ resource.ResourceWithConfigure   = (*apiEndpointResource)(nil)
	_ resource.ResourceWithImportState = (*apiEndpointResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &apiEndpointResource{}
}

type apiEndpointResource struct {
	client *client.Client
}

func (r *apiEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_endpoint"
}

func (r *apiEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata API endpoint. An endpoint is one operation of an API " +
			"collection and inherits the collection's service.",
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
				Description: "Endpoint name. Changing it replaces the endpoint, because the name is part of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"api_collection": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the parent API collection. Changing it replaces the endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the API service, derived by the server from the collection.",
			},
			"service_type": rschema.StringAttribute{
				Computed:    true,
				Description: "OpenMetadata API service type, derived by the server from the collection.",
			},
			"endpoint_url": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL of the endpoint relative to the collection root.",
				Default:     stringdefault.StaticString(""),
			},
			"request_method": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "HTTP method of the endpoint. Defaults to GET.",
				Default:     stringdefault.StaticString(string(ometa.CreateAPIEndpointRequestMethodGET)),
				Validators: []validator.String{
					stringvalidator.OneOf(apiRequestMethods...),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name, in the form service.collection.name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the endpoint.",
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

func (r *apiEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiEndpointModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	// The service is derived by the server from the collection, so it is
	// deliberately left out of the create body.
	requestMethod := ometa.CreateAPIEndpointRequestMethod(plan.RequestMethod.ValueString())
	body := &ometa.CreateAPIEndpoint{
		Name:          plan.Name.ValueString(),
		ApiCollection: plan.APICollection.ValueString(),
		Description:   convert.StringPointer(plan.Description),
		RequestMethod: &requestMethod,
	}
	if plan.EndpointURL.ValueString() != "" {
		body.EndpointURL = convert.StringPointer(plan.EndpointURL)
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.APIEndpoints.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata API endpoint",
			fmt.Sprintf("Could not create API endpoint %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiEndpointModel
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

	found, err := lookupAPIEndpoint(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata API endpoint",
			fmt.Sprintf("Could not read API endpoint %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state apiEndpointModel
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
		updated, err := sdk.APIEndpoints.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata API endpoint",
				fmt.Sprintf("Could not update API endpoint %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupAPIEndpoint(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata API endpoint after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiEndpointModel
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
	err = sdk.APIEndpoints.Delete(ctx, state.ID.ValueString(), &ometa.DeleteAPIEndpointParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata API endpoint",
			fmt.Sprintf("Could not delete API endpoint %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the API endpoint UUID or its fully qualified name.
func (r *apiEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupAPIEndpoint resolves a UUID or a fully qualified name. The collection
// is a relationship rather than a stored field, so it must be requested
// explicitly; without it an import would drop the collection from state.
func lookupAPIEndpoint(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.APIEndpoint, error) {
	fields := "apiCollection"
	found, err := sdk.APIEndpoints.GetByID(ctx, reference, &ometa.GetEndpointByIdParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.APIEndpoints.GetByName(ctx, reference, &ometa.GetEndpointByFQNParams{Fields: &fields})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// name and api_collection are fixed at creation and carry RequiresReplace
// instead.
func patchOps(plan, state apiEndpointModel) []ometa.JSONPatchOp {
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

	if !plan.EndpointURL.Equal(state.EndpointURL) {
		if plan.EndpointURL.ValueString() == "" {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/endpointURL"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/endpointURL",
				Value: plan.EndpointURL.ValueString(),
			})
		}
	}

	if !plan.RequestMethod.Equal(state.RequestMethod) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/requestMethod",
			Value: plan.RequestMethod.ValueString(),
		})
	}

	return ops
}
