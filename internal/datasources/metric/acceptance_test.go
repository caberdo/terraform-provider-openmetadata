package metric_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccMetricDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-metric-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMetricDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_metric.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_metric.by_name", "metric_type", "SUM"),
					resource.TestCheckResourceAttr("data.openmetadata_metric.by_name", "unit_of_measurement", "DOLLARS"),
					resource.TestCheckResourceAttrPair("data.openmetadata_metric.by_id", "id", "openmetadata_metric.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_metric.by_id", "fully_qualified_name", "openmetadata_metric.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccMetricDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_metric" "test" {
  name                = %q
  description         = "Looked up by name"
  metric_type         = "SUM"
  unit_of_measurement = "DOLLARS"
}

data "openmetadata_metric" "by_name" {
  fully_qualified_name = openmetadata_metric.test.fully_qualified_name
}

data "openmetadata_metric" "by_id" {
  id = openmetadata_metric.test.id
}
`, name)
}
