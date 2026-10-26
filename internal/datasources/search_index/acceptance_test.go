package search_index_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccSearchIndexDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-search-index-ds")
	serviceName := name + "-service"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSearchIndexDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_name", "service", serviceName),
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_name", "fields.#", "1"),
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_name", "fields.0.name", "message"),
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_name", "fields.0.data_type", "TEXT"),
					resource.TestCheckResourceAttrPair("data.openmetadata_search_index.by_id", "id", "openmetadata_search_index.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_search_index.by_id", "fully_qualified_name", "openmetadata_search_index.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_search_index.by_id", "fields.#", "1"),
				),
			},
		},
	})
}

func testAccSearchIndexDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_search_service" "test" {
  name         = %q
  service_type = "ElasticSearch"
}

resource "openmetadata_search_index" "test" {
  name        = %q
  service     = openmetadata_search_service.test.name
  description = "Looked up by name"

  fields = [
    {
      name      = "message"
      data_type = "TEXT"
    },
  ]
}

data "openmetadata_search_index" "by_name" {
  fully_qualified_name = openmetadata_search_index.test.fully_qualified_name
}

data "openmetadata_search_index" "by_id" {
  id = openmetadata_search_index.test.id
}
`, serviceName, name)
}
