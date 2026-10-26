package stored_procedure

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
	"github.com/caberdo/terraform-provider-openmetadata/internal/providerdata"
)

var (
	_ resource.Resource                = (*storedProcedureResource)(nil)
	_ resource.ResourceWithConfigure   = (*storedProcedureResource)(nil)
	_ resource.ResourceWithImportState = (*storedProcedureResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &storedProcedureResource{}
}

type storedProcedureResource struct {
	client *client.Client
}

func (r *storedProcedureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stored_procedure"
}

func (r *storedProcedureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata stored procedure inside a database schema. The " +
			"procedure body and language are patched in place, while its name, parent schema, " +
			"and type are fixed at creation.",
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
				Description: "Stored procedure name. Changing it replaces the procedure, because the name is part of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"database_schema": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the parent database schema. Changing it replaces the procedure.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name, in the form service.database.schema.name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the stored procedure.",
				Default:     stringdefault.StaticString(""),
			},
			"code": rschema.StringAttribute{
				Required:    true,
				Description: "Stored procedure body.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"language": rschema.StringAttribute{
				Optional:    true,
				Description: "Language the procedure is written in: External, Java, JavaScript, Python, or SQL.",
				Validators: []validator.String{
					stringvalidator.OneOf("External", "Java", "JavaScript", "Python", "SQL"),
				},
			},
			"stored_procedure_type": rschema.StringAttribute{
				Optional:    true,
				Description: "Procedure type: Function, StoredPackage, StoredProcedure, or UDF. The server fixes this at creation, so changing it replaces the procedure.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Function", "StoredPackage", "StoredProcedure", "UDF"),
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

func (r *storedProcedureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storedProcedureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storedProcedureModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	body := &ometa.CreateStoredProcedure{
		Name:                plan.Name.ValueString(),
		DatabaseSchema:      convert.StringPointer(plan.DatabaseSchema),
		Description:         convert.StringPointer(plan.Description),
		DisplayName:         convert.StringPointer(plan.DisplayName),
		StoredProcedureCode: storedProcedureCode(plan),
	}
	if procedureType := createStoredProcedureType(plan); procedureType != nil {
		body.StoredProcedureType = procedureType
	}

	created, err := sdk.StoredProcedures.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata stored procedure",
			fmt.Sprintf("Could not create stored procedure %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storedProcedureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storedProcedureModel
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

	found, err := lookupStoredProcedure(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata stored procedure",
			fmt.Sprintf("Could not read stored procedure %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storedProcedureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state storedProcedureModel
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
		updated, err := sdk.StoredProcedures.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata stored procedure",
				fmt.Sprintf("Could not update stored procedure %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupStoredProcedure(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata stored procedure after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storedProcedureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storedProcedureModel
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
	err = sdk.StoredProcedures.Delete(ctx, state.ID.ValueString(), &ometa.DeleteStoredProcedureParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata stored procedure",
			fmt.Sprintf("Could not delete stored procedure %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the stored procedure UUID or its fully qualified
// name.
func (r *storedProcedureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupStoredProcedure resolves a UUID or a fully qualified name. Import and
// Read both go through it so either identifier works.
func lookupStoredProcedure(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.StoredProcedure, error) {
	found, err := sdk.StoredProcedures.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.StoredProcedures.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so owners, tags, and other metadata managed
// outside Terraform are left alone.
func patchOps(plan, state storedProcedureModel) []ometa.JSONPatchOp {
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

	// The body and its language live in one nested object, so a change to
	// either sends the whole storedProcedureCode value.
	if !plan.Code.Equal(state.Code) || !plan.Language.Equal(state.Language) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/storedProcedureCode",
			Value: storedProcedureCode(plan),
		})
	}

	// stored_procedure_type is not patchable: the server keeps the original
	// value on update. The attribute carries RequiresReplace instead.

	return ops
}

// storedProcedureCode builds the nested code object from the model.
func storedProcedureCode(model storedProcedureModel) ometa.StoredProcedureCode {
	code := ometa.StoredProcedureCode{
		Code: convert.StringPointer(model.Code),
	}
	if !model.Language.IsNull() && !model.Language.IsUnknown() {
		language := ometa.StoredProcedureCodeLanguage(model.Language.ValueString())
		code.Language = &language
	}
	return code
}

// createStoredProcedureType maps the optional type to the creation enum.
func createStoredProcedureType(model storedProcedureModel) *ometa.CreateStoredProcedureStoredProcedureType {
	if model.StoredProcedureType.IsNull() || model.StoredProcedureType.IsUnknown() {
		return nil
	}
	procedureType := ometa.CreateStoredProcedureStoredProcedureType(model.StoredProcedureType.ValueString())
	return &procedureType
}
