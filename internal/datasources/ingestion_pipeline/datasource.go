package ingestion_pipeline

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ datasource.DataSource              = (*ingestionPipelineDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ingestionPipelineDataSource)(nil)
)

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &ingestionPipelineDataSource{}
}

type ingestionPipelineDataSource struct {
	client *client.Client
}

func (d *ingestionPipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingestion_pipeline"
}

func (d *ingestionPipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata ingestion pipeline by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Ingestion pipeline UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Ingestion pipeline fully qualified name, which is <service>.<name>. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "Ingestion pipeline name.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
			},
			"service": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the service this pipeline ingests from.",
			},
			"service_type": dschema.StringAttribute{
				Computed:    true,
				Description: "Type of the referenced service.",
			},
			"pipeline_type": dschema.StringAttribute{
				Computed:    true,
				Description: "Workflow kind such as metadata, profiler or lineage.",
			},
			"source_config": dschema.StringAttribute{
				Computed:    true,
				Description: "Source configuration as a JSON object string.",
			},
			"airflow_config": dschema.StringAttribute{
				Computed:    true,
				Description: "Airflow scheduling configuration as a JSON object string.",
			},
			"logger_level": dschema.StringAttribute{
				Computed:    true,
				Description: "Log level used by the ingestion workflow.",
			},
			"enable_streamable_logs": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the workflow streams logs back to the OpenMetadata server.",
			},
			"raise_on_error": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the workflow fails instead of reporting partial errors.",
			},
			"deployed": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the pipeline workflow has been deployed.",
			},
			"enabled": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the pipeline is enabled.",
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

func (d *ingestionPipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ingestionPipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ingestionPipelineModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var found *ometa.IngestionPipeline
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		found, err = sdk.IngestionPipelines.GetByID(ctx, config.ID.ValueString(), nil)
	case !config.FullyQualifiedName.IsNull() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.IngestionPipelines.GetByName(ctx, config.FullyQualifiedName.ValueString(), nil)
	default:
		resp.Diagnostics.AddError(
			"Missing ingestion pipeline identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata ingestion pipeline",
			client.ErrorDetail(err),
		)
		return
	}

	apply(found, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
