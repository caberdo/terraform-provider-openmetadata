package api_endpoint_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccApiEndpointDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-api-endpoint-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccApiEndpointDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "name", name+"-list"),
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "service", name),
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "service_type", "Rest"),
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "api_collection", name+"."+name+"-pets"),
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "request_method", "GET"),
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "endpoint_url", "https://api.example.com/v1/pets"),
					resource.TestCheckResourceAttr("data.openmetadata_api_endpoint.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_api_endpoint.by_id", "id", "openmetadata_api_endpoint.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_api_endpoint.by_id", "fully_qualified_name", "openmetadata_api_endpoint.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccApiEndpointDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_api_service" "test" {
  name         = %q
  service_type = "Rest"
}

resource "openmetadata_api_collection" "test" {
  name    = "%s-pets"
  service = openmetadata_api_service.test.name
}

resource "openmetadata_api_endpoint" "test" {
  name           = "%s-list"
  api_collection = openmetadata_api_collection.test.fully_qualified_name
  request_method = "GET"
  endpoint_url   = "https://api.example.com/v1/pets"
  description    = "Looked up by name"
}

data "openmetadata_api_endpoint" "by_name" {
  fully_qualified_name = openmetadata_api_endpoint.test.fully_qualified_name
}

data "openmetadata_api_endpoint" "by_id" {
  id = openmetadata_api_endpoint.test.id
}
`, name, name, name)
}
