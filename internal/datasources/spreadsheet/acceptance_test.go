package spreadsheet_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccSpreadsheetDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-spreadsheet-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSpreadsheetDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_spreadsheet.by_name", "name", name+"-sheet"),
					resource.TestCheckResourceAttr("data.openmetadata_spreadsheet.by_name", "service", name),
					resource.TestCheckResourceAttrPair("data.openmetadata_spreadsheet.by_name", "parent", "openmetadata_directory.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_spreadsheet.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_spreadsheet.by_id", "id", "openmetadata_spreadsheet.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_spreadsheet.by_id", "fully_qualified_name", "openmetadata_spreadsheet.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_spreadsheet.by_id", "parent", "openmetadata_spreadsheet.test", "parent"),
				),
			},
		},
	})
}

func testAccSpreadsheetDataSourceConfig(name string) string {
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
  description = "Looked up by name"
}

data "openmetadata_spreadsheet" "by_name" {
  fully_qualified_name = openmetadata_spreadsheet.test.fully_qualified_name
}

data "openmetadata_spreadsheet" "by_id" {
  id = openmetadata_spreadsheet.test.id
}
`, name, name, name)
}
