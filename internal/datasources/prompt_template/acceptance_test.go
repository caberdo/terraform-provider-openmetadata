package prompt_template_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccPromptTemplateDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-prompt-template-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptTemplateDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_prompt_template.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_prompt_template.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_prompt_template.by_name", "template_type", "Classification"),
					resource.TestCheckResourceAttrPair("data.openmetadata_prompt_template.by_id", "id", "openmetadata_prompt_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_prompt_template.by_id", "fully_qualified_name", "openmetadata_prompt_template.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccPromptTemplateDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_prompt_template" "test" {
  name             = %q
  template_content = "Classify the ticket: {{ticket}}"
  template_type    = "Classification"
}

data "openmetadata_prompt_template" "by_name" {
  fully_qualified_name = openmetadata_prompt_template.test.fully_qualified_name
}

data "openmetadata_prompt_template" "by_id" {
  id = openmetadata_prompt_template.test.id
}
`, name)
}
