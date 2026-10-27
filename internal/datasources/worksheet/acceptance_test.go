package worksheet_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccWorksheetDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-worksheet-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWorksheetDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_worksheet.by_name", "name", name+"-q1"),
					resource.TestCheckResourceAttrPair("data.openmetadata_worksheet.by_name", "spreadsheet", "openmetadata_spreadsheet.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_worksheet.by_name", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_worksheet.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_worksheet.by_id", "id", "openmetadata_worksheet.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_worksheet.by_id", "fully_qualified_name", "openmetadata_worksheet.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_worksheet.by_id", "spreadsheet", "openmetadata_spreadsheet.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_worksheet.by_id", "service", name),
				),
			},
		},
	})
}

func testAccWorksheetDataSourceConfig(name string) string {
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
  description = "Looked up by name"
}

data "openmetadata_worksheet" "by_name" {
  fully_qualified_name = openmetadata_worksheet.test.fully_qualified_name
}

data "openmetadata_worksheet" "by_id" {
  id = openmetadata_worksheet.test.id
}
`, name, name, name)
}
