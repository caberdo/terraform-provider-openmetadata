package metric

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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*metricResource)(nil)
	_ resource.ResourceWithConfigure   = (*metricResource)(nil)
	_ resource.ResourceWithImportState = (*metricResource)(nil)
)

var (
	metricTypes = []string{
		string(ometa.MetricMetricTypeAVERAGE),
		string(ometa.MetricMetricTypeCOUNT),
		string(ometa.MetricMetricTypeMAX),
		string(ometa.MetricMetricTypeMEDIAN),
		string(ometa.MetricMetricTypeMIN),
		string(ometa.MetricMetricTypeMODE),
		string(ometa.MetricMetricTypeOTHER),
		string(ometa.MetricMetricTypePERCENTAGE),
		string(ometa.MetricMetricTypeRATIO),
		string(ometa.MetricMetricTypeSTANDARDDEVIATION),
		string(ometa.MetricMetricTypeSUM),
		string(ometa.MetricMetricTypeVARIANCE),
	}

	unitsOfMeasurement = []string{
		string(ometa.MetricUnitOfMeasurementCOUNT),
		string(ometa.MetricUnitOfMeasurementDOLLARS),
		string(ometa.MetricUnitOfMeasurementEVENTS),
		string(ometa.MetricUnitOfMeasurementOTHER),
		string(ometa.MetricUnitOfMeasurementPERCENTAGE),
		string(ometa.MetricUnitOfMeasurementREQUESTS),
		string(ometa.MetricUnitOfMeasurementSIZE),
		string(ometa.MetricUnitOfMeasurementTIMESTAMP),
		string(ometa.MetricUnitOfMeasurementTRANSACTIONS),
	}

	granularities = []string{
		string(ometa.MetricGranularityDAY),
		string(ometa.MetricGranularityHOUR),
		string(ometa.MetricGranularityMINUTE),
		string(ometa.MetricGranularityMONTH),
		string(ometa.MetricGranularityQUARTER),
		string(ometa.MetricGranularitySECOND),
		string(ometa.MetricGranularityWEEK),
		string(ometa.MetricGranularityYEAR),
	}

	expressionLanguages = []string{
		string(ometa.MetricExpressionLanguageExternal),
		string(ometa.MetricExpressionLanguageJava),
		string(ometa.MetricExpressionLanguageJavaScript),
		string(ometa.MetricExpressionLanguagePython),
		string(ometa.MetricExpressionLanguageSQL),
	}
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &metricResource{}
}

type metricResource struct {
	client *client.Client
}

func (r *metricResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metric"
}

func (r *metricResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata metric. A metric records a named calculation, its type, " +
			"its unit and the granularity it is computed at.",
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
				Description: "Metric name. Changing it replaces the metric, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a metric this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the metric.",
				Default:     stringdefault.StaticString(""),
			},
			"metric_type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Calculation performed by the metric, for example COUNT, SUM or AVERAGE.",
				Validators: []validator.String{
					stringvalidator.OneOf(metricTypes...),
				},
			},
			"unit_of_measurement": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Unit reported by the metric, for example COUNT, DOLLARS or PERCENTAGE.",
				Validators: []validator.String{
					stringvalidator.OneOf(unitsOfMeasurement...),
				},
			},
			"custom_unit_of_measurement": rschema.StringAttribute{
				Optional:    true,
				Description: "Free-form unit used when unit_of_measurement is OTHER.",
			},
			"granularity": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Time granularity the metric is computed at, for example DAY or HOUR.",
				Validators: []validator.String{
					stringvalidator.OneOf(granularities...),
				},
			},
			"expression_code": rschema.StringAttribute{
				Optional:    true,
				Description: "Code that computes the metric. Requires expression_language. Removing it replaces the metric.",
				PlanModifiers: []planmodifier.String{
					requiresReplaceOnRemoval(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("expression_language")),
				},
			},
			"expression_language": rschema.StringAttribute{
				Optional:    true,
				Description: "Language of expression_code: SQL, Python, Java, JavaScript or External. Removing it replaces the metric.",
				PlanModifiers: []planmodifier.String{
					requiresReplaceOnRemoval(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("expression_code")),
					stringvalidator.OneOf(expressionLanguages...),
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

func (r *metricResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *metricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateMetric{
		Name:        plan.Name.ValueString(),
		Description: convert.StringPointer(plan.Description),
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	body.MetricType = createMetricType(plan.MetricType)
	body.UnitOfMeasurement = createUnitOfMeasurement(plan.UnitOfMeasurement)
	body.CustomUnitOfMeasurement = convert.StringPointer(plan.CustomUnitOfMeasurement)
	body.Granularity = createGranularity(plan.Granularity)
	body.MetricExpression = createExpression(plan.ExpressionCode, plan.ExpressionLanguage)

	created, err := sdk.Metrics.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata metric",
			fmt.Sprintf("Could not create metric %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricModel
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

	found, err := lookupMetric(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata metric",
			fmt.Sprintf("Could not read metric %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state metricModel
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
		updated, err := sdk.Metrics.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata metric",
				fmt.Sprintf("Could not update metric %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupMetric(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata metric after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricModel
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
	err = sdk.Metrics.Delete(ctx, state.ID.ValueString(), &ometa.DeleteMetricParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata metric",
			fmt.Sprintf("Could not delete metric %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the metric UUID or its fully qualified name.
func (r *metricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupMetric resolves a UUID or a fully qualified name.
func lookupMetric(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Metric, error) {
	found, err := sdk.Metrics.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Metrics.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
func patchOps(plan, state metricModel) []ometa.JSONPatchOp {
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

	ops = appendStringPatch(ops, "/metricType", plan.MetricType, state.MetricType)
	ops = appendStringPatch(ops, "/unitOfMeasurement", plan.UnitOfMeasurement, state.UnitOfMeasurement)
	ops = appendStringPatch(ops, "/granularity", plan.Granularity, state.Granularity)
	ops = appendStringPatch(ops, "/customUnitOfMeasurement", plan.CustomUnitOfMeasurement, state.CustomUnitOfMeasurement)

	if !plan.ExpressionCode.Equal(state.ExpressionCode) || !plan.ExpressionLanguage.Equal(state.ExpressionLanguage) {
		if plan.ExpressionCode.IsNull() && plan.ExpressionLanguage.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/metricExpression"})
		} else {
			expression := map[string]any{}
			if !plan.ExpressionCode.IsNull() {
				expression["code"] = plan.ExpressionCode.ValueString()
			}
			if !plan.ExpressionLanguage.IsNull() {
				expression["language"] = plan.ExpressionLanguage.ValueString()
			}
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/metricExpression",
				Value: expression,
			})
		}
	}

	return ops
}

// appendStringPatch adds an add or remove operation for an optional scalar.
func appendStringPatch(ops []ometa.JSONPatchOp, path string, plan, state types.String) []ometa.JSONPatchOp {
	if plan.Equal(state) {
		return ops
	}
	if plan.IsNull() {
		return append(ops, ometa.JSONPatchOp{Op: "remove", Path: path})
	}
	return append(ops, ometa.JSONPatchOp{Op: "add", Path: path, Value: plan.ValueString()})
}

// requiresReplaceOnRemoval replaces the entity when an optional attribute is
// removed. OpenMetadata keeps the previous metric expression when a patch sets
// it to null, so an in-place removal would produce an inconsistent result.
func requiresReplaceOnRemoval() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			if !req.StateValue.IsNull() && req.PlanValue.IsNull() {
				resp.RequiresReplace = true
			}
		},
		"Removing this attribute replaces the metric",
		"OpenMetadata cannot clear an existing metric expression with a patch.",
	)
}

// createMetricType maps a Terraform string to the create payload enum.
func createMetricType(value types.String) *ometa.CreateMetricMetricType {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	converted := ometa.CreateMetricMetricType(value.ValueString())
	return &converted
}

// createUnitOfMeasurement maps a Terraform string to the create payload enum.
func createUnitOfMeasurement(value types.String) *ometa.CreateMetricUnitOfMeasurement {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	converted := ometa.CreateMetricUnitOfMeasurement(value.ValueString())
	return &converted
}

// createGranularity maps a Terraform string to the create payload enum.
func createGranularity(value types.String) *ometa.CreateMetricGranularity {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	converted := ometa.CreateMetricGranularity(value.ValueString())
	return &converted
}

// createExpression builds the optional metric expression payload.
func createExpression(code, language types.String) *ometa.MetricExpression {
	if code.IsNull() && language.IsNull() {
		return nil
	}
	expression := &ometa.MetricExpression{}
	if !code.IsNull() {
		expression.Code = convert.StringPointer(code)
	}
	if !language.IsNull() {
		converted := ometa.MetricExpressionLanguage(language.ValueString())
		expression.Language = &converted
	}
	return expression
}
