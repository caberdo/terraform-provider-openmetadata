package pipeline

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ datasource.DataSource              = (*pipelineDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*pipelineDataSource)(nil)
)

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &pipelineDataSource{}
}

type pipelineDataSource struct {
	client *client.Client
}

func (d *pipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (d *pipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata pipeline by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Pipeline UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Pipeline fully qualified name. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "Pipeline name.",
			},
			"service": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the parent pipeline service.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
			},
			"concurrency": dschema.Int64Attribute{
				Computed:    true,
				Description: "Maximum number of parallel task runs.",
			},
			"pipeline_location": dschema.StringAttribute{
				Computed:    true,
				Description: "Location of the pipeline definition.",
			},
			"schedule_interval": dschema.StringAttribute{
				Computed:    true,
				Description: "Cron expression that schedules the pipeline.",
			},
			"source_url": dschema.StringAttribute{
				Computed:    true,
				Description: "Link to the pipeline definition in the orchestrator.",
			},
			"start_date": dschema.Int64Attribute{
				Computed:    true,
				Description: "Pipeline start time as epoch milliseconds.",
			},
			"state": dschema.StringAttribute{
				Computed:    true,
				Description: "Scheduling state, either Active or Inactive.",
			},
			"tasks": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "Tasks that make up the pipeline.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"name": dschema.StringAttribute{
							Computed:    true,
							Description: "Task name.",
						},
						"display_name": dschema.StringAttribute{
							Computed:    true,
							Description: "Human-readable task name.",
						},
						"description": dschema.StringAttribute{
							Computed:    true,
							Description: "Markdown description of the task.",
						},
						"task_type": dschema.StringAttribute{
							Computed:    true,
							Description: "Task type reported by the orchestrator.",
						},
						"task_sql": dschema.StringAttribute{
							Computed:    true,
							Description: "SQL statement executed by the task.",
						},
						"source_url": dschema.StringAttribute{
							Computed:    true,
							Description: "Link to the task definition in the orchestrator.",
						},
						"start_date": dschema.StringAttribute{
							Computed:    true,
							Description: "Task start time reported by the orchestrator.",
						},
						"end_date": dschema.StringAttribute{
							Computed:    true,
							Description: "Task end time reported by the orchestrator.",
						},
						"downstream_tasks": dschema.ListAttribute{
							Computed:    true,
							Description: "Names of tasks that run after this one.",
							ElementType: types.StringType,
						},
					},
				},
			},
			"version": dschema.Float64Attribute{
				Computed:    true,
				Description: "Server-side entity version.",
			},
			"updated_at": dschema.StringAttribute{
				Computed:    true,
				Description: "Last update time in RFC 3339 format.",
			},
			"updated_by": dschema.StringAttribute{
				Computed:    true,
				Description: "User that performed the last update.",
			},
		},
	}
}

func (d *pipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = data.Client
}

func (d *pipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pipelineModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var found *ometa.Pipeline
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		found, err = sdk.Pipelines.GetByID(ctx, config.ID.ValueString(), nil)
	case !config.FullyQualifiedName.IsNull() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.Pipelines.GetByName(ctx, config.FullyQualifiedName.ValueString(), nil)
	default:
		resp.Diagnostics.AddError(
			"Missing pipeline identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata pipeline",
			client.ErrorDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &config)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
