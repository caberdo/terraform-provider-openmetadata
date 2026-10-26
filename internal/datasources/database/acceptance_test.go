package database_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDatabaseDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-db-service-ds")
	databaseName := tfacctest.RandomWithPrefix("tf-acc-db-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig(serviceName, databaseName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_database.by_name", "name", databaseName),
					resource.TestCheckResourceAttr("data.openmetadata_database.by_name", "service", serviceName),
					resource.TestCheckResourceAttrPair("data.openmetadata_database.by_name", "fully_qualified_name", "openmetadata_database.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database.by_id", "id", "openmetadata_database.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database.by_id", "name", "openmetadata_database.test", "name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database.by_id", "service", "openmetadata_database.test", "service"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database.by_id", "fully_qualified_name", "openmetadata_database.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDatabaseDataSourceConfig(serviceName, databaseName string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
}

resource "openmetadata_database" "test" {
  name        = %q
  service     = openmetadata_database_service.test.name
  description = "Looked up by name"
}

data "openmetadata_database" "by_name" {
  fully_qualified_name = openmetadata_database.test.fully_qualified_name
}

data "openmetadata_database" "by_id" {
  id = openmetadata_database.test.id
}
`, serviceName, databaseName)
}
