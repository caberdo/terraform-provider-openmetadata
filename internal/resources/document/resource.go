package document

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	_ resource.Resource                = (*documentResource)(nil)
	_ resource.ResourceWithConfigure   = (*documentResource)(nil)
	_ resource.ResourceWithImportState = (*documentResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &documentResource{}
}

type documentResource struct {
	client *client.Client
}

func (r *documentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_document"
}

func (r *documentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata document (knowledge panel). Documents are standalone " +
			"articles whose contents live in a free-form JSON payload.",
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
				Description: "Document name. Changing it replaces the document, because the name is fixed at creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"entity_type": rschema.StringAttribute{
				Required:    true,
				Description: "Document flavour, for example KnowledgePanel. Changing it replaces the document.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name, sent to the server verbatim. Changing it replaces the document.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the document.",
				Default:     stringdefault.StaticString(""),
			},
			"data": rschema.StringAttribute{
				Required:    true,
				Description: "Document payload as a JSON object. Whitespace and key order are normalized once the document is read.",
				PlanModifiers: []planmodifier.String{
					canonicalDataPlanModifier(),
				},
				Validators: []validator.String{
					dataJSONValidator(),
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

func (r *documentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *documentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan documentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data, err := parseData(plan.Data.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("data"),
			"Invalid document data",
			fmt.Sprintf("The data attribute must be a JSON object: %s", err),
		)
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateDocument{
		Name:               plan.Name.ValueString(),
		EntityType:         plan.EntityType.ValueString(),
		FullyQualifiedName: plan.FullyQualifiedName.ValueString(),
		Description:        convert.StringPointer(plan.Description),
		Data:               data,
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.Documents.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata document",
			fmt.Sprintf("Could not create document %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *documentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state documentModel
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

	found, err := lookupDocument(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata document",
			fmt.Sprintf("Could not read document %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *documentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state documentModel
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

	ops := patchOps(plan, state)
	if len(ops) > 0 {
		updated, err := sdk.Documents.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata document",
				fmt.Sprintf("Could not update document %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupDocument(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata document after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *documentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state documentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK exposes no Delete for documents, but the REST API supports a
	// permanent delete at DELETE /docStore/{id}.
	err := r.client.Delete(ctx, "docStore/"+url.PathEscape(state.ID.ValueString()), nil)
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata document",
			fmt.Sprintf("Could not delete document %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the document UUID or its fully qualified name.
func (r *documentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupDocument resolves a UUID or a fully qualified name. Import and Read
// both go through it so either identifier works.
func lookupDocument(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Document, error) {
	found, err := sdk.Documents.GetByID(ctx, reference, &ometa.Get2Params{})
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Documents.GetByName(ctx, reference, &ometa.GetDocumentByFQNParams{})
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// name, entity_type and fully_qualified_name are fixed at creation and carry
// RequiresReplace instead.
func patchOps(plan, state documentModel) []ometa.JSONPatchOp {
	var ops []ometa.JSONPatchOp

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

	if !plan.Data.Equal(state.Data) {
		if data, err := parseData(plan.Data.ValueString()); err == nil {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/data",
				Value: data,
			})
		}
	}

	return ops
}

// canonicalDataPlanModifier normalizes the planned JSON payload so equivalent
// documents that differ only in whitespace or key order do not cause a
// perpetual diff.
func canonicalDataPlanModifier() planmodifier.String {
	return dataCanonicalModifier{}
}

type dataCanonicalModifier struct{}

func (m dataCanonicalModifier) Description(_ context.Context) string {
	return "Normalizes the JSON payload to compact, key-sorted form."
}

func (m dataCanonicalModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m dataCanonicalModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	canonical, err := canonicalData(req.PlanValue.ValueString())
	if err != nil || canonical == req.PlanValue.ValueString() {
		return
	}

	resp.PlanValue = types.StringValue(canonical)
}

// dataJSONValidator rejects payloads that are not JSON objects before any API
// call is made.
func dataJSONValidator() validator.String {
	return dataObjectValidator{}
}

type dataObjectValidator struct{}

func (v dataObjectValidator) Description(_ context.Context) string {
	return "value must be a JSON object"
}

func (v dataObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v dataObjectValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := parseData(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid document data",
			fmt.Sprintf("The data attribute must be a JSON object: %s", err),
		)
	}
}
