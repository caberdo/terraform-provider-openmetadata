package event_subscription

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = (*eventSubscriptionResource)(nil)
	_ resource.ResourceWithConfigure   = (*eventSubscriptionResource)(nil)
	_ resource.ResourceWithImportState = (*eventSubscriptionResource)(nil)
)

var alertTypes = []string{
	string(ometa.EventSubscriptionAlertTypeActivityFeed),
	string(ometa.EventSubscriptionAlertTypeCustom),
	string(ometa.EventSubscriptionAlertTypeGovernanceWorkflowChangeEvent),
	string(ometa.EventSubscriptionAlertTypeNotification),
	string(ometa.EventSubscriptionAlertTypeObservability),
}

var destinationCategories = []string{
	string(ometa.SubscriptionDestinationCategoryAdmins),
	string(ometa.SubscriptionDestinationCategoryAssignees),
	string(ometa.SubscriptionDestinationCategoryExternal),
	string(ometa.SubscriptionDestinationCategoryFollowers),
	string(ometa.SubscriptionDestinationCategoryMentions),
	string(ometa.SubscriptionDestinationCategoryOwners),
	string(ometa.SubscriptionDestinationCategoryTeams),
	string(ometa.SubscriptionDestinationCategoryUsers),
}

var destinationTypes = []string{
	string(ometa.SubscriptionDestinationTypeActivityFeed),
	string(ometa.SubscriptionDestinationTypeEmail),
	string(ometa.SubscriptionDestinationTypeGChat),
	string(ometa.SubscriptionDestinationTypeGovernanceWorkflowChangeEvent),
	string(ometa.SubscriptionDestinationTypeMsTeams),
	string(ometa.SubscriptionDestinationTypeSlack),
	string(ometa.SubscriptionDestinationTypeWebhook),
}

// NewResource is the factory registered in provider.Resources.
func NewResource() resource.Resource {
	return &eventSubscriptionResource{}
}

type eventSubscriptionResource struct {
	client *client.Client
}

func (r *eventSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_subscription"
}

func (r *eventSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an OpenMetadata event subscription. A subscription watches change events " +
			"and delivers them to its destinations, such as the activity feed, Slack or a webhook.",
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
				Description: "Subscription name. Changing it replaces the subscription, because the name is also its fully qualified name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"fully_qualified_name": rschema.StringAttribute{
				Computed:    true,
				Description: "Fully qualified name. For an event subscription this equals name.",
			},
			"display_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Human-readable name shown in the OpenMetadata UI.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Markdown description of the subscription.",
				Default:     stringdefault.StaticString(""),
			},
			"alert_type": rschema.StringAttribute{
				Required:    true,
				Description: "Kind of alert the subscription evaluates. The server fixes it at creation, so changing it replaces the subscription.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(alertTypes...),
				},
			},
			"resources": rschema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Entity types the subscription watches, for example table or testCase. Notification subscriptions require exactly one resource.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"destinations": rschema.ListNestedAttribute{
				Required:    true,
				Description: "Destinations that receive matching events. Replacing the list sends the whole set through a single JSON Patch operation.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: rschema.NestedAttributeObject{
					Attributes: map[string]rschema.Attribute{
						"category": rschema.StringAttribute{
							Required:    true,
							Description: "Audience of the destination, for example Followers, Owners or Teams.",
							Validators: []validator.String{
								stringvalidator.OneOf(destinationCategories...),
							},
						},
						"type": rschema.StringAttribute{
							Required:    true,
							Description: "Transport of the destination, for example ActivityFeed, Email, Slack or Webhook.",
							Validators: []validator.String{
								stringvalidator.OneOf(destinationTypes...),
							},
						},
						"enabled": rschema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the destination receives events. Defaults to true.",
							Default:     booldefault.StaticBool(true),
						},
						"config": rschema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Destination configuration as a JSON object string. Its shape depends on type.",
							Default:     stringdefault.StaticString("{}"),
						},
					},
				},
			},
			"enabled": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the subscription is active. Defaults to true.",
				Default:     booldefault.StaticBool(true),
			},
			"poll_interval": rschema.Int64Attribute{
				Optional:    true,
				Description: "Seconds between polling runs for observability subscriptions. The server default applies when omitted.",
			},
			"retries": rschema.Int64Attribute{
				Optional:    true,
				Description: "Delivery retries before an event is dropped. The server default applies when omitted.",
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

func (r *eventSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *eventSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan eventSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sdk, err := r.client.SDK(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to authenticate to OpenMetadata", client.ErrorDetail(err))
		return
	}

	destinations, err := destinationsToAPI(ctx, plan.Destinations)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("destinations"), "Invalid destination configuration", err.Error())
		return
	}

	resources, diags := convert.StringList(ctx, plan.Resources)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &ometa.CreateEventSubscription{
		Name:         plan.Name.ValueString(),
		AlertType:    ometa.CreateEventSubscriptionAlertType(plan.AlertType.ValueString()),
		Description:  convert.StringPointer(plan.Description),
		Destinations: &destinations,
		Resources:    &resources,
		Enabled:      convert.BoolPointer(plan.Enabled),
		PollInterval: int32Pointer(plan.PollInterval),
		Retries:      int32Pointer(plan.Retries),
	}
	body.DisplayName = convert.StringPointer(plan.DisplayName)

	created, err := sdk.EventSubscriptions.Create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create OpenMetadata event subscription",
			fmt.Sprintf("Could not create event subscription %q: %s", plan.Name.ValueString(), client.ErrorDetail(err)),
		)
		return
	}

	apply(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state eventSubscriptionModel
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

	found, err := lookupEventSubscription(ctx, sdk, reference)
	if err != nil {
		if client.NotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read OpenMetadata event subscription",
			fmt.Sprintf("Could not read event subscription %q: %s", reference, client.ErrorDetail(err)),
		)
		return
	}

	apply(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eventSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state eventSubscriptionModel
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

	if _, err := destinationsToAPI(ctx, plan.Destinations); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("destinations"), "Invalid destination configuration", err.Error())
		return
	}

	ops := patchOps(ctx, plan, state)
	if len(ops) > 0 {
		updated, err := sdk.EventSubscriptions.Patch(ctx, state.ID.ValueString(), ops)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update OpenMetadata event subscription",
				fmt.Sprintf("Could not update event subscription %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
			)
			return
		}
		apply(updated, &plan)
	} else {
		found, err := lookupEventSubscription(ctx, sdk, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read OpenMetadata event subscription after update", client.ErrorDetail(err))
			return
		}
		apply(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The SDK exposes no Delete for event subscriptions, but the REST API
	// supports a permanent delete at DELETE /events/subscriptions/{id}.
	err := r.client.Delete(ctx, "events/subscriptions/"+url.PathEscape(state.ID.ValueString()), url.Values{
		"hardDelete": {"true"},
		"recursive":  {"true"},
	})
	if err != nil && !client.NotFound(err) {
		resp.Diagnostics.AddError(
			"Failed to delete OpenMetadata event subscription",
			fmt.Sprintf("Could not delete event subscription %q: %s", state.FullyQualifiedName.ValueString(), client.ErrorDetail(err)),
		)
	}
}

// ImportState accepts either the event subscription UUID or its fully
// qualified name.
func (r *eventSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// lookupEventSubscription resolves a UUID or a fully qualified name. Import
// and Read both go through it so either identifier works.
func lookupEventSubscription(ctx context.Context, sdk *ometa.Client, reference string) (*ometa.EventSubscription, error) {
	found, err := sdk.EventSubscriptions.GetByID(ctx, reference, nil)
	if err == nil {
		return found, nil
	}
	if !client.NotFound(err) && !client.BadRequest(err) {
		return nil, err
	}
	return sdk.EventSubscriptions.GetByName(ctx, reference, nil)
}

// patchOps builds the JSON Patch document for the fields this resource owns.
// Only changed fields are sent, so tags and owners managed outside Terraform
// are left alone. name and alert_type are fixed at creation and carry
// RequiresReplace instead.
func patchOps(ctx context.Context, plan, state eventSubscriptionModel) []ometa.JSONPatchOp {
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

	if !plan.Destinations.IsUnknown() && !destinationsEqual(plan.Destinations, state.Destinations) {
		if plan.Destinations.IsNull() {
			ops = append(ops, ometa.JSONPatchOp{Op: "remove", Path: "/destinations"})
		} else if destinations, err := destinationsToAPI(ctx, plan.Destinations); err == nil {
			ops = append(ops, ometa.JSONPatchOp{
				Op:    "add",
				Path:  "/destinations",
				Value: destinations,
			})
		}
	}

	if !plan.Resources.IsNull() && !plan.Resources.IsUnknown() && !plan.Resources.Equal(state.Resources) {
		ops = append(ops, ometa.JSONPatchOp{
			Op:    "add",
			Path:  "/filteringRules/resources",
			Value: listStrings(plan.Resources),
		})
	}

	ops = append(ops, boolPatchOps("/enabled", plan.Enabled, state.Enabled)...)
	ops = append(ops, intPatchOps("/pollInterval", plan.PollInterval, state.PollInterval)...)
	ops = append(ops, intPatchOps("/retries", plan.Retries, state.Retries)...)

	return ops
}

// boolPatchOps emits add/remove operations for a changeable optional bool.
func boolPatchOps(attribute string, plan, state types.Bool) []ometa.JSONPatchOp {
	if plan.Equal(state) {
		return nil
	}
	if plan.IsNull() || plan.IsUnknown() {
		return []ometa.JSONPatchOp{{Op: "remove", Path: attribute}}
	}
	return []ometa.JSONPatchOp{{Op: "add", Path: attribute, Value: plan.ValueBool()}}
}

// intPatchOps emits add/remove operations for a changeable optional number.
func intPatchOps(attribute string, plan, state types.Int64) []ometa.JSONPatchOp {
	if plan.Equal(state) {
		return nil
	}
	if plan.IsNull() || plan.IsUnknown() {
		return []ometa.JSONPatchOp{{Op: "remove", Path: attribute}}
	}
	return []ometa.JSONPatchOp{{Op: "add", Path: attribute, Value: plan.ValueInt64()}}
}

// listStrings renders a Terraform string list as a Go slice.
func listStrings(value types.List) []string {
	elements := value.Elements()
	values := make([]string, 0, len(elements))
	for _, element := range elements {
		if text, ok := element.(types.String); ok {
			values = append(values, text.ValueString())
		}
	}
	return values
}
