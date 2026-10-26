package table_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTableDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-table-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_table.by_name", "name", suffix+"-table"),
					resource.TestCheckResourceAttrPair("data.openmetadata_table.by_name", "fully_qualified_name", "openmetadata_table.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_table.by_name", "columns.#", "2"),
					resource.TestCheckResourceAttr("data.openmetadata_table.by_name", "columns.0.name", "id"),
					resource.TestCheckResourceAttr("data.openmetadata_table.by_name", "columns.1.data_type", "VARCHAR"),
					resource.TestCheckResourceAttrPair("data.openmetadata_table.by_id", "id", "openmetadata_table.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_table.by_id", "fully_qualified_name", "openmetadata_table.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_table.by_id", "columns.#", "2"),
				),
			},
		},
	})
}

func testAccTableDataSourceConfig(suffix string) string {
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

resource "openmetadata_table" "test" {
  name            = %q
  database_schema = openmetadata_database_schema.test.fully_qualified_name
  description     = "Looked up by name and id"

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
    {
      name             = "email"
      data_type        = "VARCHAR"
      precision        = 255
      ordinal_position = 2
    },
  ]
}

data "openmetadata_table" "by_name" {
  fully_qualified_name = openmetadata_table.test.fully_qualified_name
}

data "openmetadata_table" "by_id" {
  id = openmetadata_table.test.id
}
`, suffix+"-service", suffix+"-db", suffix+"-schema", suffix+"-table")
}
