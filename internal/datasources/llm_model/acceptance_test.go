package llm_model_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccLlmModelDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-llm-service-ds")
	name := tfacctest.RandomWithPrefix("tf-acc-llm-model-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccLlmModelDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_llm_model.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_llm_model.by_name", "service", serviceName),
					resource.TestCheckResourceAttr("data.openmetadata_llm_model.by_name", "base_model", "gpt-4o"),
					resource.TestCheckResourceAttr("data.openmetadata_llm_model.by_name", "model_provider", "openai"),
					resource.TestCheckResourceAttr("data.openmetadata_llm_model.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_llm_model.by_id", "id", "openmetadata_llm_model.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_llm_model.by_id", "fully_qualified_name", "openmetadata_llm_model.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccLlmModelDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_llm_service" "test" {
  name         = %[1]q
  service_type = "OpenAI"
}

resource "openmetadata_llm_model" "test" {
  name           = %[2]q
  service        = openmetadata_llm_service.test.fully_qualified_name
  base_model     = "gpt-4o"
  model_provider = "openai"
  description    = "Looked up by name"
}

data "openmetadata_llm_model" "by_name" {
  fully_qualified_name = openmetadata_llm_model.test.fully_qualified_name
}

data "openmetadata_llm_model" "by_id" {
  id = openmetadata_llm_model.test.id
}
`, serviceName, name)
}
