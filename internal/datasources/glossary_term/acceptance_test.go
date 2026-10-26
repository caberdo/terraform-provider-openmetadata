package glossary_term_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccGlossaryTermDataSource_basic(t *testing.T) {
	glossaryName := tfacctest.RandomWithPrefix("tf-acc-term-ds-glossary")
	name := tfacctest.RandomWithPrefix("tf-acc-term-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryTermDataSourceConfig(glossaryName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_glossary_term.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_glossary_term.by_name", "glossary", glossaryName),
					resource.TestCheckResourceAttr("data.openmetadata_glossary_term.by_name", "fully_qualified_name", glossaryName+"."+name),
					resource.TestCheckResourceAttrPair("data.openmetadata_glossary_term.by_id", "id", "openmetadata_glossary_term.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_glossary_term.by_id", "fully_qualified_name", "openmetadata_glossary_term.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccGlossaryTermDataSourceConfig(glossaryName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_glossary" "parent" {
  name = %q
}

resource "openmetadata_glossary_term" "test" {
  glossary    = openmetadata_glossary.parent.name
  name        = %q
  description = "Looked up by name"
}

data "openmetadata_glossary_term" "by_name" {
  fully_qualified_name = openmetadata_glossary_term.test.fully_qualified_name
}

data "openmetadata_glossary_term" "by_id" {
  id = openmetadata_glossary_term.test.id
}
`, glossaryName, name)
}
