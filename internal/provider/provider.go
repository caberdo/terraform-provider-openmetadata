package provider

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	chartdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/chart"
	classificationdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/classification"
	dashboarddatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/dashboard"
	dashboarddatamodeldatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/dashboard_data_model"
	dashboardservicedatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/dashboard_service"
	dataproductdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/data_product"
	databasedatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/database"
	databaseschemadatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/database_schema"
	databaseservicedatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/database_service"
	domaindatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/domain"
	glossarydatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/glossary"
	glossarytermdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/glossary_term"
	messagingservicedatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/messaging_service"
	metricdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/metric"
	pipelinedatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/pipeline"
	pipelineservicedatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/pipeline_service"
	querydatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/query"
	storedproceduredatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/stored_procedure"
	tabledatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/table"
	tagdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/tag"
	topicdatasource "github.com/caberdo/terraform-provider-openmetadata/internal/datasources/topic"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
	chartresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/chart"
	classificationresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/classification"
	dashboardresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/dashboard"
	dashboarddatamodelresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/dashboard_data_model"
	dashboardserviceresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/dashboard_service"
	dataproductresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/data_product"
	databaseresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/database"
	databaseschemaresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/database_schema"
	databaseserviceresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/database_service"
	domainresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/domain"
	glossaryresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/glossary"
	glossarytermresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/glossary_term"
	messagingserviceresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/messaging_service"
	metricresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/metric"
	pipelineresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/pipeline"
	pipelineserviceresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/pipeline_service"
	queryresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/query"
	storedprocedureresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/stored_procedure"
	tableresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/table"
	tagresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/tag"
	topicresource "github.com/caberdo/terraform-provider-openmetadata/internal/resources/topic"
)

// providerTypeName is the Terraform provider type name: the prefix of every
// resource and data source, and of the registry address.
const providerTypeName = "openmetadata"

var _ provider.Provider = (*OpenMetadataProvider)(nil)

// OpenMetadataProvider is the provider entry point.
type OpenMetadataProvider struct {
	version string
}

// New returns a provider factory carrying the build version reported to
// Terraform in Metadata.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpenMetadataProvider{version: version}
	}
}

// Metadata sets the provider type name and version.
func (p *OpenMetadataProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerTypeName
	resp.Version = p.version
}

// Schema defines provider-level configuration. Every attribute can also be
// supplied through the documented environment variable so credentials never
// need to live in version control.
func (p *OpenMetadataProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{
		Description: "Manage OpenMetadata metadata: glossaries, classifications, tags, services, and the " +
			"entities that hang off them. Requires OpenMetadata 1.13 or later.",
		Attributes: map[string]providerschema.Attribute{
			"host": providerschema.StringAttribute{
				Optional: true,
				Description: "OpenMetadata server root including the scheme, for example " +
					"https://openmetadata.example.com. Falls back to OPENMETADATA_HOST.",
			},
			"token": providerschema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "JWT or personal access token sent as a bearer token. Takes precedence over " +
					"username and password. Falls back to OPENMETADATA_TOKEN.",
			},
			"username": providerschema.StringAttribute{
				Optional: true,
				Description: "OpenMetadata user email used to obtain a JWT when token is unset. " +
					"Falls back to OPENMETADATA_USERNAME.",
			},
			"password": providerschema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for username. Falls back to OPENMETADATA_PASSWORD.",
			},
			"api_version": providerschema.StringAttribute{
				Optional: true,
				Description: "OpenMetadata REST API version. Defaults to v1. " +
					"Falls back to OPENMETADATA_API_VERSION.",
			},
			"retry_count": providerschema.Int64Attribute{
				Optional: true,
				Description: "Retries after the initial attempt for retryable responses (429 and 504). " +
					"Defaults to 3. Falls back to OPENMETADATA_RETRY_COUNT.",
			},
			"retry_wait": providerschema.StringAttribute{
				Optional: true,
				Description: "Delay between retries as a Go duration, for example 30s. Defaults to 30s. " +
					"Falls back to OPENMETADATA_RETRY_WAIT.",
			},
		},
	}
}

// providerModel mirrors the provider schema.
type providerModel struct {
	Host       types.String `tfsdk:"host"`
	Token      types.String `tfsdk:"token"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	APIVersion types.String `tfsdk:"api_version"`
	RetryCount types.Int64  `tfsdk:"retry_count"`
	RetryWait  types.String `tfsdk:"retry_wait"`
}

// Configure resolves configuration and hands a client to every resource and
// data source. No network call happens here: the client authenticates lazily on
// first use so `terraform validate` works without a reachable server.
func (p *OpenMetadataProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := clientConfig(model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	openmetadataClient, err := client.New(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Invalid OpenMetadata provider configuration", err.Error())
		return
	}

	tflog.Debug(ctx, "configured OpenMetadata provider", map[string]any{
		"host":        openmetadataClient.Host(),
		"api_version": openmetadataClient.APIVersion(),
		"auth":        authMode(cfg),
	})

	data := &providerdata.ProviderData{Client: openmetadataClient}
	resp.ResourceData = data
	resp.DataSourceData = data
}

// Resources returns the resource factories. Entries stay sorted alphabetically
// so parallel contributors insert in predictable places.
func (p *OpenMetadataProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		chartresource.NewResource,
		classificationresource.NewResource,
		dashboardresource.NewResource,
		dashboarddatamodelresource.NewResource,
		dashboardserviceresource.NewResource,
		dataproductresource.NewResource,
		databaseresource.NewResource,
		databaseschemaresource.NewResource,
		databaseserviceresource.NewResource,
		domainresource.NewResource,
		glossaryresource.NewResource,
		glossarytermresource.NewResource,
		messagingserviceresource.NewResource,
		metricresource.NewResource,
		pipelineresource.NewResource,
		pipelineserviceresource.NewResource,
		queryresource.NewResource,
		storedprocedureresource.NewResource,
		tableresource.NewResource,
		tagresource.NewResource,
		topicresource.NewResource,
	}
}

// DataSources returns the data source factories, sorted alphabetically.
func (p *OpenMetadataProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		chartdatasource.NewDataSource,
		classificationdatasource.NewDataSource,
		dashboarddatasource.NewDataSource,
		dashboarddatamodeldatasource.NewDataSource,
		dashboardservicedatasource.NewDataSource,
		dataproductdatasource.NewDataSource,
		databasedatasource.NewDataSource,
		databaseschemadatasource.NewDataSource,
		databaseservicedatasource.NewDataSource,
		domaindatasource.NewDataSource,
		glossarydatasource.NewDataSource,
		glossarytermdatasource.NewDataSource,
		messagingservicedatasource.NewDataSource,
		metricdatasource.NewDataSource,
		pipelinedatasource.NewDataSource,
		pipelineservicedatasource.NewDataSource,
		querydatasource.NewDataSource,
		storedproceduredatasource.NewDataSource,
		tabledatasource.NewDataSource,
		tagdatasource.NewDataSource,
		topicdatasource.NewDataSource,
	}
}

// clientConfig merges Terraform configuration with environment fallbacks.
// It is pure apart from reading the environment, which keeps it unit-testable.
func clientConfig(model providerModel) (client.Config, diag.Diagnostics) {
	var diags diag.Diagnostics

	host := strings.TrimSpace(resolve(model.Host, "OPENMETADATA_HOST"))
	if host == "" {
		diags.AddAttributeError(
			path.Root("host"),
			"Missing OpenMetadata host",
			"Set host in the provider block or OPENMETADATA_HOST in the environment.",
		)
	}

	token := resolve(model.Token, "OPENMETADATA_TOKEN")
	username := resolve(model.Username, "OPENMETADATA_USERNAME")
	password := resolve(model.Password, "OPENMETADATA_PASSWORD")
	if token == "" && (username == "" || password == "") {
		diags.AddAttributeError(
			path.Root("token"),
			"Missing OpenMetadata credentials",
			"Set token, or set both username and password. The same values can be supplied through "+
				"OPENMETADATA_TOKEN, or OPENMETADATA_USERNAME and OPENMETADATA_PASSWORD.",
		)
	}

	apiVersion := strings.TrimSpace(resolve(model.APIVersion, "OPENMETADATA_API_VERSION"))
	if apiVersion == "" {
		apiVersion = client.DefaultAPIVersion
	}

	retryCount := int64(client.DefaultRetryCount)
	if !model.RetryCount.IsNull() && !model.RetryCount.IsUnknown() {
		retryCount = model.RetryCount.ValueInt64()
	} else if raw := strings.TrimSpace(os.Getenv("OPENMETADATA_RETRY_COUNT")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			diags.AddAttributeError(
				path.Root("retry_count"),
				"Invalid retry_count",
				"OPENMETADATA_RETRY_COUNT must be a whole number.",
			)
		} else {
			retryCount = parsed
		}
	}

	retryWait := client.DefaultRetryWait
	if raw := strings.TrimSpace(resolve(model.RetryWait, "OPENMETADATA_RETRY_WAIT")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			diags.AddAttributeError(
				path.Root("retry_wait"),
				"Invalid retry_wait",
				"Use a Go duration such as 30s or 1m.",
			)
		} else {
			retryWait = parsed
		}
	}

	if diags.HasError() {
		return client.Config{}, diags
	}

	return client.Config{
		Host:       host,
		Token:      token,
		Username:   username,
		Password:   password,
		APIVersion: apiVersion,
		RetryCount: int(retryCount),
		RetryWait:  retryWait,
	}, diags
}

// resolve returns the configured value, falling back to the environment when
// the attribute is null or unknown.
func resolve(value types.String, env string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}
	return os.Getenv(env)
}

// authMode names the configured authentication path for debug logs. It never
// logs the credential itself.
func authMode(cfg client.Config) string {
	if cfg.Token != "" {
		return "token"
	}
	return "login"
}
