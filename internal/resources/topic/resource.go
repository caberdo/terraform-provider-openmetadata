package topic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = (*topicResource)(nil)
	_ resource.ResourceWithConfigure   = (*topicResource)(nil)
	_ resource.ResourceWithImportState = (*topicResource)(nil)
)

var cleanupPolicyValues = []string{
	string(ometa.CreateTopicCleanupPoliciesCompact),
	string(ometa.CreateTopicCleanupPoliciesDelete),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &topicResource{}
}

type topicResource struct {
	client *client.Client
}

func (r *topicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topic"
}

func (r *topicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata topic. A topic belongs to a messaging service and " +
			"records its partitions, retention and message schema.",
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
				Description: "Topic name. Changing it replaces the topic, because the name is part of its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"service": rschema.StringAttribute{
				Required:    true,
				Description: "Fully qualified name of the parent messaging service. Changing it replaces the topic.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name, in the form service.topic.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the topic.",
				Default:     stringdefault.StaticString(""),
			},
			"partitions": rschema.Int64Attribute{
				Required:    true,
				Description: "Number of partitions. The server requires it at creation.",
			},
			"maximum_message_size": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum message size in bytes. Filled by the server when omitted.",
			},
			"minimum_in_sync_replicas": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Minimum number of in-sync replicas.",
			},
			"replication_factor": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Replication factor.",
			},
			"retention_size": rschema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Retention size in bytes.",
			},
			"retention_time": rschema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Retention time in milliseconds.",
			},
			"cleanup_policies": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cleanup policies, compact and/or delete. Filled by the server when omitted.",
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(cleanupPolicyValues...)),
				},
			},
			"message_schema": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Message schema as a JSON object string with schemaType, schemaText and schemaFields keys.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					semanticJSONPlanModifier(),
				},
			},
			"topic_config": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Broker configuration as a JSON object string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					semanticJSONPlanModifier(),
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

func (r *topicResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *topicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan topicModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	policies, diags := cleanupPoliciesToAPI(ctx, plan.CleanupPolicies)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreateTopic{
		Name:                  plan.Name.ValueString(),
		Service:               plan.Service.ValueString(),
		Partitions:            int32(plan.Partitions.ValueInt64()),
		Description:           convert.StringPointer(plan.Description),
		DisplayName:           convert.StringPointer(plan.DisplayName),
		MaximumMessageSize:    int32Pointer(plan.MaximumMessageSize),
		MinimumInSyncReplicas: int32Pointer(plan.MinimumInSyncReplicas),
		ReplicationFactor:     int32Pointer(plan.ReplicationFactor),
		RetentionSize:         float64Pointer(plan.RetentionSize),
		RetentionTime:         float64Pointer(plan.RetentionTime),
	}
	if policies != nil {
		body.CleanupPolicies = &policies
	}

	if jsonConfigured(plan.MessageSchema) {
		schema, err := parseMessageSchemaJSON(plan.MessageSchema.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("message_schema"),
				"Invalid message schema JSON",
				fmt.Sprintf("The message_schema attribute must be a JSON object: %s", err),
			)
			return
		}
		body.MessageSchema = schema
	}

	if jsonConfigured(plan.TopicConfig) {
		config, err := parseTopicConfigJSON(plan.TopicConfig.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("topic_config"),
				"Invalid topic configuration JSON",
				fmt.Sprintf("The topic_config attribute must be a JSON object: %s", err),
			)
			return
		}
		body.TopicConfig = config
	}

	created, err := sdk.Topics.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata topic",
			fmt.Sprintf("Could not create topic %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *topicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state topicModel
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

	found, err := lookupTopic(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata topic",
			fmt.Sprintf("Could not read topic %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	resp.Diagnostics.Append(apply(ctx, found, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *topicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state topicModel
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
		updated, err := sdk.Topics.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata topic",
				fmt.Sprintf("Could not update topic %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		resp.Diagnostics.Append(apply(ctx, updated, &plan)...)
	} else {
		found, err := lookupTopic(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata topic after update", client.ErrorDetail(err))
			return
		}
		resp.Diagnostics.Append(apply(ctx, found, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *topicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state topicModel
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
	err = sdk.Topics.Delete(ctx, state.ID.ValueString(), &ometa.DeleteTopicParams{
		HardDelete: &hardDelete,
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata topic",
			fmt.Sprintf("Could not delete topic %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the topic UUID or its fully qualified name.
func (r *topicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupTopic resolves a UUID or a fully qualified name. Read and Import both
// go through it so either identifier works.
func lookupTopic(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.Topic, error) {
	found, err := sdk.Topics.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.Topics.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags, owners and sample data managed outside
// Terraform are left alone. name and service carry RequiresReplace and are
// never patched.
func patchOps(ctx context.Context, plan, state topicModel) ([]ometa.JSONPatchOp, diag.Diagnostics) {
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

	if !plan.Partitions.Equal(state.Partitions) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/partitions",
			Value: plan.Partitions.ValueInt64(),
		})
	}

	ops = appendIntIfChanged(ops, "/maximumMessageSize", plan.MaximumMessageSize, state.MaximumMessageSize)
	ops = appendIntIfChanged(ops, "/minimumInSyncReplicas", plan.MinimumInSyncReplicas, state.MinimumInSyncReplicas)
	ops = appendIntIfChanged(ops, "/replicationFactor", plan.ReplicationFactor, state.ReplicationFactor)

	if !plan.RetentionSize.IsUnknown() && !plan.RetentionSize.Equal(state.RetentionSize) {
		if plan.RetentionSize.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/retentionSize"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/retentionSize",
				Value: plan.RetentionSize.ValueFloat64(),
			})
		}
	}

	if !plan.RetentionTime.IsUnknown() && !plan.RetentionTime.Equal(state.RetentionTime) {
		if plan.RetentionTime.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/retentionTime"})
		} else {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/retentionTime",
				Value: plan.RetentionTime.ValueFloat64(),
			})
		}
	}

	if !plan.CleanupPolicies.Equal(state.CleanupPolicies) && !plan.CleanupPolicies.IsUnknown() {
		if plan.CleanupPolicies.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/cleanupPolicies"})
		} else {
			policies, policyDiags := cleanupPoliciesToAPI(ctx, plan.CleanupPolicies)
			diags.Append(policyDiags...)
			if diags.HasError() {
				return nil, diags
			}
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/cleanupPolicies",
				Value: policies,
			})
		}
	}

	if !plan.MessageSchema.IsUnknown() && !state.MessageSchema.IsUnknown() {
		switch {
		case plan.MessageSchema.IsNull() && state.MessageSchema.IsNull():
		case plan.MessageSchema.IsNull():
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/messageSchema"})
		case state.MessageSchema.IsNull() || !jsonEqual(plan.MessageSchema.ValueString(), state.MessageSchema.ValueString()):
			if schema, err := parseMessageSchemaJSON(plan.MessageSchema.ValueString()); err == nil {
				ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/messageSchema", Value: schema})
			}
		}
	}

	if !plan.TopicConfig.IsUnknown() && !state.TopicConfig.IsUnknown() {
		switch {
		case plan.TopicConfig.IsNull() && state.TopicConfig.IsNull():
		case plan.TopicConfig.IsNull():
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/topicConfig"})
		case state.TopicConfig.IsNull() || !jsonEqual(plan.TopicConfig.ValueString(), state.TopicConfig.ValueString()):
			if config, err := parseTopicConfigJSON(plan.TopicConfig.ValueString()); err == nil {
				ops = append(ops, ometa.JSONPatchOp{Op: "add", Path: "/topicConfig", Value: config})
			}
		}
	}

	return ops, diags
}

// appendIntIfChanged adds a patch operation for an optional number attribute.
func appendIntIfChanged(ops []ometa.JSONPatchOp, path string, plan, state types.Int64) []ometa.JSONPatchOp {
	if plan.IsUnknown() || plan.Equal(state) {
		return ops
	}
	if plan.IsNull() {
		return append(ops, ometa.JSONPatchOp{Op: "remove", Path: path})
	}
	return append(ops, ometa.JSONPatchOp{Op: "add", Path: path, Value: plan.ValueInt64()})
}

// semanticJSONPlanModifier keeps the prior JSON value when the planned
// document is semantically equal to state, so a reformatted but equivalent
// attribute does not produce a perpetual diff.
func semanticJSONPlanModifier() planmodifier.String {
	return jsonEquivalentModifier{}
}

type jsonEquivalentModifier struct{}

func (m jsonEquivalentModifier) Description(_ context.Context) string {
	return "Keeps the prior value when the planned value is JSON-equal to state."
}

func (m jsonEquivalentModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m jsonEquivalentModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	if jsonEqual(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
