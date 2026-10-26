package stored_procedure_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccStoredProcedureDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-sp-ds-service")
	databaseName := tfacctest.RandomWithPrefix("tf-acc-sp-ds-database")
	schemaName := tfacctest.RandomWithPrefix("tf-acc-sp-ds-schema")
	procedureName := tfacctest.RandomWithPrefix("tf-acc-sp-ds-procedure")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccStoredProcedureDataSourceConfig(serviceName, databaseName, schemaName, procedureName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_stored_procedure.by_fqn", "name", procedureName),
					resource.TestCheckResourceAttr("data.openmetadata_stored_procedure.by_fqn", "code", "BEGIN SELECT 1; END"),
					resource.TestCheckResourceAttr("data.openmetadata_stored_procedure.by_fqn", "language", "SQL"),
					resource.TestCheckResourceAttrPair("data.openmetadata_stored_procedure.by_fqn", "fully_qualified_name", "openmetadata_stored_procedure.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_stored_procedure.by_id", "id", "openmetadata_stored_procedure.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_stored_procedure.by_id", "fully_qualified_name", "openmetadata_stored_procedure.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_stored_procedure.by_id", "database_schema", "openmetadata_stored_procedure.test", "database_schema"),
				),
			},
		},
	})
}

func testAccStoredProcedureDataSourceConfig(serviceName, databaseName, schemaName, procedureName string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
}

resource "openmetadata_database" "test" {
  name    = %q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name     = %q
  database = openmetadata_database.test.fully_qualified_name
}

resource "openmetadata_stored_procedure" "test" {
  name            = %q
  database_schema = openmetadata_database_schema.test.fully_qualified_name
  code            = "BEGIN SELECT 1; END"
  language        = "SQL"
}

data "openmetadata_stored_procedure" "by_fqn" {
  fully_qualified_name = openmetadata_stored_procedure.test.fully_qualified_name
}

data "openmetadata_stored_procedure" "by_id" {
  id = openmetadata_stored_procedure.test.id
}
`, serviceName, databaseName, schemaName, procedureName)
}
