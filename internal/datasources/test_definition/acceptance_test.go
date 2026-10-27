package test_definition_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTestDefinitionDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-test-definition-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTestDefinitionDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_test_definition.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_test_definition.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_test_definition.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_test_definition.by_name", "entity_type", "TABLE"),
					resource.TestCheckResourceAttr("data.openmetadata_test_definition.by_name", "test_platforms.#", "1"),
					resource.TestCheckResourceAttr("data.openmetadata_test_definition.by_name", "test_platforms.0", "Other"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_definition.by_id", "id", "openmetadata_test_definition.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_definition.by_id", "fully_qualified_name", "openmetadata_test_definition.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccTestDefinitionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_test_definition" "test" {
  name           = %q
  description    = "Looked up by name"
  entity_type    = "TABLE"
  test_platforms = ["Other"]
}

data "openmetadata_test_definition" "by_name" {
  fully_qualified_name = openmetadata_test_definition.test.fully_qualified_name
}

data "openmetadata_test_definition" "by_id" {
  id = openmetadata_test_definition.test.id
}
`, name)
}
