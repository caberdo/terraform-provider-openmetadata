package file_test

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

func TestAccFile_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-file")
	resourceName := "openmetadata_file.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-file"),
					resource.TestCheckResourceAttr(resourceName, "service", name),
					resource.TestCheckResourceAttrPair(resourceName, "directory", "openmetadata_directory.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name+"."+name+"-dir."+name+"-file"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccFileConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance File"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "directory", name+"."+name+"-dir"),
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

func testAccFileConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "test" {
  name        = "%s-dir"
  service     = openmetadata_drive_service.test.name
  description = "Acceptance directory"
}

resource "openmetadata_file" "test" {
  name        = "%s-file"
  service     = openmetadata_drive_service.test.name
  directory   = openmetadata_directory.test.fully_qualified_name
  description = %q
}
`, name, name, name, description)
}

func testAccFileConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "test" {
  name        = "%s-dir"
  service     = openmetadata_drive_service.test.name
  description = "Acceptance directory"
}

resource "openmetadata_file" "test" {
  name         = "%s-file"
  service      = openmetadata_drive_service.test.name
  directory    = openmetadata_directory.test.fully_qualified_name
  display_name = "Acceptance File"
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

func testAccCheckFileDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_file" {
			continue
		}

		_, err := sdk.Files.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("file %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking file %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
