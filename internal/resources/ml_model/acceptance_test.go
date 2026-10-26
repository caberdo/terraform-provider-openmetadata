package ml_model_test

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

func TestAccMlModel_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-ml-model-service")
	name := tfacctest.RandomWithPrefix("tf-acc-ml-model")
	resourceName := "openmetadata_ml_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckMlModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMlModelConfig(serviceName, name, "linear_regression", "", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
					resource.TestCheckResourceAttr(resourceName, "algorithm", "linear_regression"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+"."+name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccMlModelConfig(serviceName, name, "logistic_regression", "Acceptance ML Model", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "algorithm", "logistic_regression"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance ML Model"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
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

func testAccMlModelConfig(serviceName, name, algorithm, displayName, description string) string {
	displayLine := ""
	if displayName != "" {
		displayLine = fmt.Sprintf("  display_name = %q\n", displayName)
	}
	return fmt.Sprintf(`
resource "openmetadata_ml_model_service" "test" {
  name         = %q
  service_type = "Mlflow"
}

resource "openmetadata_ml_model" "test" {
  name        = %q
  service     = openmetadata_ml_model_service.test.fully_qualified_name
  algorithm   = %q
%[4]s  description = %[5]q
}
`, serviceName, name, algorithm, displayLine, description)
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

func testAccCheckMlModelDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_ml_model" {
			continue
		}

		_, err := sdk.MlModels.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("ML model %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking ML model %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
