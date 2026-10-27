package test_suite_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTestSuiteDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-test-suite-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTestSuiteDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_test_suite.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_test_suite.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_test_suite.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_suite.by_name", "basic"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_suite.by_name", "executable"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_suite.by_id", "id", "openmetadata_test_suite.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_suite.by_id", "fully_qualified_name", "openmetadata_test_suite.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccTestSuiteDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_test_suite" "test" {
  name        = %q
  description = "Looked up by name"
}

data "openmetadata_test_suite" "by_name" {
  fully_qualified_name = openmetadata_test_suite.test.fully_qualified_name
}

data "openmetadata_test_suite" "by_id" {
  id = openmetadata_test_suite.test.id
}
`, name)
}
