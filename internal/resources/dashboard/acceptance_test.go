package dashboard_test

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

func TestAccDashboard_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-dashboard")
	resourceName := "openmetadata_dashboard.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_type", "Dashboard"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "service", "openmetadata_dashboard_service.test", "fully_qualified_name"),
				),
			},
			{
				Config: testAccDashboardConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Dashboard"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "dashboard_type", "Report"),
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

func testAccDashboardConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard" "test" {
  name           = %q
  service        = openmetadata_dashboard_service.test.fully_qualified_name
  description    = %q
  dashboard_type = "Dashboard"
}
`, name, name, description)
}

func testAccDashboardConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard" "test" {
  name           = %q
  service        = openmetadata_dashboard_service.test.fully_qualified_name
  display_name   = "Acceptance Dashboard"
  description    = %q
  dashboard_type = "Report"
  source_url     = "https://bi.example.com/report"
}
`, name, name, description)
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

func testAccCheckDashboardDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_dashboard" {
			continue
		}

		_, err := sdk.Dashboards.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("dashboard %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking dashboard %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
