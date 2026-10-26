package dashboard_data_model_test

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

func TestAccDashboardDataModel_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-data-model-service")
	name := tfacctest.RandomWithPrefix("tf-acc-dashboard-data-model")
	resourceName := "openmetadata_dashboard_data_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDashboardDataModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardDataModelConfig(serviceName, name, "Initial description", "SELECT 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+".model."+name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "data_model_type", "PowerBIDataModel"),
					resource.TestCheckResourceAttrPair(resourceName, "service", "openmetadata_dashboard_service.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "sql", "SELECT 1"),
					resource.TestCheckResourceAttr(resourceName, "columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "columns.0.name", "region"),
					resource.TestCheckResourceAttr(resourceName, "columns.0.data_type", "VARCHAR"),
					resource.TestCheckResourceAttr(resourceName, "columns.1.name", "total_sales"),
					resource.TestCheckResourceAttr(resourceName, "columns.1.data_type", "DOUBLE"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccDashboardDataModelConfigUpdated(serviceName, name, "Updated description", "SELECT 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Data Model"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "sql", "SELECT 2"),
					resource.TestCheckResourceAttr(resourceName, "columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "columns.0.description", "Updated region description."),
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

func testAccDashboardDataModelConfig(serviceName, name, description, sql string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard_data_model" "test" {
  name            = %q
  service         = openmetadata_dashboard_service.test.fully_qualified_name
  data_model_type = "PowerBIDataModel"
  description     = %q
  sql             = %q

  columns = [
    {
      name        = "region"
      data_type   = "VARCHAR"
      description = "Sales region."
    },
    {
      name      = "total_sales"
      data_type = "DOUBLE"
    },
  ]
}
`, serviceName, name, description, sql)
}

func testAccDashboardDataModelConfigUpdated(serviceName, name, description, sql string) string {
	return fmt.Sprintf(`
resource "openmetadata_dashboard_service" "test" {
  name         = %q
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard_data_model" "test" {
  name            = %q
  display_name    = "Acceptance Data Model"
  service         = openmetadata_dashboard_service.test.fully_qualified_name
  data_model_type = "PowerBIDataModel"
  description     = %q
  sql             = %q

  columns = [
    {
      name        = "region"
      data_type   = "VARCHAR"
      description = "Updated region description."
    },
    {
      name      = "total_sales"
      data_type = "DOUBLE"
    },
  ]
}
`, serviceName, name, description, sql)
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

func testAccCheckDashboardDataModelDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_dashboard_data_model" {
			continue
		}

		_, err := sdk.DashboardDataModels.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("dashboard data model %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking dashboard data model %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
