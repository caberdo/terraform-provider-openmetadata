package directory_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDirectoryDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-directory-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_directory.by_name", "name", name+"-child"),
					resource.TestCheckResourceAttr("data.openmetadata_directory.by_name", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_directory.by_name", "parent", name+"."+name+"-root"),
					resource.TestCheckResourceAttr("data.openmetadata_directory.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_directory.by_id", "id", "openmetadata_directory.child", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_directory.by_id", "fully_qualified_name", "openmetadata_directory.child", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_directory.by_id", "parent", "openmetadata_directory.root", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDirectoryDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_drive_service" "test" {
  name         = %q
  service_type = "CustomDrive"
}

resource "openmetadata_directory" "root" {
  name        = "%s-root"
  service     = openmetadata_drive_service.test.name
  description = "Root directory"
}

resource "openmetadata_directory" "child" {
  name        = "%s-child"
  service     = openmetadata_drive_service.test.name
  parent      = openmetadata_directory.root.fully_qualified_name
  description = "Looked up by name"
}

data "openmetadata_directory" "by_name" {
  fully_qualified_name = openmetadata_directory.child.fully_qualified_name
}

data "openmetadata_directory" "by_id" {
  id = openmetadata_directory.child.id
}
`, name, name, name)
}
