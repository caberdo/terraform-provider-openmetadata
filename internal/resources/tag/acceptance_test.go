package tag_test

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

func TestAccTag_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-tag")
	classificationName := tfacctest.RandomWithPrefix("tf-acc-tag-classification")
	resourceName := "openmetadata_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig(classificationName, name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "classification", classificationName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", classificationName+"."+name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccTagConfigUpdated(classificationName, name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Tag"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
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

func testAccTagConfig(classificationName, name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_classification" "parent" {
  name = %q
}

resource "openmetadata_tag" "test" {
  classification = openmetadata_classification.parent.name
  name           = %q
  description    = %q
}
`, classificationName, name, description)
}

func testAccTagConfigUpdated(classificationName, name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_classification" "parent" {
  name = %q
}

resource "openmetadata_tag" "test" {
  classification = openmetadata_classification.parent.name
  name           = %q
  display_name   = "Acceptance Tag"
  description    = %q
}
`, classificationName, name, description)
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

func testAccCheckTagDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_tag" {
			continue
		}

		_, err := sdk.Tags.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("tag %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking tag %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
