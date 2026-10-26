package pipeline_test

import (
	"context"
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
)

func TestAccPipeline_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-pipeline")
	resourceName := "openmetadata_pipeline.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig(suffix, "Initial description", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suffix+"-pipeline"),
					resource.TestCheckResourceAttrPair(resourceName, "service", "openmetadata_pipeline_service.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "concurrency", "2"),
					resource.TestCheckResourceAttr(resourceName, "schedule_interval", "0 2 * * *"),
					resource.TestCheckResourceAttr(resourceName, "state", "Active"),
					resource.TestCheckResourceAttr(resourceName, "tasks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.name", "extract_orders"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.task_type", "Extract"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccPipelineConfig(suffix, "Updated description", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "tasks.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.name", "load_orders_mart"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.task_type", "Load"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDByFullyQualifiedName(resourceName),
			},
		},
	})
}

func testAccPipelineConfig(suffix, description string, withSecondTask bool) string {
	tasks := `
    {
      name        = "extract_orders"
      task_type   = "Extract"
      task_sql    = "select * from raw.orders"
      description = "Reads new orders from the landing zone."
    },`
	if withSecondTask {
		tasks += `
    {
      name      = "load_orders_mart"
      task_type = "Load"
      task_sql  = "insert into mart.orders select * from staging.orders"
    },`
	}

	return fmt.Sprintf(`
resource "openmetadata_pipeline_service" "test" {
  name         = %q
  service_type = "CustomPipeline"
}

resource "openmetadata_pipeline" "test" {
  name              = %q
  service           = openmetadata_pipeline_service.test.fully_qualified_name
  description       = %q
  concurrency       = 2
  schedule_interval = "0 2 * * *"

  tasks = [%s
  ]
}
`, suffix+"-service", suffix+"-pipeline", description, tasks)
}

// importStateIDByFullyQualifiedName exercises the name-based import path rather
// than the UUID path used by the default import step.
func importStateIDByFullyQualifiedName(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		fqn := rs.Primary.Attributes["fully_qualified_name"]
		if fqn == "" {
			return "", fmt.Errorf("resource %s has no fully_qualified_name", resourceName)
		}
		return fqn, nil
	}
}

func testAccCheckPipelineDestroy(state *terraform.State) error {
	openmetadataClient, err := acctest.NewClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	sdk, err := openmetadataClient.SDK(ctx)
	if err != nil {
		return err
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "openmetadata_pipeline" {
			continue
		}

		_, err := sdk.Pipelines.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("pipeline %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking pipeline %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
