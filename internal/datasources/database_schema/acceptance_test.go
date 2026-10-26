package database_schema_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDatabaseSchemaDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-db-service-ds")
	databaseName := tfacctest.RandomWithPrefix("tf-acc-db-ds")
	schemaName := tfacctest.RandomWithPrefix("tf-acc-schema-ds")
	wantFQN := serviceName + "." + databaseName + "." + schemaName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseSchemaDataSourceConfig(serviceName, databaseName, schemaName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_database_schema.by_fqn", "name", schemaName),
					resource.TestCheckResourceAttr("data.openmetadata_database_schema.by_fqn", "fully_qualified_name", wantFQN),
					resource.TestCheckResourceAttr("data.openmetadata_database_schema.by_fqn", "database", serviceName+"."+databaseName),
					resource.TestCheckResourceAttr("data.openmetadata_database_schema.by_fqn", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database_schema.by_id", "id", "openmetadata_database_schema.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database_schema.by_id", "fully_qualified_name", "openmetadata_database_schema.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_database_schema.by_id", "database", "openmetadata_database_schema.test", "database"),
				),
			},
		},
	})
}

func testAccDatabaseSchemaDataSourceConfig(serviceName, databaseName, schemaName string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %[1]q
  service_type = "Postgres"
}

resource "openmetadata_database" "test" {
  name    = %[2]q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name        = %[3]q
  database    = openmetadata_database.test.fully_qualified_name
  description = "Looked up by name"
}

data "openmetadata_database_schema" "by_fqn" {
  fully_qualified_name = openmetadata_database_schema.test.fully_qualified_name
}

data "openmetadata_database_schema" "by_id" {
  id = openmetadata_database_schema.test.id
}
`, serviceName, databaseName, schemaName)
}
