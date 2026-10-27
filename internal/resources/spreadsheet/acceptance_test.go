package spreadsheet_test

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

func TestAccSpreadsheet_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-spreadsheet")
	resourceName := "openmetadata_spreadsheet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSpreadsheetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpreadsheetConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-sheet"),
					resource.TestCheckResourceAttr(resourceName, "service", name),
					resource.TestCheckResourceAttrPair(resourceName, "parent", "openmetadata_directory.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name+"."+name+"-reports."+name+"-sheet"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccSpreadsheetConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Spreadsheet"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "parent", name+"."+name+"-reports"),
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

func testAccSpreadsheetConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "test" {
  name    = "%s-reports"
  service = openmetadata_drive_service.test.fully_qualified_name
}

resource "openmetadata_spreadsheet" "test" {
  name        = "%s-sheet"
  service     = openmetadata_drive_service.test.fully_qualified_name
  parent      = openmetadata_directory.test.fully_qualified_name
  description = %q
}
`, name, name, name, description)
}

func testAccSpreadsheetConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "test" {
  name    = "%s-reports"
  service = openmetadata_drive_service.test.fully_qualified_name
}

resource "openmetadata_spreadsheet" "test" {
  name         = "%s-sheet"
  service      = openmetadata_drive_service.test.fully_qualified_name
  parent       = openmetadata_directory.test.fully_qualified_name
  display_name = "Acceptance Spreadsheet"
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

func testAccCheckSpreadsheetDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_spreadsheet" {
			continue
		}

		_, err := sdk.Spreadsheets.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("spreadsheet %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking spreadsheet %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
