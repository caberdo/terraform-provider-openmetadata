package dashboard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	_ resource.Resource                = (*dashboardResource)(nil)
	_ resource.ResourceWithConfigure   = (*dashboardResource)(nil)
	_ resource.ResourceWithImportState = (*dashboardResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &dashboardResource{}
}

type dashboardResource struct {
	client *client.Client
}

func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata dashboard. A dashboard belongs to a dashboard service " +
			"and can reference the charts and data models it renders.",
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
				Description: "Dashboard name. Changing it replaces the dashboard, because the name is the suffix of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name: the dashboard service name followed by the dashboard name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the dashboard.",
				Default:     stringdefault.StaticString(""),
			},
			"service": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the dashboard service that owns the dashboard. The service is fixed at creation, so changing it replaces the dashboard.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"dashboard_type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Kind of dashboard. One of Dashboard or Report.",
				Default:     stringdefault.StaticString("Dashboard"),
				Validators: []validator.String{
					stringvalidator.OneOf("Dashboard", "Report"),
				},
			},
			"source_url": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL of the dashboard in the source system.",
				Default:     stringdefault.StaticString(""),
			},
			"charts": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Chart fully qualified names.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"data_models": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Dashboard data model fully qualified names.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
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

func (r *dashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateDashboard{
		Name:        plan.Name.ValueString(),
		Service:     plan.Service.ValueString(),
		Description: convert.StringPointer(plan.Description),
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	body.SourceUrl = convert.StringPointer(plan.SourceURL)

	if !plan.DashboardType.IsNull() && !plan.DashboardType.IsUnknown() {
		dashboardType := ometa.CreateDashboardDashboardType(plan.DashboardType.ValueString())
		body.DashboardType = &dashboardType
	}

	if !plan.Charts.IsNull() && !plan.Charts.IsUnknown() {
		charts, diags := stringSlice(ctx, plan.Charts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.Charts = &charts
	}

	if !plan.DataModels.IsNull() && !plan.DataModels.IsUnknown() {
		dataModels, diags := stringSlice(ctx, plan.DataModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.DataModels = &dataModels
	}

	created, err := sdk.Dashboards.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata dashboard",
			fmt.Sprintf("Could not create dashboard %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardModel
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

	found, err := lookupDashboard(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata dashboard",
			fmt.Sprintf("Could not read dashboard %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dashboardModel
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

	var charts, dataModels []ometa.EntityReference
	if !plan.Charts.Equal(state.Charts) {
		values, diags := chartReferences(ctx, sdk, plan.Charts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		charts = values
	}

	if !plan.DataModels.Equal(state.DataModels) {
		values, diags := dataModelReferences(ctx, sdk, plan.DataModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		dataModels = values
	}

	ops := patchOps(plan, state, charts, dataModels)
	if len(ops) > 0 {
		updated, err := sdk.Dashboards.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata dashboard",
				fmt.Sprintf("Could not update dashboard %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupDashboard(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata dashboard after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardModel
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
	err = sdk.Dashboards.Delete(ctx, state.ID.ValueString(), &ometa.DeleteDashboardParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata dashboard",
			fmt.Sprintf("Could not delete dashboard %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the dashboard UUID or its fully qualified name.
func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupDashboard resolves a UUID or a fully qualified name. Charts and data
// models are requested explicitly because the default read omits them.
func lookupDashboard(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Dashboard, error) {
	fields := "charts,dataModels"
	found, err := sdk.Dashboards.GetByID(ctx, reference, &ometa.GetDashboardByIDParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Dashboards.GetByName(ctx, reference, &ometa.GetDashboardByFQNParams{Fields: &fields})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// The service and name are not patchable because the attributes force
// replacement, so other people's tags, owners and domains stay untouched.
func patchOps(plan, state dashboardModel, charts, dataModels []ometa.EntityReference) []ometa.JSONPatchOp {
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

	if !plan.DashboardType.Equal(state.DashboardType) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/dashboardType",
			Value: plan.DashboardType.ValueString(),
		})
	}

	if !plan.SourceURL.Equal(state.SourceURL) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/sourceUrl",
			Value: plan.SourceURL.ValueString(),
		})
	}

	if !plan.Charts.Equal(state.Charts) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/charts",
			Value: charts,
		})
	}

	if !plan.DataModels.Equal(state.DataModels) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/dataModels",
			Value: dataModels,
		})
	}

	return ops
}

// chartReferences turns the chart fully qualified names in the plan into the
// entity references the dashboard update endpoint expects. Create accepts
// plain names; patch does not.
func chartReferences(ctx context.Context, sdk *ometa.Client, values types.List) ([]ometa.EntityReference, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, nameDiags := convert.StringList(ctx, values)
	diags.Append(nameDiags...)
	if diags.HasError() {
		return nil, diags
	}

	references := make([]ometa.EntityReference, 0, len(names))
	for _, name := range names {
		chart, err := sdk.Charts.GetByName(ctx, name, nil)
		if err != nil {
			diags.AddError(
				"Failed to resolve OpenMetadata chart",
				fmt.Sprintf("Could not resolve chart %q: %s", name, client.ErrorDetail(err)),
			)
			continue
		}

		reference := ometa.EntityReference{
			Id:                 chart.Id,
			Type:               "chart",
			FullyQualifiedName: chart.FullyQualifiedName,
		}
		reference.Name = &chart.Name
		references = append(references, reference)
	}

	return references, diags
}

// dataModelReferences resolves dashboard data model names for a patch.
func dataModelReferences(ctx context.Context, sdk *ometa.Client, values types.List) ([]ometa.EntityReference, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, nameDiags := convert.StringList(ctx, values)
	diags.Append(nameDiags...)
	if diags.HasError() {
		return nil, diags
	}

	references := make([]ometa.EntityReference, 0, len(names))
	for _, name := range names {
		dataModel, err := sdk.DashboardDataModels.GetByName(ctx, name, nil)
		if err != nil {
			diags.AddError(
				"Failed to resolve OpenMetadata dashboard data model",
				fmt.Sprintf("Could not resolve dashboard data model %q: %s", name, client.ErrorDetail(err)),
			)
			continue
		}

		reference := ometa.EntityReference{
			Id:                 dataModel.Id,
			Type:               "dashboardDataModel",
			FullyQualifiedName: dataModel.FullyQualifiedName,
		}
		reference.Name = &dataModel.Name
		references = append(references, reference)
	}

	return references, diags
}

// stringSlice converts a Terraform list to the string slice the API accepts.
// A removal must send an empty list rather than null, so nil becomes empty.
func stringSlice(ctx context.Context, value types.List) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return []string{}, nil
	}
	values, diags := convert.StringList(ctx, value)
	if diags.HasError() {
		return nil, diags
	}
	if values == nil {
		values = []string{}
	}
	return values, diags
}
