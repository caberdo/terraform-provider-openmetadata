package file_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccFileDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-file-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccFileDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_file.by_name", "name", name+"-file"),
					resource.TestCheckResourceAttr("data.openmetadata_file.by_name", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_file.by_name", "directory", name+"."+name+"-dir"),
					resource.TestCheckResourceAttr("data.openmetadata_file.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_file.by_id", "id", "openmetadata_file.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_file.by_id", "fully_qualified_name", "openmetadata_file.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_file.by_id", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_file.by_id", "directory", name+"."+name+"-dir"),
				),
			},
		},
	})
}

func testAccFileDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "test" {
  name    = "%s-dir"
  service = openmetadata_drive_service.test.name
}

resource "openmetadata_file" "test" {
  name        = "%s-file"
  service     = openmetadata_drive_service.test.name
  directory   = openmetadata_directory.test.fully_qualified_name
  description = "Looked up by name"
}

data "openmetadata_file" "by_name" {
  fully_qualified_name = openmetadata_file.test.fully_qualified_name
}

data "openmetadata_file" "by_id" {
  id = openmetadata_file.test.id
}
`, name, name, name)
}
