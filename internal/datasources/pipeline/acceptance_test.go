package pipeline_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccPipelineDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-pipeline-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDataSourceConfig(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_pipeline.by_name", "name", suffix+"-pipeline"),
					resource.TestCheckResourceAttrPair("data.openmetadata_pipeline.by_name", "fully_qualified_name", "openmetadata_pipeline.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_pipeline.by_name", "tasks.#", "2"),
					resource.TestCheckResourceAttr("data.openmetadata_pipeline.by_name", "tasks.0.name", "extract_orders"),
					resource.TestCheckResourceAttrPair("data.openmetadata_pipeline.by_id", "id", "openmetadata_pipeline.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_pipeline.by_id", "fully_qualified_name", "openmetadata_pipeline.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_pipeline.by_id", "tasks.#", "2"),
				),
			},
		},
	})
}

func testAccPipelineDataSourceConfig(suffix string) string {
	return fmt.Sprintf(`
resource "openmetadata_pipeline_service" "test" {
  name         = %q
  service_type = "CustomPipeline"
}

resource "openmetadata_pipeline" "test" {
  name        = %q
  service     = openmetadata_pipeline_service.test.fully_qualified_name
  description = "Looked up by name and id"

  tasks = [
    {
      name      = "extract_orders"
      task_type = "Extract"
    },
    {
      name      = "load_orders_mart"
      task_type = "Load"
    },
  ]
}

data "openmetadata_pipeline" "by_name" {
  fully_qualified_name = openmetadata_pipeline.test.fully_qualified_name
}

data "openmetadata_pipeline" "by_id" {
  id = openmetadata_pipeline.test.id
}
`, suffix+"-service", suffix+"-pipeline")
}
