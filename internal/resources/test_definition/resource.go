package test_definition

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*testDefinitionResource)(nil)
	_ resource.ResourceWithConfigure   = (*testDefinitionResource)(nil)
	_ resource.ResourceWithImportState = (*testDefinitionResource)(nil)
)

var (
	entityTypes = []string{
		string(ometa.CreateTestDefinitionEntityTypeTABLE),
		string(ometa.CreateTestDefinitionEntityTypeCOLUMN),
	}

	testPlatforms = []string{
		string(ometa.CreateTestDefinitionTestPlatformsOpenMetadata),
		string(ometa.CreateTestDefinitionTestPlatformsDbt),
		string(ometa.CreateTestDefinitionTestPlatformsDeequ),
		string(ometa.CreateTestDefinitionTestPlatformsGreatExpectations),
		string(ometa.CreateTestDefinitionTestPlatformsSoda),
		string(ometa.CreateTestDefinitionTestPlatformsOther),
	}

	dataQualityDimensions = []string{
		string(ometa.CreateTestDefinitionDataQualityDimensionAccuracy),
		string(ometa.CreateTestDefinitionDataQualityDimensionCompleteness),
		string(ometa.CreateTestDefinitionDataQualityDimensionConsistency),
		string(ometa.CreateTestDefinitionDataQualityDimensionIntegrity),
		string(ometa.CreateTestDefinitionDataQualityDimensionNoDimension),
		string(ometa.CreateTestDefinitionDataQualityDimensionSQL),
		string(ometa.CreateTestDefinitionDataQualityDimensionUniqueness),
		string(ometa.CreateTestDefinitionDataQualityDimensionValidity),
	}

	providers = []string{
		string(ometa.CreateTestDefinitionProviderAutomation),
		string(ometa.CreateTestDefinitionProviderSystem),
		string(ometa.CreateTestDefinitionProviderUser),
	}
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &testDefinitionResource{}
}

type testDefinitionResource struct {
	client *client.Client
}

func (r *testDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_definition"
}

func (r *testDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata test definition, the catalogue entry that describes a data " +
			"quality test: the entity it runs against, the platforms that can execute it, and the SQL it uses.",
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
				Description: "Test definition name. Changing it replaces the definition, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For a test definition this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the test definition.",
				Default:     stringdefault.StaticString(""),
			},
			"entity_type": rschema.StringAttribute{
				Required:    true,
				Description: "Entity the test runs against: TABLE or COLUMN. The server fixes it at creation, so changing it replaces the definition.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(entityTypes...),
				},
			},
			"test_platforms": rschema.ListAttribute{
				Required:    true,
				Description: "Test platforms that can execute the definition. The server fixes the set at creation, so changing it replaces the definition.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(testPlatforms...)),
				},
			},
			"data_quality_dimension": rschema.StringAttribute{
				Optional:    true,
				Description: "Data quality dimension the test measures, for example Accuracy or Completeness.",
				Validators: []validator.String{
					stringvalidator.OneOf(dataQualityDimensions...),
				},
			},
			"supported_data_types": rschema.ListAttribute{
				Optional:    true,
				Description: "Column data types the test supports, for example INT or VARCHAR. Empty means every type.",
				ElementType: types.StringType,
			},
			"supported_services": rschema.ListAttribute{
				Optional:    true,
				Description: "Service types the test supports, for example Mysql or Snowflake. Empty means every service.",
				ElementType: types.StringType,
			},
			"sql_expression": rschema.StringAttribute{
				Optional:    true,
				Description: "SQL the platform runs for the definition. Required by some platforms, server-side validated.",
			},
			"source": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Who owns the definition: automation, system, or user. The server fixes it at creation, so changing it replaces the definition.",
				Default:     stringdefault.StaticString("user"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(providers...),
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

func (r *testDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *testDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan testDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	platforms, platformDiags := convert.StringList(ctx, plan.TestPlatforms)
	resp.Diagnostics.Append(platformDiags...)
	dataTypes, dataTypeDiags := createSupportedDataTypes(ctx, plan.SupportedDataTypes)
	resp.Diagnostics.Append(dataTypeDiags...)
	services, serviceDiags := createSupportedServices(ctx, plan.SupportedServices)
	resp.Diagnostics.Append(serviceDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreateTestDefinition{
		Name:               plan.Name.ValueString(),
		Description:        convert.StringPointer(plan.Description),
		EntityType:         ometa.CreateTestDefinitionEntityType(plan.EntityType.ValueString()),
		TestPlatforms:      createTestPlatforms(platforms),
		SupportedDataTypes: dataTypes,
		SupportedServices:  services,
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)
	body.DataQualityDimension = createDataQualityDimension(plan.DataQualityDimension)
	body.SqlExpression = convert.StringPointer(plan.SQLExpression)
	body.Provider = createProvider(plan.Provider)

	created, err := sdk.TestDefinitions.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata test definition",
			fmt.Sprintf("Could not create test definition %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state testDefinitionModel
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

	found, err := lookupTestDefinition(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata test definition",
			fmt.Sprintf("Could not read test definition %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *testDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state testDefinitionModel
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

	ops, diags := patchOps(ctx, plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(ops) > 0 {
		updated, err := sdk.TestDefinitions.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata test definition",
				fmt.Sprintf("Could not update test definition %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupTestDefinition(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata test definition after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state testDefinitionModel
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
	err = sdk.TestDefinitions.Delete(ctx, state.ID.ValueString(), &ometa.DeleteTestDefinitionParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata test definition",
			fmt.Sprintf("Could not delete test definition %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the test definition UUID or its fully qualified
// name.
func (r *testDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupTestDefinition resolves a UUID or a fully qualified name. Import and
// Read both go through it so either identifier works.
func lookupTestDefinition(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.TestDefinition, error) {
	found, err := sdk.TestDefinitions.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.TestDefinitions.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags and owners managed outside Terraform
// are left alone. name, entity_type, test_platforms, and provider carry
// RequiresReplace and are never patched.
func patchOps(ctx context.Context, plan, state testDefinitionModel) ([]ometa.JSONPatchOp, diag.Diagnostics) {
	var ops []ometa.JSONPatchOp
	var diags diag.Diagnostics

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

	if !plan.DataQualityDimension.IsUnknown() && !plan.DataQualityDimension.Equal(state.DataQualityDimension) {
		if plan.DataQualityDimension.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/dataQualityDimension"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/dataQualityDimension",
				Value: plan.DataQualityDimension.ValueString(),
			})
		}
	}

	ops, diags = appendListIfChanged(ctx, ops, "/supportedDataTypes", plan.SupportedDataTypes, state.SupportedDataTypes, diags)
	ops, diags = appendListIfChanged(ctx, ops, "/supportedServices", plan.SupportedServices, state.SupportedServices, diags)

	if !plan.SQLExpression.IsUnknown() && !plan.SQLExpression.Equal(state.SQLExpression) {
		if plan.SQLExpression.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/sqlExpression"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/sqlExpression",
				Value: plan.SQLExpression.ValueString(),
			})
		}
	}

	return ops, diags
}

// appendListIfChanged adds an add or remove operation for a string list that
// differs from the state.
func appendListIfChanged(
	ctx context.Context,
	ops []ometa.JSONPatchOp,
	pathName string,
	plan, state types.List,
	diags diag.Diagnostics,
) ([]ometa.JSONPatchOp, diag.Diagnostics) {
	if plan.IsUnknown() || plan.Equal(state) {
		return ops, diags
	}

	if plan.IsNull() {
		return append(ops, ometa.JSONPatchOp{Op: "remove", Path: pathName}), diags
	}

	values, listDiags := convert.StringList(ctx, plan)
	diags.Append(listDiags...)
	if diags.HasError() {
		return nil, diags
	}

	return append(ops, ometa.JSONPatchOp{Op: "add", Path: pathName, Value: values}), diags
}

// createTestPlatforms casts the validated platform list to the API enum.
func createTestPlatforms(values []string) []ometa.CreateTestDefinitionTestPlatforms {
	platforms := make([]ometa.CreateTestDefinitionTestPlatforms, len(values))
	for i, value := range values {
		platforms[i] = ometa.CreateTestDefinitionTestPlatforms(value)
	}
	return platforms
}

// createSupportedDataTypes casts the optional data-type list to the API enum.
// A null list is omitted so the server applies its own default.
func createSupportedDataTypes(
	ctx context.Context,
	value types.List,
) (*[]ometa.CreateTestDefinitionSupportedDataTypes, diag.Diagnostics) {
	values, diags := convert.StringList(ctx, value)
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	dataTypes := make([]ometa.CreateTestDefinitionSupportedDataTypes, len(values))
	for i, item := range values {
		dataTypes[i] = ometa.CreateTestDefinitionSupportedDataTypes(item)
	}
	return &dataTypes, diags
}

// createSupportedServices keeps a null list out of the request so the server
// applies its own default, while an explicit empty list is sent as-is.
func createSupportedServices(ctx context.Context, value types.List) (*[]string, diag.Diagnostics) {
	values, diags := convert.StringList(ctx, value)
	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}
	return &values, diags
}

// createDataQualityDimension casts the optional dimension to the API enum.
func createDataQualityDimension(value types.String) *ometa.CreateTestDefinitionDataQualityDimension {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	dimension := ometa.CreateTestDefinitionDataQualityDimension(value.ValueString())
	return &dimension
}

// createProvider casts the defaulted provider to the API enum.
func createProvider(value types.String) *ometa.CreateTestDefinitionProvider {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	provider := ometa.CreateTestDefinitionProvider(value.ValueString())
	return &provider
}
