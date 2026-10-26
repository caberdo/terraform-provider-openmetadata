package metric_test

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

func TestAccMetric_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-metric")
	resourceName := "openmetadata_metric.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMetricDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMetricConfig(name, "Initial description", "SUM", "DOLLARS", "WEEK", "sum(revenue)", "SQL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "metric_type", "SUM"),
					resource.TestCheckResourceAttr(resourceName, "unit_of_measurement", "DOLLARS"),
					resource.TestCheckResourceAttr(resourceName, "granularity", "WEEK"),
					resource.TestCheckResourceAttr(resourceName, "expression_code", "sum(revenue)"),
					resource.TestCheckResourceAttr(resourceName, "expression_language", "SQL"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccMetricConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Metric"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "metric_type", "AVERAGE"),
					resource.TestCheckResourceAttr(resourceName, "granularity", "DAY"),
					resource.TestCheckResourceAttr(resourceName, "expression_code", "avg(revenue)"),
					resource.TestCheckResourceAttr(resourceName, "expression_language", "Python"),
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

func testAccMetricConfig(name, description, metricType, unit, granularity, code, language string) string {
	return fmt.Sprintf(`
resource "openmetadata_metric" "test" {
  name                  = %q
  description           = %q
  metric_type           = %q
  unit_of_measurement   = %q
  granularity           = %q
  expression_code       = %q
  expression_language   = %q
}
`, name, description, metricType, unit, granularity, code, language)
}

func testAccMetricConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_metric" "test" {
  name                = %q
  display_name        = "Acceptance Metric"
  description         = %q
  metric_type         = "AVERAGE"
  unit_of_measurement = "DOLLARS"
  granularity         = "DAY"
  expression_code     = "avg(revenue)"
  expression_language = "Python"
}
`, name, description)
}

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

func testAccCheckMetricDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_metric" {
			continue
		}

		_, err := sdk.Metrics.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("metric %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking metric %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
