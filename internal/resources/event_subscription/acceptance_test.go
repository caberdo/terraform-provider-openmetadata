package event_subscription_test

import (
	"context"
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
)

func TestAccEventSubscription_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-event-subscription")
	resourceName := "openmetadata_event_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "alert_type", "Notification"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.category", "Followers"),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.type", "ActivityFeed"),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.config", "{}"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccEventSubscriptionConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Subscription"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destinations.0.category", "Followers"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDByFullyQualifiedName(resourceName),
			},
		},
	})
}

func testAccEventSubscriptionConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_event_subscription" "test" {
  name        = %q
  alert_type  = "Notification"
  description = %q
  resources   = ["table"]

  destinations = [
    {
      category = "Followers"
      type     = "ActivityFeed"
    },
  ]
}
`, name, description)
}

func testAccEventSubscriptionConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_event_subscription" "test" {
  name         = %q
  alert_type   = "Notification"
  display_name = "Acceptance Subscription"
  description  = %q
  resources    = ["table"]

  destinations = [
    {
      category = "Followers"
      type     = "ActivityFeed"
    },
  ]
}
`, name, description)
}

// importStateIDByFullyQualifiedName exercises the name-based import path rather
// than the UUID path used by the default import step.
func importStateIDByFullyQualifiedName(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		fqn := rs.Primary.Attributes["fully_qualified_name"]
		if fqn == "" {
			return "", fmt.Errorf("resource %s has no fully_qualified_name", resourceName)
		}
		return fqn, nil
	}
}

func testAccCheckEventSubscriptionDestroy(state *terraform.State) error {
	openmetadataClient, err := acctest.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	sdk, err := openmetadataClient.SDK(ctx)
	if err != nil {
		return err
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "openmetadata_event_subscription" {
			continue
		}

		_, err := sdk.EventSubscriptions.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("event subscription %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking event subscription %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
