package api_collection_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccApiCollectionDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-api-collection-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccApiCollectionDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_api_collection.by_name", "name", name+"-pets"),
					resource.TestCheckResourceAttr("data.openmetadata_api_collection.by_name", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_api_collection.by_name", "service_type", "Rest"),
					resource.TestCheckResourceAttr("data.openmetadata_api_collection.by_name", "endpoint_url", "https://api.example.com/v1"),
					resource.TestCheckResourceAttr("data.openmetadata_api_collection.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_api_collection.by_id", "id", "openmetadata_api_collection.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_api_collection.by_id", "fully_qualified_name", "openmetadata_api_collection.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccApiCollectionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_api_service" "test" {
  name         = %q
  service_type = "Rest"
}

resource "openmetadata_api_collection" "test" {
  name         = "%s-pets"
  service      = openmetadata_api_service.test.name
  endpoint_url = "https://api.example.com/v1"
  description  = "Looked up by name"
}

data "openmetadata_api_collection" "by_name" {
  fully_qualified_name = openmetadata_api_collection.test.fully_qualified_name
}

data "openmetadata_api_collection" "by_id" {
  id = openmetadata_api_collection.test.id
}
`, name, name)
}
