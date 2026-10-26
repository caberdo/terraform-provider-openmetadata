package chart

import (
	"context"
	"fmt"

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
	_ resource.Resource                = (*chartResource)(nil)
	_ resource.ResourceWithConfigure   = (*chartResource)(nil)
	_ resource.ResourceWithImportState = (*chartResource)(nil)
)

var chartTypes = []string{
	string(ometa.CreateChartChartTypeArea),
	string(ometa.CreateChartChartTypeBar),
	string(ometa.CreateChartChartTypeBoxPlot),
	string(ometa.CreateChartChartTypeGauge),
	string(ometa.CreateChartChartTypeGraph),
	string(ometa.CreateChartChartTypeHeatmap),
	string(ometa.CreateChartChartTypeHistogram),
	string(ometa.CreateChartChartTypeLine),
	string(ometa.CreateChartChartTypeMap),
	string(ometa.CreateChartChartTypeOther),
	string(ometa.CreateChartChartTypePie),
	string(ometa.CreateChartChartTypeSanKey),
	string(ometa.CreateChartChartTypeScatter),
	string(ometa.CreateChartChartTypeTable),
	string(ometa.CreateChartChartTypeText),
	string(ometa.CreateChartChartTypeTimeline),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &chartResource{}
}

type chartResource struct {
	client *client.Client
}

func (r *chartResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chart"
}

func (r *chartResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata chart. A chart belongs to a dashboard service and " +
			"can be attached to one or more dashboards.",
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
				Description: "Chart name. Changing it replaces the chart, because the name is also part of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. It combines the dashboard service name with the chart name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the chart.",
				Default:     stringdefault.StaticString(""),
			},
			"service": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the dashboard service that owns the chart. Changing it replaces the chart.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"chart_type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Chart visualization type, for example Bar, Line or Pie.",
				Default:     stringdefault.StaticString("Other"),
				Validators: []validator.String{
					stringvalidator.OneOf(chartTypes...),
				},
			},
			"source_url": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL of the chart in the source BI tool.",
				Default:     stringdefault.StaticString(""),
			},
			"dashboards": rschema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Fully qualified names of the dashboards the chart is placed on. An empty list detaches the chart from every dashboard.",
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

func (r *chartResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *chartResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan chartModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	chartType := ometa.CreateChartChartType(plan.ChartType.ValueString())
	body := &ometa.CreateChart{
		Name:        plan.Name.ValueString(),
		Service:     plan.Service.ValueString(),
		Description: convert.StringPointer(plan.Description),
		ChartType:   &chartType,
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	body.SourceUrl = convert.StringPointer(plan.SourceURL)

	if !plan.Dashboards.IsUnknown() && !plan.Dashboards.IsNull() {
		dashboards, diags := convert.StringList(ctx, plan.Dashboards)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		body.Dashboards = &dashboards
	}

	created, err := sdk.Charts.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata chart",
			fmt.Sprintf("Could not create chart %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *chartResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state chartModel
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

	found, err := lookupChart(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata chart",
			fmt.Sprintf("Could not read chart %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *chartResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state chartModel
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

	var dashboards []ometa.EntityReference
	if !plan.Dashboards.IsUnknown() && !plan.Dashboards.Equal(state.Dashboards) {
		references, diags := dashboardReferences(ctx, sdk, plan.Dashboards)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		dashboards = references
	}

	ops := patchOps(plan, state, dashboards)
	if len(ops) > 0 {
		updated, err := sdk.Charts.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata chart",
				fmt.Sprintf("Could not update chart %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupChart(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata chart after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *chartResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state chartModel
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
	err = sdk.Charts.Delete(ctx, state.ID.ValueString(), &ometa.DeleteChartParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata chart",
			fmt.Sprintf("Could not delete chart %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the chart UUID or its fully qualified name.
func (r *chartResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupChart resolves a UUID or a fully qualified name. The dashboards field
// is always requested so apply can populate the dashboard list.
func lookupChart(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Chart, error) {
	fields := "dashboards"

	found, err := sdk.Charts.GetByID(ctx, reference, &ometa.GetChartByIDParams{Fields: &fields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Charts.GetByName(ctx, reference, &ometa.GetChartByFQNParams{Fields: &fields})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, tags and other metadata managed
// outside Terraform are left alone.
func patchOps(plan, state chartModel, dashboards []ometa.EntityReference) []ometa.JSONPatchOp {
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

	if !plan.ChartType.Equal(state.ChartType) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/chartType",
			Value: plan.ChartType.ValueString(),
		})
	}

	if !plan.SourceURL.Equal(state.SourceURL) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/sourceUrl",
			Value: plan.SourceURL.ValueString(),
		})
	}

	if !plan.Dashboards.IsUnknown() && !plan.Dashboards.Equal(state.Dashboards) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/dashboards",
			Value: dashboards,
		})
	}

	// service and name are not patchable: the chart inherits its fully
	// qualified name from them. Both attributes carry RequiresReplace.

	return ops
}

// dashboardReferences resolves the dashboard fully qualified names in the plan
// to the entity references the update endpoint expects. Create accepts plain
// names; patch does not.
func dashboardReferences(ctx context.Context, sdk *ometa.Client, values types.List) ([]ometa.EntityReference, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, nameDiags := convert.StringList(ctx, values)
	diags.Append(nameDiags...)
	if diags.HasError() {
		return nil, diags
	}

	references := make([]ometa.EntityReference, 0, len(names))
	for _, name := range names {
		dashboard, err := sdk.Dashboards.GetByName(ctx, name, nil)
		if err != nil {
			diags.AddError(
				"Failed to resolve OpenMetadata dashboard",
				fmt.Sprintf("Could not resolve dashboard %q: %s", name, client.ErrorDetail(err)),
			)
			continue
		}

		reference := ometa.EntityReference{
			Id:                 dashboard.Id,
			Type:               "dashboard",
			FullyQualifiedName: dashboard.FullyQualifiedName,
		}
		reference.Name = &dashboard.Name
		references = append(references, reference)
	}

	return references, diags
}
