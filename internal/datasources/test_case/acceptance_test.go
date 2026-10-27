package test_case_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTestCaseDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-test-case-ds")
	resourceName := "openmetadata_test_case.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTestCaseDataSourceConfig(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_test_case.by_name", "name", suffix+"-case"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_name", "fully_qualified_name", resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_test_case.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_id", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_id", "fully_qualified_name", resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_test_case.by_id", "description", "Looked up by name"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_case.by_id", "entity_link"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_case.by_id", "test_definition"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_case.by_id", "entity_fqn"),
				),
			},
		},
	})
}

func testAccTestCaseDataSourceConfig(suffix string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Acceptance database service for test case data source tests"
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

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
  ]
}

resource "openmetadata_test_definition" "test" {
  name           = %q
  entity_type    = "TABLE"
  test_platforms = ["Other"]
}

resource "openmetadata_test_case" "test" {
  name            = %q
  entity_link     = "<#E::table::${openmetadata_table.test.fully_qualified_name}>"
  test_definition = openmetadata_test_definition.test.fully_qualified_name
  description     = "Looked up by name"
}

data "openmetadata_test_case" "by_name" {
  fully_qualified_name = openmetadata_test_case.test.fully_qualified_name
}

data "openmetadata_test_case" "by_id" {
  id = openmetadata_test_case.test.id
}
`, suffix+"-service", suffix+"-db", suffix+"-schema", suffix+"-table",
		suffix+"-definition", suffix+"-case")
}
