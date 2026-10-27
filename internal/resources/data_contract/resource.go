package data_contract

import (
	"context"
	"fmt"
	"time"

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
	_ resource.Resource                = (*dataContractResource)(nil)
	_ resource.ResourceWithConfigure   = (*dataContractResource)(nil)
	_ resource.ResourceWithImportState = (*dataContractResource)(nil)
)

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &dataContractResource{}
}

type dataContractResource struct {
	client *client.Client
}

func (r *dataContractResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_contract"
}

func (r *dataContractResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata data contract. A contract describes the terms of use, " +
			"quality expectations, and effective period of a table.",
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
				Description: "Data contract name. Changing it replaces the contract.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Server-assigned fully qualified name of the contract.",
			},
			"entity": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the table governed by the contract. Changing it replaces the contract.",
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
				Description: "Markdown description of the contract.",
				Default:     stringdefault.StaticString(""),
			},
			"terms_of_use": rschema.StringAttribute{
				Optional:    true,
				Description: "Terms of use content for the contract.",
			},
			"effective_from": rschema.StringAttribute{
				Optional:    true,
				Description: "Start of the contract's effective period in RFC 3339 format, for example 2026-01-01T00:00:00Z.",
				Validators: []validator.String{
					rfc3339Validator(),
				},
			},
			"effective_until": rschema.StringAttribute{
				Optional:    true,
				Description: "End of the contract's effective period in RFC 3339 format, for example 2026-12-31T23:59:59Z.",
				Validators: []validator.String{
					rfc3339Validator(),
				},
			},
			"entity_status": rschema.StringAttribute{
				Computed:    true,
				Description: "Server-computed status of the governed entity, for example Active.",
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

func (r *dataContractResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dataContractResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataContractModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	entity := plan.Entity.ValueString()
	table, err := sdk.Tables.GetByName(ctx, entity, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to resolve data contract entity",
			fmt.Sprintf("Could not resolve table %q: %s", entity, client.ErrorDetail(err)),
		)
		return
	}

	entityReference := ometa.EntityReference{
		Id:                 table.Id,
		Type:               "table",
		Name:               &table.Name,
		FullyQualifiedName: &entity,
	}

	body := &ometa.CreateDataContract{
		Name:        plan.Name.ValueString(),
		Entity:      entityReference,
		Description: convert.StringPointer(plan.Description),
		DisplayName: convert.StringPointer(plan.DisplayName),
		TermsOfUse:  convert.StringPointer(plan.TermsOfUse),
	}
	body.EffectiveFrom, _ = parseTimestamp(plan.EffectiveFrom)
	body.EffectiveUntil, _ = parseTimestamp(plan.EffectiveUntil)

	created, err := sdk.DataContracts.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata data contract",
			fmt.Sprintf("Could not create data contract %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataContractResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataContractModel
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

	found, err := lookupDataContract(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata data contract",
			fmt.Sprintf("Could not read data contract %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataContractResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dataContractModel
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
		updated, err := sdk.DataContracts.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata data contract",
				fmt.Sprintf("Could not update data contract %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupDataContract(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata data contract after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataContractResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataContractModel
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
	err = sdk.DataContracts.Delete(ctx, state.ID.ValueString(), &ometa.DeleteDataContractParams{
		HardDelete: &hardDelete,
		Recursive:  &recursive,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata data contract",
			fmt.Sprintf("Could not delete data contract %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the contract UUID or its fully qualified name.
func (r *dataContractResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupDataContract resolves a UUID or a fully qualified name. Import and Read
// both go through it so either identifier works.
func lookupDataContract(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.DataContract, error) {
	found, err := sdk.DataContracts.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.DataContracts.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// name and entity are fixed at creation and carry RequiresReplace instead.
func patchOps(plan, state dataContractModel) []ometa.JSONPatchOp {
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

	if !plan.TermsOfUse.Equal(state.TermsOfUse) {
		if plan.TermsOfUse.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/termsOfUse"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/termsOfUse",
				Value: map[string]any{"content": plan.TermsOfUse.ValueString()},
			})
		}
	}

	if !plan.EffectiveFrom.Equal(state.EffectiveFrom) {
		if plan.EffectiveFrom.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/effectiveFrom"})
		} else if millis, ok := parseTimestamp(plan.EffectiveFrom); ok {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/effectiveFrom",
				Value: *millis,
			})
		}
	}

	if !plan.EffectiveUntil.Equal(state.EffectiveUntil) {
		if plan.EffectiveUntil.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/effectiveUntil"})
		} else if millis, ok := parseTimestamp(plan.EffectiveUntil); ok {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/effectiveUntil",
				Value: *millis,
			})
		}
	}

	return ops
}

// rfc3339Validator rejects effective timestamp strings that time.Parse cannot
// read, so create and patch never see an unparseable value.
func rfc3339Validator() validator.String {
	return timestampValidator{}
}

type timestampValidator struct{}

func (v timestampValidator) Description(_ context.Context) string {
	return "value must be an RFC 3339 timestamp"
}

func (v timestampValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timestampValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := time.Parse(time.RFC3339, req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid timestamp",
			fmt.Sprintf("Value must be an RFC 3339 timestamp such as 2026-01-01T00:00:00Z: %s", err),
		)
	}
}
