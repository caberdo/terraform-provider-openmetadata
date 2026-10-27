package ingestion_pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*ingestionPipelineResource)(nil)
	_ resource.ResourceWithConfigure   = (*ingestionPipelineResource)(nil)
	_ resource.ResourceWithImportState = (*ingestionPipelineResource)(nil)
)

// ingestionPipelineServiceTypes are the service types an ingestion pipeline can
// belong to.
var ingestionPipelineServiceTypes = []string{
	"databaseService",
	"dashboardService",
	"messagingService",
	"pipelineService",
	"mlmodelService",
	"metadataService",
	"storageService",
	"searchService",
	"driveService",
	"apiService",
	"securityService",
	"llmService",
}

// ingestionPipelineTypes are the workflow kinds the server accepts.
var ingestionPipelineTypes = []string{
	"metadata",
	"usage",
	"lineage",
	"profiler",
	"autoClassification",
	"TestSuite",
	"dataInsight",
	"elasticSearchReindex",
	"dbt",
	"application",
}

// ingestionPipelineLoggerLevels are the log levels the server accepts.
var ingestionPipelineLoggerLevels = []string{"DEBUG", "INFO", "WARN", "ERROR"}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &ingestionPipelineResource{}
}

type ingestionPipelineResource struct {
	client *client.Client
}

func (r *ingestionPipelineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingestion_pipeline"
}

func (r *ingestionPipelineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata ingestion pipeline: the workflow definition that reads " +
			"metadata, usage, lineage or profiling data from a service.",
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
				Description: "Ingestion pipeline name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name, which is <service>.<name>.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the pipeline.",
				Default:     stringdefault.StaticString(""),
			},
			"service": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the service this pipeline ingests from. The server fixes it at creation, so changing it replaces the pipeline.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service_type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Type of the referenced service, for example databaseService or dashboardService. Defaults to databaseService and is fixed at creation, so changing it replaces the pipeline.",
				Default:     stringdefault.StaticString("databaseService"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(ingestionPipelineServiceTypes...),
				},
			},
			"pipeline_type": rschema.StringAttribute{
				Required:    true,
				Description: "Workflow kind such as metadata, profiler or lineage. The server fixes it at creation, so changing it replaces the pipeline.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(ingestionPipelineTypes...),
				},
			},
			"source_config": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Source configuration as a JSON object string, for example {\"type\": \"DatabaseMetadata\"}. Its shape depends on pipeline_type.",
				Default:     stringdefault.StaticString("{}"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					semanticJSONPlanModifier(),
				},
				Validators: []validator.String{
					jsonObjectValidator{},
				},
			},
			"airflow_config": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Airflow scheduling configuration as a JSON object string, for example {\"scheduleInterval\": \"0 * * * *\"}.",
				Default:     stringdefault.StaticString("{}"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					semanticJSONPlanModifier(),
				},
				Validators: []validator.String{
					jsonObjectValidator{},
				},
			},
			"logger_level": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Log level used by the ingestion workflow. One of DEBUG, INFO, WARN or ERROR.",
				Validators: []validator.String{
					stringvalidator.OneOf(ingestionPipelineLoggerLevels...),
				},
			},
			"enable_streamable_logs": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the workflow streams logs back to the OpenMetadata server. Defaults to false.",
				Default:     booldefault.StaticBool(false),
			},
			"raise_on_error": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the workflow fails instead of reporting partial errors. Defaults to false.",
				Default:     booldefault.StaticBool(false),
			},
			"deployed": rschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the pipeline workflow has been deployed.",
			},
			"enabled": rschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the pipeline is enabled.",
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

func (r *ingestionPipelineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ingestionPipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ingestionPipelineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	service, err := resolveServiceReference(ctx, sdk, plan.ServiceType.ValueString(), plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve ingestion pipeline service",
			fmt.Sprintf("Could not resolve service %q: %s", plan.Service.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	body, err := createBody(plan, service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ingestion pipeline configuration",
			fmt.Sprintf("Could not build the request for %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	created, err := sdk.IngestionPipelines.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata ingestion pipeline",
			fmt.Sprintf("Could not create ingestion pipeline %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ingestionPipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ingestionPipelineModel
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

	found, err := lookupIngestionPipeline(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata ingestion pipeline",
			fmt.Sprintf("Could not read ingestion pipeline %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ingestionPipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ingestionPipelineModel
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
		updated, err := sdk.IngestionPipelines.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata ingestion pipeline",
				fmt.Sprintf("Could not update ingestion pipeline %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupIngestionPipeline(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata ingestion pipeline after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ingestionPipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ingestionPipelineModel
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
	err = sdk.IngestionPipelines.Delete(ctx, state.ID.ValueString(), &ometa.DeleteIngestionPipelineParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata ingestion pipeline",
			fmt.Sprintf("Could not delete ingestion pipeline %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the ingestion pipeline UUID or its fully
// qualified name.
func (r *ingestionPipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// resolveServiceReference looks up the service a pipeline belongs to. The API
// requires a full entity reference (including the UUID), so the FQN alone is
// not enough.
func resolveServiceReference(ctx context.Context, sdk *ometa.Client, serviceType, fqn string) (ometa.EntityReference, error) {
	switch serviceType {
	case "databaseService":
		service, err := sdk.DatabaseServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "dashboardService":
		service, err := sdk.DashboardServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "messagingService":
		service, err := sdk.MessagingServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "pipelineService":
		service, err := sdk.PipelineServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "mlmodelService":
		service, err := sdk.MlModelServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "metadataService":
		service, err := sdk.MetadataServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "storageService":
		service, err := sdk.StorageServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "searchService":
		service, err := sdk.SearchServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "driveService":
		service, err := sdk.DriveServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "apiService":
		service, err := sdk.ApiServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "securityService":
		service, err := sdk.SecurityServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	case "llmService":
		service, err := sdk.LLMServices.GetByName(ctx, fqn, nil)
		if err != nil {
			return ometa.EntityReference{}, err
		}
		return ometa.EntityReference{
			Id:                 service.Id,
			Type:               serviceType,
			Name:               &service.Name,
			FullyQualifiedName: serviceFQN(service.FullyQualifiedName, fqn),
		}, nil
	default:
		return ometa.EntityReference{}, fmt.Errorf("unsupported service_type %q", serviceType)
	}
}

// serviceFQN returns the server-side fully qualified name when present.
func serviceFQN(value *string, fallback string) *string {
	if value == nil {
		return &fallback
	}
	return value
}

// lookupIngestionPipeline resolves a UUID or a fully qualified name, so
// import and Read accept either identifier.
func lookupIngestionPipeline(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.IngestionPipeline, error) {
	found, err := sdk.IngestionPipelines.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.IngestionPipelines.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, tags and the deployment state
// managed outside Terraform are left alone.
func patchOps(plan, state ingestionPipelineModel) []ometa.JSONPatchOp {
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

	if jsonAttributeChanged(plan.SourceConfig, state.SourceConfig) {
		if decoded, err := parseJSON(plan.SourceConfig.ValueString()); err == nil {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/sourceConfig",
				Value: &ometa.SourceConfig{Config: decoded},
			})
		}
	}

	if jsonAttributeChanged(plan.AirflowConfig, state.AirflowConfig) {
		if airflow, err := decodeAirflowConfig(plan.AirflowConfig.ValueString()); err == nil {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/airflowConfig",
				Value: airflow,
			})
		}
	}

	if !plan.LoggerLevel.IsUnknown() {
		switch {
		case plan.LoggerLevel.Equal(state.LoggerLevel):
		case plan.LoggerLevel.IsNull():
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/loggerLevel"})
		default:
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/loggerLevel",
				Value: ometa.IngestionPipelineLoggerLevel(plan.LoggerLevel.ValueString()),
			})
		}
	}

	if !plan.EnableStreamable.Equal(state.EnableStreamable) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/enableStreamableLogs",
			Value: plan.EnableStreamable.ValueBool(),
		})
	}

	if !plan.RaiseOnError.Equal(state.RaiseOnError) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/raiseOnError",
			Value: plan.RaiseOnError.ValueBool(),
		})
	}

	// service, service_type, pipeline_type and name are not patchable: the
	// server fixes them at creation. Those attributes carry RequiresReplace.

	return ops
}

// jsonAttributeChanged reports whether a JSON attribute differs semantically.
// Unknown values are treated as unchanged.
func jsonAttributeChanged(plan, state types.String) bool {
	if plan.IsUnknown() || state.IsUnknown() {
		return false
	}
	if plan.IsNull() || state.IsNull() {
		return !plan.Equal(state)
	}
	return !jsonEqual(plan.ValueString(), state.ValueString())
}

// semanticJSONPlanModifier keeps the prior value when the planned document is
// semantically equal to state, so a reformatted but equivalent JSON attribute
// does not produce a perpetual diff.
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

// jsonObjectValidator rejects attribute values that are not JSON objects.
type jsonObjectValidator struct{}

func (v jsonObjectValidator) Description(_ context.Context) string {
	return "Value must be a JSON object."
}

func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonObjectValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &decoded); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON object",
			fmt.Sprintf("Value must be a JSON object: %s.", err),
		)
	}
}
