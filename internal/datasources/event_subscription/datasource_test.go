package event_subscription

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_event_subscription" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_event_subscription")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
	}

	for _, attribute := range []string{"name", "description", "alert_type", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	if resources, ok := resp.Schema.Attributes["resources"].(dschema.ListAttribute); !ok || !resources.Computed {
		t.Fatal("resources must be a computed list of strings")
	}

	destinations, ok := resp.Schema.Attributes["destinations"].(dschema.ListNestedAttribute)
	if !ok || !destinations.Computed {
		t.Fatalf("destinations must be a computed list of objects")
	}
	for _, nested := range []string{"category", "type", "enabled", "config"} {
		if !destinations.NestedObject.Attributes[nested].IsComputed() {
			t.Fatalf("destinations.%s must be computed", nested)
		}
	}

	if enabled, ok := resp.Schema.Attributes["enabled"].(dschema.BoolAttribute); !ok || !enabled.Computed {
		t.Fatal("enabled must be a computed bool")
	}
	if version, ok := resp.Schema.Attributes["version"].(dschema.Float64Attribute); !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "finance-updates"
	displayName := "Finance updates"
	pollInterval := int32(15)
	retries := int32(2)
	destinationEnabled := true
	description := "Notifies the finance team."

	var model eventSubscriptionModel
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
		},
		PollInterval: &pollInterval,
		Retries:      &retries,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "finance-updates" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.AlertType.ValueString() != "Notification" {
		t.Fatalf("AlertType = %q", model.AlertType.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Fatal("Enabled = false, want the true default")
	}
	if model.PollInterval.ValueInt64() != 15 {
		t.Fatalf("PollInterval = %d", model.PollInterval.ValueInt64())
	}
	if model.Retries.ValueInt64() != 2 {
		t.Fatalf("Retries = %d", model.Retries.ValueInt64())
	}

	var destinations []struct {
		Category string `tfsdk:"category"`
		Type     string `tfsdk:"type"`
		Enabled  bool   `tfsdk:"enabled"`
		Config   string `tfsdk:"config"`
	}
	if diags := model.Destinations.ElementsAs(context.Background(), &destinations, false); diags.HasError() {
		t.Fatalf("Destinations diagnostics = %v", diags.Errors())
	}
	if len(destinations) != 1 {
		t.Fatalf("len(Destinations) = %d, want 1", len(destinations))
	}
	if destinations[0].Category != "Followers" || destinations[0].Type != "ActivityFeed" {
		t.Fatalf("destination[0] = %+v", destinations[0])
	}
	if destinations[0].Config != `{"limit":10}` {
		t.Fatalf("destination[0].Config = %q", destinations[0].Config)
	}
	if !destinations[0].Enabled {
		t.Fatal("destination[0].Enabled = false")
	}
}
