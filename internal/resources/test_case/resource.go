package test_case

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*testCaseResource)(nil)
	_ resource.ResourceWithConfigure   = (*testCaseResource)(nil)
	_ resource.ResourceWithImportState = (*testCaseResource)(nil)
)

// testCaseFields asks the API to include the references Terraform owns.
var testCaseFields = "testDefinition,testSuite"

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &testCaseResource{}
}

type testCaseResource struct {
	client *client.Client
}

func (r *testCaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test_case"
}

func (r *testCaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata test case. A test case binds a test definition to an " +
			"entity such as a table and records the pass or fail status of each run.",
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
				Description: "Test case name. Changing it replaces the test case.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"entity_link": rschema.StringAttribute{
				Required:    true,
				Description: "Entity the test runs against, in the form <#E::table::service.database.schema.table>. Changing it replaces the test case.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"test_definition": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the test definition, for example tableRowCountToBeBetween. Changing it replaces the test case.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name assigned by the server.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the test case.",
				Default:     stringdefault.StaticString(""),
			},
			"entity_fqn": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the entity under test.",
			},
			"test_suite": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name of the test suite that owns the test case.",
			},
			"parameter_values": rschema.ListNestedAttribute{
				Optional:    true,
				Description: "Parameters passed to the test definition, for example min and max for a row-count test.",
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"name": rschema.StringAttribute{
							Required:    true,
							Description: "Parameter name.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"value": rschema.StringAttribute{
							Optional:    true,
							Description: "Parameter value, always sent as a string.",
						},
					},
				},
			},
			"dimension_columns": rschema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Columns used to group the test results into dimensions.",
			},
			"compute_passed_failed_row_count": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, the server also counts passing and failing rows.",
				Default:     booldefault.StaticBool(false),
			},
			"use_dynamic_assertion": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, the test definition evaluates rows dynamically instead of using a single query for the whole table.",
				Default:     booldefault.StaticBool(false),
			},
			"top_dimensions": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of top dimensions kept in the test results. The server default applies when omitted.",
			},
			"test_case_status": rschema.StringAttribute{
				Computed:    true,
				Description: "Result of the latest run, for example Success, Failed or Aborted.",
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

func (r *testCaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *testCaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan testCaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	parameters, diags := parameterValuesToAPI(ctx, plan.ParameterValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	dimensionColumns, diags := convert.StringList(ctx, plan.DimensionColumns)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var parameterValues *[]ometa.TestCaseParameterValue
	if parameters != nil {
		parameterValues = &parameters
	}
	var dimensionColumnsValue *[]string
	if dimensionColumns != nil {
		dimensionColumnsValue = &dimensionColumns
	}

	body := &ometa.CreateTestCase{
		Name:                        plan.Name.ValueString(),
		EntityLink:                  plan.EntityLink.ValueString(),
		TestDefinition:              plan.TestDefinition.ValueString(),
		Description:                 convert.StringPointer(plan.Description),
		DisplayName:                 convert.StringPointer(plan.DisplayName),
		ParameterValues:             parameterValues,
		DimensionColumns:            dimensionColumnsValue,
		ComputePassedFailedRowCount: convert.BoolPointer(plan.ComputePassedFailedRowCount),
		UseDynamicAssertion:         convert.BoolPointer(plan.UseDynamicAssertion),
		TopDimensions:               int32Pointer(plan.TopDimensions),
	}

	created, err := sdk.TestCases.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata test case",
			fmt.Sprintf("Could not create test case %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testCaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state testCaseModel
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

	found, err := lookupTestCase(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata test case",
			fmt.Sprintf("Could not read test case %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *testCaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state testCaseModel
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
		updated, err := sdk.TestCases.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata test case",
				fmt.Sprintf("Could not update test case %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupTestCase(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata test case after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *testCaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state testCaseModel
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
	err = sdk.TestCases.Delete(ctx, state.ID.ValueString(), &ometa.DeleteTestCaseParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata test case",
			fmt.Sprintf("Could not delete test case %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the test case UUID or its fully qualified name.
func (r *testCaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupTestCase resolves a UUID or a fully qualified name. Read and Import
// both go through it so either identifier works.
func lookupTestCase(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.TestCase, error) {
	found, err := sdk.TestCases.GetByID(ctx, reference, &ometa.Get3Params{Fields: &testCaseFields})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.TestCases.GetByName(ctx, reference, &ometa.GetTestCaseByNameParams{Fields: &testCaseFields})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners, and test results managed
// outside Terraform are left alone. name, entity_link, and test_definition
// carry RequiresReplace and are never patched.
func patchOps(ctx context.Context, plan, state testCaseModel) ([]ometa.JSONPatchOp, diag.Diagnostics) {
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

	if !plan.ParameterValues.Equal(state.ParameterValues) {
		if plan.ParameterValues.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/parameterValues"})
		} else {
			parameters, parameterDiags := parameterValuesToAPI(ctx, plan.ParameterValues)
			diags.Append(parameterDiags...)
			if diags.HasError() {
				return nil, diags
			}
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/parameterValues",
				Value: parameters,
			})
		}
	}

	if !plan.DimensionColumns.Equal(state.DimensionColumns) {
		if plan.DimensionColumns.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/dimensionColumns"})
		} else {
			columns, columnDiags := convert.StringList(ctx, plan.DimensionColumns)
			diags.Append(columnDiags...)
			if diags.HasError() {
				return nil, diags
			}
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/dimensionColumns",
				Value: columns,
			})
		}
	}

	if !plan.ComputePassedFailedRowCount.Equal(state.ComputePassedFailedRowCount) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/computePassedFailedRowCount",
			Value: plan.ComputePassedFailedRowCount.ValueBool(),
		})
	}

	if !plan.UseDynamicAssertion.Equal(state.UseDynamicAssertion) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/useDynamicAssertion",
			Value: plan.UseDynamicAssertion.ValueBool(),
		})
	}

	if !plan.TopDimensions.IsUnknown() && !plan.TopDimensions.Equal(state.TopDimensions) {
		if plan.TopDimensions.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/topDimensions"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/topDimensions",
				Value: plan.TopDimensions.ValueInt64(),
			})
		}
	}

	return ops, diags
}
