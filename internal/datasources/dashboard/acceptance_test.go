package dashboard_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDashboardDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-dashboard-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_dashboard.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_dashboard.by_name", "dashboard_type", "Dashboard"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard.by_name", "fully_qualified_name", "openmetadata_dashboard.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard.by_name", "service", "openmetadata_dashboard.test", "service"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard.by_id", "id", "openmetadata_dashboard.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_dashboard.by_id", "fully_qualified_name", "openmetadata_dashboard.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDashboardDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard" "test" {
  name        = %q
  service     = openmetadata_dashboard_service.test.fully_qualified_name
  description = "Looked up by name"
}

data "openmetadata_dashboard" "by_name" {
  fully_qualified_name = openmetadata_dashboard.test.fully_qualified_name
}

data "openmetadata_dashboard" "by_id" {
  id = openmetadata_dashboard.test.id
}
`, name, name)
}
