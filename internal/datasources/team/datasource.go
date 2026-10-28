package team

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
	_ datasource.DataSource              = (*teamDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*teamDataSource)(nil)
)

// teamFields lists the relationship fields the read endpoints omit unless
// requested.
const teamFields = "parents,policies,defaultRoles,defaultPersona,users,teamType,email"

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &teamDataSource{}
}

type teamDataSource struct {
	client *client.Client
}

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata team by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Team UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Team fully qualified name, which equals the team name. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "Team name.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
			},
			"team_type": dschema.StringAttribute{
				Computed:    true,
				Description: "Kind of team: Department, Division, BusinessUnit, Group or Organization.",
			},
			"email": dschema.StringAttribute{
				Computed:    true,
				Description: "Contact email of the team.",
			},
			"is_joinable": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether users may join the team themselves.",
			},
			"parents": dschema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Fully qualified names of the parent teams.",
			},
			"policies": dschema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Fully qualified names of policies assigned to the team.",
			},
			"default_roles": dschema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Fully qualified names of roles granted to the team's users by default.",
			},
			"default_persona": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the team's default persona.",
			},
			"users": dschema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Names of the users that belong to the team.",
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

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config teamModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	fields := teamFields

	var found *ometa.Team
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		found, err = sdk.Teams.GetByID(ctx, config.ID.ValueString(), &ometa.GetTeamByIDParams{Fields: &fields})
	case !config.FullyQualifiedName.IsNull() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.Teams.GetByName(ctx, config.FullyQualifiedName.ValueString(), &ometa.GetTeamByFQNParams{Fields: &fields})
	default:
		resp.Diagnostics.AddError(
			"Missing team identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata team",
			client.ErrorDetail(err),
		)
		return
	}

	apply(found, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
