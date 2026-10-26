package classification_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccClassificationDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-classification-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccClassificationDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_classification.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_classification.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_classification.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_classification.by_id", "id", "openmetadata_classification.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_classification.by_id", "fully_qualified_name", "openmetadata_classification.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccClassificationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_classification" "test" {
  name        = %q
  description = "Looked up by name"
}

data "openmetadata_classification" "by_name" {
  fully_qualified_name = openmetadata_classification.test.fully_qualified_name
}

data "openmetadata_classification" "by_id" {
  id = openmetadata_classification.test.id
}
`, name)
}
