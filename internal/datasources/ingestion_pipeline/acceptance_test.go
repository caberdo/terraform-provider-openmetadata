package ingestion_pipeline_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccIngestionPipelineDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-ingestion-pipeline-ds-svc")
	name := tfacctest.RandomWithPrefix("tf-acc-ingestion-pipeline-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionPipelineDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.openmetadata_ingestion_pipeline.by_name", "id", "openmetadata_ingestion_pipeline.test", "id"),
					resource.TestCheckResourceAttr("data.openmetadata_ingestion_pipeline.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_ingestion_pipeline.by_name", "pipeline_type", "metadata"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ingestion_pipeline.by_id", "id", "openmetadata_ingestion_pipeline.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ingestion_pipeline.by_id", "fully_qualified_name", "openmetadata_ingestion_pipeline.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ingestion_pipeline.by_id", "service", "openmetadata_database_service.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccIngestionPipelineDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %[1]q
  service_type = "Postgres"
}

resource "openmetadata_ingestion_pipeline" "test" {
  name           = %[2]q
  service        = openmetadata_database_service.test.fully_qualified_name
  pipeline_type  = "metadata"
  source_config  = jsonencode({ type = "DatabaseMetadata" })
  airflow_config = jsonencode({ scheduleInterval = "0 * * * *" })
}

data "openmetadata_ingestion_pipeline" "by_name" {
  fully_qualified_name = openmetadata_ingestion_pipeline.test.fully_qualified_name
}

data "openmetadata_ingestion_pipeline" "by_id" {
  id = openmetadata_ingestion_pipeline.test.id
}
`, serviceName, name)
}
