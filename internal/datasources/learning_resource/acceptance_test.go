package learning_resource_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccLearningResourceDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-learning-resource-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccLearningResourceDataSourceConfig(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "name", suffix+"-resource"),
					resource.TestCheckResourceAttrPair("data.openmetadata_learning_resource.by_name", "fully_qualified_name", "openmetadata_learning_resource.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "resource_type", "Article"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "categories.#", "1"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "categories.0", "AI"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "contexts.0.page_id", "home"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "source.#", "1"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_name", "source.0.url", "https://example.com/academy"),
					resource.TestCheckResourceAttrPair("data.openmetadata_learning_resource.by_id", "id", "openmetadata_learning_resource.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_learning_resource.by_id", "fully_qualified_name", "openmetadata_learning_resource.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_learning_resource.by_id", "categories.0", "AI"),
				),
			},
		},
	})
}

func testAccLearningResourceDataSourceConfig(suffix string) string {
	return fmt.Sprintf(`
resource "openmetadata_learning_resource" "test" {
  name          = %q
  resource_type = "Article"
  description   = "Looked up by name and id"
  categories    = ["AI"]
  difficulty    = "Intro"
  status        = "Active"

  contexts = [
    {
      page_id  = "home"
      priority = 1
    },
  ]

  source = [
    {
      url = "https://example.com/academy"
    },
  ]
}

data "openmetadata_learning_resource" "by_name" {
  fully_qualified_name = openmetadata_learning_resource.test.fully_qualified_name
}

data "openmetadata_learning_resource" "by_id" {
  id = openmetadata_learning_resource.test.id
}
`, suffix+"-resource")
}
