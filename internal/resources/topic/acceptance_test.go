package topic_test

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

func TestAccTopic_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-topic")
	resourceName := "openmetadata_topic.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig(suffix, "Initial description", 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suffix+"-topic"),
					resource.TestCheckResourceAttrPair(resourceName, "service", "openmetadata_messaging_service.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "partitions", "3"),
					resource.TestCheckResourceAttr(resourceName, "cleanup_policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cleanup_policies.0", "delete"),
					resource.TestCheckResourceAttrSet(resourceName, "message_schema"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccTopicConfig(suffix, "Updated description", 6),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "partitions", "6"),
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

func testAccTopicConfig(suffix, description string, partitions int) string {
	return fmt.Sprintf(`
resource "openmetadata_messaging_service" "test" {
  name         = %q
  service_type = "CustomMessaging"
}

resource "openmetadata_topic" "test" {
  name        = %q
  service     = openmetadata_messaging_service.test.fully_qualified_name
  description = %q
  partitions  = %d

  cleanup_policies = ["delete"]

  message_schema = jsonencode({
    schemaType = "JSON"
    schemaText = "{\"type\":\"object\",\"properties\":{\"orderId\":{\"type\":\"string\"}}}"
  })
}
`, suffix+"-service", suffix+"-topic", description, partitions)
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

func testAccCheckTopicDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_topic" {
			continue
		}

		_, err := sdk.Topics.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("topic %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking topic %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
