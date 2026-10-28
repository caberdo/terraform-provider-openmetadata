package ai_application_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccAIApplicationDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-ai-application-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAIApplicationDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_ai_application.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_ai_application.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_ai_application.by_name", "application_type", "Chatbot"),
					resource.TestCheckResourceAttr("data.openmetadata_ai_application.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ai_application.by_id", "id", "openmetadata_ai_application.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ai_application.by_id", "fully_qualified_name", "openmetadata_ai_application.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccAIApplicationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_ai_application" "test" {
  name             = %q
  application_type = "Chatbot"
  description      = "Looked up by name"
}

data "openmetadata_ai_application" "by_name" {
  fully_qualified_name = openmetadata_ai_application.test.fully_qualified_name
}

data "openmetadata_ai_application" "by_id" {
  id = openmetadata_ai_application.test.id
}
`, name)
}
