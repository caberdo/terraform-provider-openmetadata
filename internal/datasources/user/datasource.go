package user

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
	_ datasource.DataSource              = (*userDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userDataSource)(nil)
)

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.Client
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata user by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User fully qualified name. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "User name.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
			},
			"email": dschema.StringAttribute{
				Computed:    true,
				Description: "Email address.",
			},
			"is_admin": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the user is an administrator.",
			},
			"is_bot": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the user backs an OpenMetadata bot.",
			},
			"teams": dschema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of the teams the user belongs to.",
			},
			"roles": dschema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Fully qualified names of the roles assigned to the user.",
			},
			"default_persona": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the user's default persona.",
			},
			"timezone": dschema.StringAttribute{
				Computed:    true,
				Description: "IANA timezone of the user.",
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

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var found *ometa.User
	// Relationship lists are omitted unless explicitly requested.
	fields := "teams,roles,defaultPersona"
	switch {
	case !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != "":
		found, err = sdk.Users.GetByID(ctx, config.ID.ValueString(), &ometa.GetUserByIDParams{Fields: &fields})
	case !config.FullyQualifiedName.IsNull() && !config.FullyQualifiedName.IsUnknown() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.Users.GetByName(ctx, config.FullyQualifiedName.ValueString(), &ometa.GetUserByFQNParams{Fields: &fields})
	default:
		resp.Diagnostics.AddError(
			"Missing user identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata user",
			fmt.Sprintf("Could not read user: %s", client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
