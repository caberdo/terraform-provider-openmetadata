package test_case

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
	_ datasource.DataSource              = (*testCaseDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*testCaseDataSource)(nil)
)

// testCaseFields asks the API to include the references the data source exposes.
var testCaseFields = "testDefinition,testSuite"

// NewDataSource is the factory registered in provider.DataSources.
func NewDataSource() datasource.DataSource {
	return &testCaseDataSource{}
}

type testCaseDataSource struct {
	client *client.Client
}

func (d *testCaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_case"
}

func (d *testCaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Looks up a single OpenMetadata test case by id or fully qualified name.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Test case UUID. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("fully_qualified_name")),
				},
			},
			"fully_qualified_name": dschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Test case fully qualified name. Exactly one of id or fully_qualified_name must be set.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id")),
				},
			},
			"name": dschema.StringAttribute{
				Computed:    true,
				Description: "Test case name.",
			},
			"display_name": dschema.StringAttribute{
				Computed:    true,
				Description: "Human-readable name.",
			},
			"description": dschema.StringAttribute{
				Computed:    true,
				Description: "Markdown description.",
			},
			"entity_link": dschema.StringAttribute{
				Computed:    true,
				Description: "Entity the test runs against, in the form <#E::table::service.database.schema.table>.",
			},
			"test_definition": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the test definition.",
			},
			"entity_fqn": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the entity under test.",
			},
			"test_suite": dschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the owning test suite.",
			},
			"parameter_values": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "Parameters passed to the test definition.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"name": dschema.StringAttribute{
							Computed:    true,
							Description: "Parameter name.",
						},
						"value": dschema.StringAttribute{
							Computed:    true,
							Description: "Parameter value.",
						},
					},
				},
			},
			"dimension_columns": dschema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Columns used to group the test results into dimensions.",
			},
			"compute_passed_failed_row_count": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the server also counts passing and failing rows.",
			},
			"use_dynamic_assertion": dschema.BoolAttribute{
				Computed:    true,
				Description: "Whether the test evaluates rows dynamically.",
			},
			"top_dimensions": dschema.Int64Attribute{
				Computed:    true,
				Description: "Number of top dimensions kept in the test results.",
			},
			"test_case_status": dschema.StringAttribute{
				Computed:    true,
				Description: "Result of the latest run, for example Success, Failed or Aborted.",
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

func (d *testCaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *testCaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config testCaseModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := d.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	var found *ometa.TestCase
	switch {
	case !config.ID.IsNull() && config.ID.ValueString() != "":
		found, err = sdk.TestCases.GetByID(ctx, config.ID.ValueString(), &ometa.Get3Params{Fields: &testCaseFields})
	case !config.FullyQualifiedName.IsNull() && config.FullyQualifiedName.ValueString() != "":
		found, err = sdk.TestCases.GetByName(ctx, config.FullyQualifiedName.ValueString(), &ometa.GetTestCaseByNameParams{Fields: &testCaseFields})
	default:
		resp.Diagnostics.AddError(
			"Missing test case identifier",
			"Set exactly one of id or fully_qualified_name.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata test case",
			client.ErrorDetail(err),
		)
		return
	}

	apply(found, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
