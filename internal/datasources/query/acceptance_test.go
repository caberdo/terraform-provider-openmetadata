package query_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccQueryDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-query-ds-svc")
	name := tfacctest.RandomWithPrefix("tf-acc-query-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_query.by_fqn", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_query.by_fqn", "fully_qualified_name", serviceName+"."+name),
					resource.TestCheckResourceAttr("data.openmetadata_query.by_fqn", "query", "SELECT * FROM orders"),
					resource.TestCheckResourceAttr("data.openmetadata_query.by_fqn", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_query.by_id", "id", "openmetadata_query.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_query.by_id", "fully_qualified_name", "openmetadata_query.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_query.by_id", "service", "openmetadata_query.test", "service"),
				),
			},
		},
	})
}

func testAccQueryDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Service for query data source tests"
}

resource "openmetadata_query" "test" {
  service     = openmetadata_database_service.test.fully_qualified_name
  name        = %q
  query       = "SELECT * FROM orders"
  query_type  = "SELECT"
  description = "Looked up by name"
}

data "openmetadata_query" "by_fqn" {
  fully_qualified_name = openmetadata_query.test.fully_qualified_name
}

data "openmetadata_query" "by_id" {
  id = openmetadata_query.test.id
}
`, serviceName, name)
}
