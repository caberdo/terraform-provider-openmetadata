package ml_model_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccMlModelDataSource_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-ml-model-service-ds")
	name := tfacctest.RandomWithPrefix("tf-acc-ml-model-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccMlModelDataSourceConfig(serviceName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_ml_model.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_ml_model.by_name", "service", serviceName),
					resource.TestCheckResourceAttr("data.openmetadata_ml_model.by_name", "algorithm", "linear_regression"),
					resource.TestCheckResourceAttr("data.openmetadata_ml_model.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ml_model.by_id", "id", "openmetadata_ml_model.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ml_model.by_id", "fully_qualified_name", "openmetadata_ml_model.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccMlModelDataSourceConfig(serviceName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_ml_model_service" "test" {
  name         = %[1]q
  service_type = "Mlflow"
}

resource "openmetadata_ml_model" "test" {
  name        = %[2]q
  service     = openmetadata_ml_model_service.test.fully_qualified_name
  algorithm   = "linear_regression"
  description = "Looked up by name"
}

data "openmetadata_ml_model" "by_name" {
  fully_qualified_name = openmetadata_ml_model.test.fully_qualified_name
}

data "openmetadata_ml_model" "by_id" {
  id = openmetadata_ml_model.test.id
}
`, serviceName, name)
}
