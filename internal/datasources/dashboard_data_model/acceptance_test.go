package dashboard_data_model_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDashboardDataModelDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-data-model-ds-service")
	name := tfacctest.RandomWithPrefix("tf-acc-dashboard-data-model-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardDataModelDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_data_model.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_data_model.by_name", "data_model_type", "PowerBIDataModel"),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_data_model.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_data_model.by_name", "sql", "SELECT 1"),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_data_model.by_name", "columns.#", "2"),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_data_model.by_name", "columns.0.name", "region"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard_data_model.by_id", "id", "openmetadata_dashboard_data_model.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard_data_model.by_id", "fully_qualified_name", "openmetadata_dashboard_data_model.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard_data_model.by_id", "service", "openmetadata_dashboard_service.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDashboardDataModelDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard_data_model" "test" {
  name            = %q
  service         = openmetadata_dashboard_service.test.fully_qualified_name
  data_model_type = "PowerBIDataModel"
  description     = "Looked up by name"
  sql             = "SELECT 1"

  columns = [
    {
      name        = "region"
      data_type   = "VARCHAR"
      data_length = 256
    },
    {
      name      = "total_sales"
      data_type = "DOUBLE"
    },
  ]
}

data "openmetadata_dashboard_data_model" "by_name" {
  fully_qualified_name = openmetadata_dashboard_data_model.test.fully_qualified_name
}

data "openmetadata_dashboard_data_model" "by_id" {
  id = openmetadata_dashboard_data_model.test.id
}
`, serviceName, name)
}
