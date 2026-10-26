package container_test

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

func TestAccContainer_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-container")
	resourceName := "openmetadata_container.child"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-child"),
					resource.TestCheckResourceAttr(resourceName, "service", name),
					resource.TestCheckResourceAttrPair(resourceName, "parent", "openmetadata_container.parent", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name+"."+name+"-parent."+name+"-child"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccContainerConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Container"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "parent", name+"."+name+"-parent"),
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

func testAccContainerConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_storage_service" "test" {
  name         = %q
  service_type = "S3"
}

resource "openmetadata_container" "parent" {
  name        = "%s-parent"
  service     = openmetadata_storage_service.test.name
  description = "Parent container"
}

resource "openmetadata_container" "child" {
  name        = "%s-child"
  service     = openmetadata_storage_service.test.name
  parent      = openmetadata_container.parent.fully_qualified_name
  description = %q
}
`, name, name, name, description)
}

func testAccContainerConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_storage_service" "test" {
  name         = %q
  service_type = "S3"
}

resource "openmetadata_container" "parent" {
  name        = "%s-parent"
  service     = openmetadata_storage_service.test.name
  description = "Parent container"
}

resource "openmetadata_container" "child" {
  name         = "%s-child"
  service      = openmetadata_storage_service.test.name
  parent       = openmetadata_container.parent.fully_qualified_name
  display_name = "Acceptance Container"
  description  = %q
}
`, name, name, name, description)
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

func testAccCheckContainerDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_container" {
			continue
		}

		_, err := sdk.Containers.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("container %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking container %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
