package learning_resource

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
	_ datasource.DataSource              = (*learningResourceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*learningResourceDataSource)(nil)
)

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &learningResourceDataSource{}
}

type learningResourceDataSource struct {
	client *client.Client
}

func (d *learningResourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_learning_resource"
}

func (d *learningResourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata learning resource by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Learning resource UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Learning resource fully qualified name. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "Learning resource name.",
			},
			"resource_type": dschema.StringAttribute{
				Computed:    true,
				Description: "Content type, one of Article, Video or Storylane.",
			},
			"categories": dschema.ListAttribute{
				Computed:    true,
				Description: "Academy categories the resource belongs to.",
				ElementType: types.StringType,
			},
			"contexts": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "UI contexts the resource appears in.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"page_id": dschema.StringAttribute{
							Computed:    true,
							Description: "Page the resource is shown on.",
						},
						"component_id": dschema.StringAttribute{
							Computed:    true,
							Description: "Component on the page the resource is attached to.",
						},
						"priority": dschema.Int64Attribute{
							Computed:    true,
							Description: "Ordering priority inside the context.",
						},
					},
				},
			},
			"source": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "Location of the content.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"url": dschema.StringAttribute{
							Computed:    true,
							Description: "URL of the content.",
						},
						"provider": dschema.StringAttribute{
							Computed:    true,
							Description: "System that provides the content.",
						},
					},
				},
			},
			"difficulty": dschema.StringAttribute{
				Computed:    true,
				Description: "Difficulty level, one of Intro, Intermediate or Advanced.",
			},
			"estimated_duration": dschema.Int64Attribute{
				Computed:    true,
				Description: "Estimated time to finish the resource, in minutes.",
			},
			"status": dschema.StringAttribute{
				Computed:    true,
				Description: "Publication status, one of Active, Draft or Deprecated.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
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

func (d *learningResourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *learningResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config learningResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var found *ometa.LearningResource
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		found, err = sdk.LearningResources.GetByID(ctx, config.ID.ValueString(), nil)
	case !config.FullyQualifiedName.IsNull() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.LearningResources.GetByName(ctx, config.FullyQualifiedName.ValueString(), nil)
	default:
		resp.Diagnostics.AddError(
			"Missing learning resource identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata learning resource",
			client.ErrorDetail(err),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &config)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
