package ingestion_pipeline_test

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

func TestAccIngestionPipeline_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-ingestion-pipeline")
	serviceName := tfacctest.RandomWithPrefix("tf-acc-ingestion-pipeline-svc")
	resourceName := "openmetadata_ingestion_pipeline.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckIngestionPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionPipelineConfig(serviceName, name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrPair(resourceName, "service", "openmetadata_database_service.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "service_type", "databaseService"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", "metadata"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "enable_streamable_logs", "false"),
					resource.TestCheckResourceAttr(resourceName, "raise_on_error", "false"),
				),
			},
			{
				Config: testAccIngestionPipelineConfig(serviceName, name, `
  display_name = "Acceptance Ingestion Pipeline"
  description  = "Updated by the acceptance test."`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Ingestion Pipeline"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated by the acceptance test."),
					resource.TestCheckResourceAttr(resourceName, "pipeline_type", "metadata"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"airflow_config", "source_config"},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"airflow_config", "source_config"},
				ImportStateIdFunc:       importStateIDByFullyQualifiedName(resourceName),
			},
		},
	})
}

func testAccIngestionPipelineConfig(serviceName, name, extra string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %[1]q
  service_type = "Postgres"
  description  = "Service created for the ingestion pipeline acceptance test."
}

resource "openmetadata_ingestion_pipeline" "test" {
  name          = %[2]q
  service       = openmetadata_database_service.test.fully_qualified_name
  pipeline_type = "metadata"
  source_config  = jsonencode({ type = "DatabaseMetadata" })
  airflow_config = jsonencode({ scheduleInterval = "0 * * * *" })
%[3]s
}
`, serviceName, name, extra)
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

func testAccCheckIngestionPipelineDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_ingestion_pipeline" {
			continue
		}

		_, err := sdk.IngestionPipelines.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("ingestion pipeline %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking ingestion pipeline %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
