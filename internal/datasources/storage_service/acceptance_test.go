package storage_service_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccStorageServiceDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-storage-service-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageServiceDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_storage_service.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_storage_service.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_storage_service.by_name", "service_type", "S3"),
					resource.TestCheckResourceAttr("data.openmetadata_storage_service.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_storage_service.by_id", "id", "openmetadata_storage_service.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_storage_service.by_id", "fully_qualified_name", "openmetadata_storage_service.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccStorageServiceDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_storage_service" "test" {
  name         = %q
  service_type = "S3"
  description  = "Looked up by name"
}

data "openmetadata_storage_service" "by_name" {
  fully_qualified_name = openmetadata_storage_service.test.fully_qualified_name
}

data "openmetadata_storage_service" "by_id" {
  id = openmetadata_storage_service.test.id
}
`, name)
}
