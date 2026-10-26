package tag_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTagDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-tag-ds")
	classificationName := tfacctest.RandomWithPrefix("tf-acc-tag-ds-classification")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTagDataSourceConfig(classificationName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_tag.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_tag.by_name", "classification", classificationName),
					resource.TestCheckResourceAttr("data.openmetadata_tag.by_name", "fully_qualified_name", classificationName+"."+name),
					resource.TestCheckResourceAttrPair("data.openmetadata_tag.by_id", "id", "openmetadata_tag.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_tag.by_id", "fully_qualified_name", "openmetadata_tag.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccTagDataSourceConfig(classificationName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_classification" "parent" {
  name = %q
}

resource "openmetadata_tag" "test" {
  classification = openmetadata_classification.parent.name
  name           = %q
  description    = "Looked up by name"
}

data "openmetadata_tag" "by_name" {
  fully_qualified_name = openmetadata_tag.test.fully_qualified_name
}

data "openmetadata_tag" "by_id" {
  id = openmetadata_tag.test.id
}
`, classificationName, name)
}
