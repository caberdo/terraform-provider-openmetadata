package worksheet_test

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

func TestAccWorksheet_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-worksheet")
	resourceName := "openmetadata_worksheet.test"
	spreadsheetResource := "openmetadata_spreadsheet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckWorksheetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorksheetConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-q1"),
					resource.TestCheckResourceAttrPair(resourceName, "spreadsheet", spreadsheetResource, "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "service", name),
					resource.TestCheckResourceAttrPair(resourceName, "service", "openmetadata_drive_service.test", "fully_qualified_name"),
					checkWorksheetFullyQualifiedName(resourceName, spreadsheetResource),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccWorksheetConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Worksheet"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "spreadsheet", name+".budget"),
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

func testAccWorksheetConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_spreadsheet" "test" {
  name    = "%s-budget"
  service = openmetadata_drive_service.test.name
}

resource "openmetadata_worksheet" "test" {
  name        = "%s-q1"
  spreadsheet = openmetadata_spreadsheet.test.fully_qualified_name
  description = %q
}
`, name, name, name, description)
}

func testAccWorksheetConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_spreadsheet" "test" {
  name    = "%s-budget"
  service = openmetadata_drive_service.test.name
}

resource "openmetadata_worksheet" "test" {
  name         = "%s-q1"
  spreadsheet  = openmetadata_spreadsheet.test.fully_qualified_name
  display_name = "Acceptance Worksheet"
  description  = %q
}
`, name, name, name, description)
}

// checkWorksheetFullyQualifiedName asserts the server composes the worksheet
// FQN as spreadsheet FQN + "." + name without duplicating the format in the
// config.
func checkWorksheetFullyQualifiedName(resourceName, spreadsheetResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		parent, ok := state.RootModule().Resources[spreadsheetResourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", spreadsheetResourceName)
		}

		want := parent.Primary.Attributes["fully_qualified_name"] + "." + rs.Primary.Attributes["name"]
		if got := rs.Primary.Attributes["fully_qualified_name"]; got != want {
			return fmt.Errorf("worksheet fully_qualified_name = %q, want %q", got, want)
		}

		return nil
	}
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

func testAccCheckWorksheetDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_worksheet" {
			continue
		}

		_, err := sdk.Worksheets.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("worksheet %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking worksheet %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
