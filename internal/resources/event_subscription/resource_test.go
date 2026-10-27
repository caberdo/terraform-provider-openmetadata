package event_subscription

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_event_subscription" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_event_subscription")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "alert_type", "resources", "destinations"} {
		if !resp.Schema.Attributes[attribute].IsRequired() {
			t.Fatalf("%s must be required", attribute)
		}
	}

	destinations, ok := resp.Schema.Attributes["destinations"].(rschema.ListNestedAttribute)
	if !ok {
		t.Fatalf("destinations attribute is %T, want ListNestedAttribute", resp.Schema.Attributes["destinations"])
	}
	for _, nested := range []string{"category", "type"} {
		attribute, ok := destinations.NestedObject.Attributes[nested].(rschema.StringAttribute)
		if !ok || !attribute.Required {
			t.Fatalf("destinations.%s must be a required string", nested)
		}
	}
	for _, nested := range []string{"enabled", "config"} {
		if !destinations.NestedObject.Attributes[nested].IsComputed() {
			t.Fatalf("destinations.%s must be computed", nested)
		}
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"version",
		"updated_at",
		"updated_by",
	} {
		if attribute == "version" {
			floatAttribute, ok := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !ok || !floatAttribute.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	enabled, ok := resp.Schema.Attributes["enabled"].(rschema.BoolAttribute)
	if !ok || !enabled.Computed || !enabled.Optional {
		t.Fatal("enabled must be an optional computed bool")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "finance-updates"
	displayName := "Finance updates"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	description := "Notifies the finance team."
	enabled := false
	destinationEnabled := true
	pollInterval := int32(30)
	retries := int32(3)

	var model eventSubscriptionModel
	model.PollInterval = types.Int64Value(60)
	model.Retries = types.Int64Value(1)
	apply(&ometa.EventSubscription{
		Id:                 id,
		Name:               "finance-updates",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		AlertType:          ometa.EventSubscriptionAlertTypeNotification,
		Destinations: []ometa.SubscriptionDestination{
			{
				Category: ometa.SubscriptionDestinationCategoryFollowers,
				Type:     ometa.SubscriptionDestinationTypeActivityFeed,
				Enabled:  &destinationEnabled,
				Config:   map[string]any{"limit": float64(10)},
			},
			{
				Category: ometa.SubscriptionDestinationCategoryOwners,
				Type:     ometa.SubscriptionDestinationTypeWebhook,
			},
		},
		Enabled:      &enabled,
		PollInterval: &pollInterval,
		Retries:      &retries,
		Version:      &version,
		UpdatedAt:    &updatedAt,
		UpdatedBy:    &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "finance-updates" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Notifies the finance team." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.AlertType.ValueString() != "Notification" {
		t.Fatalf("AlertType = %q", model.AlertType.ValueString())
	}
	if model.Enabled.ValueBool() {
		t.Fatal("Enabled = true")
	}
	if model.PollInterval.ValueInt64() != 30 {
		t.Fatalf("PollInterval = %d", model.PollInterval.ValueInt64())
	}
	if model.Retries.ValueInt64() != 3 {
		t.Fatalf("Retries = %d", model.Retries.ValueInt64())
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
	if model.UpdatedBy.ValueString() != updatedBy {
		t.Fatalf("UpdatedBy = %q", model.UpdatedBy.ValueString())
	}

	if model.Destinations.IsNull() {
		t.Fatal("Destinations is null")
	}
	var destinations []destinationModel
	if diags := model.Destinations.ElementsAs(context.Background(), &destinations, false); diags.HasError() {
		t.Fatalf("Destinations diagnostics = %v", diags.Errors())
	}
	if len(destinations) != 2 {
		t.Fatalf("len(Destinations) = %d, want 2", len(destinations))
	}
	if destinations[0].Category.ValueString() != "Followers" || destinations[0].Type.ValueString() != "ActivityFeed" {
		t.Fatalf("destination[0] = %v/%v", destinations[0].Category, destinations[0].Type)
	}
	if destinations[0].Config.ValueString() != `{"limit":10}` {
		t.Fatalf("destination[0].Config = %q", destinations[0].Config.ValueString())
	}
	if !destinations[0].Enabled.ValueBool() {
		t.Fatal("destination[0].Enabled = false")
	}
	if destinations[1].Config.ValueString() != "{}" {
		t.Fatalf("destination[1].Config = %q, want the empty object default", destinations[1].Config.ValueString())
	}
	if !destinations[1].Enabled.ValueBool() {
		t.Fatal("destination[1].Enabled = false, want the true default")
	}
}

func TestApplyKeepsExistingDestinationFormatting(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	existing := testDestinationsList(destinationModel{
		Category: types.StringValue("Followers"),
		Type:     types.StringValue("ActivityFeed"),
		Enabled:  types.BoolValue(true),
		Config:   types.StringValue("{ \"limit\": 10 }"),
	})

	enabled := true
	model := eventSubscriptionModel{Destinations: existing}
	apply(&ometa.EventSubscription{
		Id:        id,
		Name:      "finance-updates",
		AlertType: ometa.EventSubscriptionAlertTypeNotification,
		Destinations: []ometa.SubscriptionDestination{
			{
				Category: ometa.SubscriptionDestinationCategoryFollowers,
				Type:     ometa.SubscriptionDestinationTypeActivityFeed,
				Enabled:  &enabled,
				Config:   map[string]any{"limit": float64(10)},
			},
		},
	}, &model)

	elements := model.Destinations.Elements()
	if len(elements) != 1 {
		t.Fatalf("len(Destinations) = %d", len(elements))
	}
	attributes := elements[0].(types.Object).Attributes()
	if config := attributes["config"].(types.String); config.ValueString() != "{ \"limit\": 10 }" {
		t.Fatalf("Config = %q, want the configured formatting", config.ValueString())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model eventSubscriptionModel
	pollInterval := int32(10)
	retries := int32(4)
	apply(&ometa.EventSubscription{
		Id:           id,
		Name:         "bare",
		PollInterval: &pollInterval,
		Retries:      &retries,
	}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.Enabled.ValueBool() {
		t.Fatal("Enabled = false, want the true default")
	}
	if !model.PollInterval.IsNull() {
		t.Fatalf("PollInterval = %v, want null", model.PollInterval)
	}
	if !model.Retries.IsNull() {
		t.Fatalf("Retries = %v, want null", model.Retries)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
	if !model.AlertType.IsNull() {
		t.Fatalf("AlertType = %v, want null when the server omits it", model.AlertType)
	}
}

func TestPatchOps(t *testing.T) {
	base := eventSubscriptionModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Destinations: testDestinationsList(destinationModel{
			Category: types.StringValue("Followers"),
			Type:     types.StringValue("ActivityFeed"),
			Enabled:  types.BoolValue(true),
			Config:   types.StringValue("{}"),
		}),
		Enabled:      types.BoolValue(true),
		PollInterval: types.Int64Null(),
		Retries:      types.Int64Null(),
	}

	withDestination := func(destination destinationModel) types.List {
		return testDestinationsList(destination)
	}

	tests := []struct {
		name      string
		plan      eventSubscriptionModel
		state     eventSubscriptionModel
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "equivalent destination JSON",
			plan: eventSubscriptionModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Destinations: withDestination(destinationModel{
					Category: types.StringValue("Followers"),
					Type:     types.StringValue("ActivityFeed"),
					Enabled:  types.BoolValue(true),
					Config:   types.StringValue(`{ "limit": 10 }`),
				}),
				Enabled: base.Enabled,
			},
			state: eventSubscriptionModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Destinations: withDestination(destinationModel{
					Category: types.StringValue("Followers"),
					Type:     types.StringValue("ActivityFeed"),
					Enabled:  types.BoolValue(true),
					Config:   types.StringValue(`{"limit":10}`),
				}),
				Enabled: base.Enabled,
			},
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: eventSubscriptionModel{
				DisplayName:  types.StringValue("New display"),
				Description:  base.Description,
				Destinations: base.Destinations,
				Enabled:      base.Enabled,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: eventSubscriptionModel{
				DisplayName:  types.StringNull(),
				Description:  base.Description,
				Destinations: base.Destinations,
				Enabled:      base.Enabled,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: eventSubscriptionModel{
				DisplayName:  base.DisplayName,
				Description:  types.StringValue("after"),
				Destinations: base.Destinations,
				Enabled:      base.Enabled,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "destination changed",
			plan: eventSubscriptionModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Destinations: withDestination(destinationModel{
					Category: types.StringValue("Owners"),
					Type:     types.StringValue("Webhook"),
					Enabled:  types.BoolValue(true),
					Config:   types.StringValue(`{"endpoint":"https://example.com/hook"}`),
				}),
				Enabled: base.Enabled,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/destinations",
			wantOp:   "add",
		},
		{
			name: "enabled turned off",
			plan: eventSubscriptionModel{
				DisplayName:  base.DisplayName,
				Description:  base.Description,
				Destinations: base.Destinations,
				Enabled:      types.BoolValue(false),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/enabled",
			wantOp:    "add",
			wantValue: false,
		},
		{
			name: "poll interval added",
			plan: eventSubscriptionModel{
				DisplayName:  base.DisplayName,
				Description:  base.Description,
				Destinations: base.Destinations,
				Enabled:      base.Enabled,
				PollInterval: types.Int64Value(60),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/pollInterval",
			wantOp:    "add",
			wantValue: int64(60),
		},
		{
			name: "retries removed",
			plan: eventSubscriptionModel{
				DisplayName:  base.DisplayName,
				Description:  base.Description,
				Destinations: base.Destinations,
				Enabled:      base.Enabled,
				Retries:      types.Int64Null(),
			},
			state: eventSubscriptionModel{
				DisplayName:  base.DisplayName,
				Description:  base.Description,
				Destinations: base.Destinations,
				Enabled:      base.Enabled,
				Retries:      types.Int64Value(5),
			},
			wantOps:  1,
			wantPath: "/retries",
			wantOp:   "remove",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(context.Background(), tt.plan, tt.state)
			if len(ops) != tt.wantOps {
				t.Fatalf("len(ops) = %d, want %d: %+v", len(ops), tt.wantOps, ops)
			}
			if tt.wantOps == 0 {
				return
			}
			if ops[0].Path != tt.wantPath {
				t.Fatalf("Path = %q, want %q", ops[0].Path, tt.wantPath)
			}
			if ops[0].Op != tt.wantOp {
				t.Fatalf("Op = %q, want %q", ops[0].Op, tt.wantOp)
			}
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("Value = %v, want %v", ops[0].Value, tt.wantValue)
			}
		})
	}
}

// testDestinationsList builds a nested destination list for a model.
func testDestinationsList(destinations ...destinationModel) types.List {
	elements := make([]attr.Value, 0, len(destinations))
	for _, destination := range destinations {
		elements = append(elements, types.ObjectValueMust(destinationObjectType.AttrTypes, map[string]attr.Value{
			"category": destination.Category,
			"type":     destination.Type,
			"enabled":  destination.Enabled,
			"config":   destination.Config,
		}))
	}
	return types.ListValueMust(destinationObjectType, elements)
}
