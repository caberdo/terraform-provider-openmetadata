package chart_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccChartDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-chart-ds")
	serviceName := tfacctest.RandomWithPrefix("tf-acc-chart-ds-svc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccChartDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_chart.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_chart.by_name", "chart_type", "Bar"),
					resource.TestCheckResourceAttr("data.openmetadata_chart.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_chart.by_id", "id", "openmetadata_chart.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_chart.by_id", "fully_qualified_name", "openmetadata_chart.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_chart.by_id", "service", "openmetadata_chart.test", "service"),
				),
			},
		},
	})
}

func testAccChartDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_chart" "test" {
  name        = %q
  service     = openmetadata_dashboard_service.test.fully_qualified_name
  description = "Looked up by name"
  chart_type  = "Bar"
}

data "openmetadata_chart" "by_name" {
  fully_qualified_name = openmetadata_chart.test.fully_qualified_name
}

data "openmetadata_chart" "by_id" {
  id = openmetadata_chart.test.id
}
`, serviceName, name)
}
