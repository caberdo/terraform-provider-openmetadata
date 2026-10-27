package data_contract_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDataContractDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-data-contract-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataContractDataSourceConfig(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.openmetadata_data_contract.by_id", "id", "openmetadata_data_contract.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_data_contract.by_id", "fully_qualified_name", "openmetadata_data_contract.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_data_contract.by_name", "id", "openmetadata_data_contract.test", "id"),
					resource.TestCheckResourceAttr("data.openmetadata_data_contract.by_name", "description", "Looked up by fully qualified name"),
					resource.TestCheckResourceAttr("data.openmetadata_data_contract.by_name", "terms_of_use", "Lookup terms"),
					resource.TestCheckResourceAttrPair("data.openmetadata_data_contract.by_name", "entity", "openmetadata_table.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDataContractDataSourceConfig(suffix string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Acceptance database service for data contract data source tests"
}

resource "openmetadata_database" "test" {
  name    = %q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name     = %q
  database = openmetadata_database.test.fully_qualified_name
}

resource "openmetadata_table" "test" {
  name            = %q
  database_schema = openmetadata_database_schema.test.fully_qualified_name

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
  ]
}

resource "openmetadata_data_contract" "test" {
  name         = %q
  entity       = openmetadata_table.test.fully_qualified_name
  description  = "Looked up by fully qualified name"
  terms_of_use = "Lookup terms"
}

data "openmetadata_data_contract" "by_name" {
  fully_qualified_name = openmetadata_data_contract.test.fully_qualified_name
}

data "openmetadata_data_contract" "by_id" {
  id = openmetadata_data_contract.test.id
}
`, suffix+"-service", suffix+"-db", suffix+"-schema", suffix+"-table", suffix+"-contract")
}
