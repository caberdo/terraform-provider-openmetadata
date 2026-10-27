package learning_resource_test

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

func TestAccLearningResource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-learning-resource")
	resourceName := "openmetadata_learning_resource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLearningResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLearningResourceConfig(suffix, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suffix+"-resource"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", "Article"),
					resource.TestCheckResourceAttr(resourceName, "categories.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "categories.0", "AI"),
					resource.TestCheckResourceAttr(resourceName, "contexts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "contexts.0.page_id", "home"),
					resource.TestCheckResourceAttr(resourceName, "contexts.0.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.url", "https://example.com/academy"),
					resource.TestCheckResourceAttr(resourceName, "difficulty", "Intro"),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccLearningResourceConfig(suffix, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "Academy Resource"),
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

func testAccLearningResourceConfig(suffix, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_learning_resource" "test" {
  name          = %q
  resource_type = "Article"
  display_name  = "Academy Resource"
  description   = %q
  categories    = ["AI"]
  difficulty    = "Intro"
  status        = "Active"

  contexts = [
    {
      page_id  = "home"
      priority = 1
    },
  ]

  source = [
    {
      url = "https://example.com/academy"
    },
  ]
}
`, suffix+"-resource", description)
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

func testAccCheckLearningResourceDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_learning_resource" {
			continue
		}

		_, err := sdk.LearningResources.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("learning resource %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking learning resource %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
