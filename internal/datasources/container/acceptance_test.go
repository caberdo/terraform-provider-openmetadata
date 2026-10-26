package container_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccContainerDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-container-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_container.by_name", "name", name+"-child"),
					resource.TestCheckResourceAttr("data.openmetadata_container.by_name", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_container.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_container.by_id", "id", "openmetadata_container.child", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_container.by_id", "fully_qualified_name", "openmetadata_container.child", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccContainerDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_storage_service" "test" {
  name         = %q
  service_type = "S3"
}

resource "openmetadata_container" "child" {
  name        = "%s-child"
  service     = openmetadata_storage_service.test.name
  description = "Looked up by name"
}

data "openmetadata_container" "by_name" {
  fully_qualified_name = openmetadata_container.child.fully_qualified_name
}

data "openmetadata_container" "by_id" {
  id = openmetadata_container.child.id
}
`, name, name)
}
