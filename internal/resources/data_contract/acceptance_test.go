package data_contract_test

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

func TestAccDataContract_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-data-contract")
	resourceName := "openmetadata_data_contract.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDataContractDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataContractConfig(suffix, "Initial contract description", "Initial terms of use"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suffix+"-contract"),
					resource.TestCheckResourceAttrPair(resourceName, "entity", "openmetadata_table.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial contract description"),
					resource.TestCheckResourceAttr(resourceName, "terms_of_use", "Initial terms of use"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccDataContractConfig(suffix, "Updated contract description", "Updated terms of use"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated contract description"),
					resource.TestCheckResourceAttr(resourceName, "terms_of_use", "Updated terms of use"),
					resource.TestCheckResourceAttrPair(resourceName, "entity", "openmetadata_table.test", "fully_qualified_name"),
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

func testAccDataContractConfig(suffix, description, terms string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Acceptance database service for data contract tests"
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
  description     = "Table governed by the data contract"

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
    {
      name             = "email"
      data_type        = "VARCHAR"
      data_length      = 255
      ordinal_position = 2
    },
  ]
}

resource "openmetadata_data_contract" "test" {
  name         = %q
  entity       = openmetadata_table.test.fully_qualified_name
  description  = %q
  terms_of_use = %q
}
`, suffix+"-service", suffix+"-db", suffix+"-schema", suffix+"-table", suffix+"-contract", description, terms)
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

func testAccCheckDataContractDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_data_contract" {
			continue
		}

		_, err := sdk.DataContracts.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("data contract %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking data contract %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
