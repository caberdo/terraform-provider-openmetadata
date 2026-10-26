package classification_test

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

func TestAccClassification_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-classification")
	resourceName := "openmetadata_classification.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckClassificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClassificationConfig(name, "Initial description", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "mutually_exclusive", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccClassificationConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Classification"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
			{
				// The server fixes mutually_exclusive at creation, so the
				// attribute forces replacement instead of an in-place patch.
				Config: testAccClassificationConfig(name, "Updated description", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "mutually_exclusive", "true"),
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

func testAccClassificationConfig(name, description string, mutuallyExclusive bool) string {
	return fmt.Sprintf(`
resource "openmetadata_classification" "test" {
  name               = %q
  description        = %q
  mutually_exclusive = %t
}
`, name, description, mutuallyExclusive)
}

func testAccClassificationConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_classification" "test" {
  name         = %q
  display_name = "Acceptance Classification"
  description  = %q
}
`, name, description)
}

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

func testAccCheckClassificationDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_classification" {
			continue
		}

		_, err := sdk.Classifications.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("classification %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking classification %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
