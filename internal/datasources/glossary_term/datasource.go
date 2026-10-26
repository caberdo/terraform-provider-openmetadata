package glossary_term

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
	_ datasource.DataSource              = (*glossaryTermDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*glossaryTermDataSource)(nil)
)

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &glossaryTermDataSource{}
}

type glossaryTermDataSource struct {
	client *client.Client
}

func (d *glossaryTermDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_glossary_term"
}

func (d *glossaryTermDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata glossary term by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Glossary term UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Glossary term fully qualified name. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"glossary": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the glossary that owns the term.",
			},
			"parent": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the parent term, if any.",
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "Term name.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
			},
			"synonyms": dschema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Alternative names for the term.",
			},
			"mutually_exclusive": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether a term may belong to only one glossary under the same parent.",
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

func (d *glossaryTermDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *glossaryTermDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config glossaryTermModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var found *ometa.GlossaryTerm
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		found, err = sdk.GlossaryTerms.GetByID(ctx, config.ID.ValueString(), nil)
	case !config.FullyQualifiedName.IsNull() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.GlossaryTerms.GetByName(ctx, config.FullyQualifiedName.ValueString(), nil)
	default:
		resp.Diagnostics.AddError(
			"Missing glossary term identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata glossary term",
			client.ErrorDetail(err),
		)
		return
	}

	apply(found, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
