package pipeline_service

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
	_ resource.Resource                = (*pipelineServiceResource)(nil)
	_ resource.ResourceWithConfigure   = (*pipelineServiceResource)(nil)
	_ resource.ResourceWithImportState = (*pipelineServiceResource)(nil)
)

var pipelineServiceTypes = []string{
	string(ometa.CreatePipelineServiceServiceTypeAirbyte),
	string(ometa.CreatePipelineServiceServiceTypeAirflow),
	string(ometa.CreatePipelineServiceServiceTypeCustomPipeline),
	string(ometa.CreatePipelineServiceServiceTypeDBTCloud),
	string(ometa.CreatePipelineServiceServiceTypeDagster),
	string(ometa.CreatePipelineServiceServiceTypeDataFactory),
	string(ometa.CreatePipelineServiceServiceTypeDatabricksPipeline),
	string(ometa.CreatePipelineServiceServiceTypeDomoPipeline),
	string(ometa.CreatePipelineServiceServiceTypeFivetran),
	string(ometa.CreatePipelineServiceServiceTypeFlink),
	string(ometa.CreatePipelineServiceServiceTypeGluePipeline),
	string(ometa.CreatePipelineServiceServiceTypeKafkaConnect),
	string(ometa.CreatePipelineServiceServiceTypeKinesisFirehose),
	string(ometa.CreatePipelineServiceServiceTypeMatillion),
	string(ometa.CreatePipelineServiceServiceTypeMicrosoftFabricPipeline),
	string(ometa.CreatePipelineServiceServiceTypeMulesoft),
	string(ometa.CreatePipelineServiceServiceTypeNifi),
	string(ometa.CreatePipelineServiceServiceTypeOpenLineage),
	string(ometa.CreatePipelineServiceServiceTypeSSIS),
	string(ometa.CreatePipelineServiceServiceTypeSnowplow),
	string(ometa.CreatePipelineServiceServiceTypeSpark),
	string(ometa.CreatePipelineServiceServiceTypeSpline),
	string(ometa.CreatePipelineServiceServiceTypeStitch),
	string(ometa.CreatePipelineServiceServiceTypeWherescape),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &pipelineServiceResource{}
}

type pipelineServiceResource struct {
	client *client.Client
}

func (r *pipelineServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_service"
}

func (r *pipelineServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata pipeline service. A pipeline service registers a " +
			"connection to an orchestrator so its pipelines can be ingested.",
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
				Description: "Pipeline service name. Changing it replaces the service, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a pipeline service this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the service.",
				Default:     stringdefault.StaticString(""),
			},
			"service_type": rschema.StringAttribute{
				Required:    true,
				Description: "Connector type, for example CustomPipeline, Airflow or GluePipeline. The server fixes it at creation, so changing it replaces the service.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(pipelineServiceTypes...),
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

func (r *pipelineServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pipelineServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pipelineServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreatePipelineService{
		Name:        plan.Name.ValueString(),
		ServiceType: ometa.CreatePipelineServiceServiceType(plan.ServiceType.ValueString()),
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
		body.Connection = &ometa.PipelineConnection{Config: decoded}
	}

	created, err := sdk.PipelineServices.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata pipeline service",
			fmt.Sprintf("Could not create pipeline service %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pipelineServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pipelineServiceModel
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

	found, err := lookupPipelineService(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata pipeline service",
			fmt.Sprintf("Could not read pipeline service %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pipelineServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state pipelineServiceModel
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
		updated, err := sdk.PipelineServices.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata pipeline service",
				fmt.Sprintf("Could not update pipeline service %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupPipelineService(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata pipeline service after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pipelineServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pipelineServiceModel
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
	err = sdk.PipelineServices.Delete(ctx, state.ID.ValueString(), &ometa.DeletePipelineServiceParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata pipeline service",
			fmt.Sprintf("Could not delete pipeline service %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the pipeline service UUID or its fully qualified
// name.
func (r *pipelineServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupPipelineService resolves a UUID or a fully qualified name. The
// connection field is requested explicitly, because the API omits it by
// default.
func lookupPipelineService(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.PipelineService, error) {
	fields := "connection"
	found, err := sdk.PipelineServices.GetByID(ctx, reference, &ometa.GetPipelineServiceByIDParams{
		Fields: &fields,
	})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.PipelineServices.GetByName(ctx, reference, &ometa.GetPipelineServiceByFQNParams{
		Fields: &fields,
	})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, tags and ingestion pipelines managed
// outside Terraform are left alone.
func patchOps(plan, state pipelineServiceModel) []ometa.JSONPatchOp {
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
					Value: &ometa.PipelineConnection{
						Config: decoded,
					},
				})
			}
		}
	}

	// service_type is not patchable: the server fixes it at creation. The
	// attribute carries RequiresReplace instead.

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
