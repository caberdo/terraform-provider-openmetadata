package llm_service_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccLLMServiceDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-llm-service-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccLLMServiceDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_llm_service.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_llm_service.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_llm_service.by_name", "service_type", "OpenAI"),
					resource.TestCheckResourceAttr("data.openmetadata_llm_service.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_llm_service.by_id", "id", "openmetadata_llm_service.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_llm_service.by_id", "fully_qualified_name", "openmetadata_llm_service.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccLLMServiceDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_llm_service" "test" {
  name         = %q
  service_type = "OpenAI"
  description  = "Looked up by name"
}

data "openmetadata_llm_service" "by_name" {
  fully_qualified_name = openmetadata_llm_service.test.fully_qualified_name
}

data "openmetadata_llm_service" "by_id" {
  id = openmetadata_llm_service.test.id
}
`, name)
}
