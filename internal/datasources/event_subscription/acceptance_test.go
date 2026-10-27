package event_subscription_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccEventSubscriptionDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-event-subscription-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSubscriptionDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_event_subscription.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_event_subscription.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_event_subscription.by_name", "alert_type", "Notification"),
					resource.TestCheckResourceAttr("data.openmetadata_event_subscription.by_name", "destinations.#", "1"),
					resource.TestCheckResourceAttr("data.openmetadata_event_subscription.by_name", "destinations.0.type", "ActivityFeed"),
					resource.TestCheckResourceAttrPair("data.openmetadata_event_subscription.by_id", "id", "openmetadata_event_subscription.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_event_subscription.by_id", "fully_qualified_name", "openmetadata_event_subscription.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccEventSubscriptionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_event_subscription" "test" {
  name        = %q
  alert_type  = "Notification"
  description = "Looked up by name"
  resources   = ["table"]

  destinations = [
    {
      category = "Followers"
      type     = "ActivityFeed"
    },
  ]
}

data "openmetadata_event_subscription" "by_name" {
  fully_qualified_name = openmetadata_event_subscription.test.fully_qualified_name
}

data "openmetadata_event_subscription" "by_id" {
  id = openmetadata_event_subscription.test.id
}
`, name)
}
