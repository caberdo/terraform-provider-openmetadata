package dashboard_service_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDashboardServiceDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-dashboard-service-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardServiceDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_service.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_service.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_service.by_name", "service_type", "CustomDashboard"),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard_service.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard_service.by_id", "id", "openmetadata_dashboard_service.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard_service.by_id", "fully_qualified_name", "openmetadata_dashboard_service.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDashboardServiceDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
  description  = "Looked up by name"
}

data "openmetadata_dashboard_service" "by_name" {
  fully_qualified_name = openmetadata_dashboard_service.test.fully_qualified_name
}

data "openmetadata_dashboard_service" "by_id" {
  id = openmetadata_dashboard_service.test.id
}
`, name)
}
