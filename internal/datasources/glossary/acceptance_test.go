package glossary_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccGlossaryDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-glossary-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_glossary.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_glossary.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_glossary.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_glossary.by_id", "id", "openmetadata_glossary.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_glossary.by_id", "fully_qualified_name", "openmetadata_glossary.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccGlossaryDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_glossary" "test" {
  name        = %q
  description = "Looked up by name"
}

data "openmetadata_glossary" "by_name" {
  fully_qualified_name = openmetadata_glossary.test.fully_qualified_name
}

data "openmetadata_glossary" "by_id" {
  id = openmetadata_glossary.test.id
}
`, name)
}
