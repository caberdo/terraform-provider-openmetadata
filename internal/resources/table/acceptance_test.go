package table_test

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

func TestAccTable_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-table")
	resourceName := "openmetadata_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig(suffix, "Initial table description", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suffix+"-table"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial table description"),
					resource.TestCheckResourceAttr(resourceName, "columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "columns.0.name", "id"),
					resource.TestCheckResourceAttr(resourceName, "columns.0.data_type", "BIGINT"),
					resource.TestCheckResourceAttr(resourceName, "columns.1.name", "email"),
					resource.TestCheckResourceAttr(resourceName, "columns.1.data_type", "VARCHAR"),
					resource.TestCheckResourceAttr(resourceName, "columns.1.precision", "255"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccTableConfig(suffix, "Updated table description", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated table description"),
					resource.TestCheckResourceAttr(resourceName, "columns.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "columns.2.name", "created_at"),
					resource.TestCheckResourceAttr(resourceName, "columns.2.data_type", "TIMESTAMP"),
					resource.TestCheckResourceAttr(resourceName, "columns.2.ordinal_position", "3"),
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

func testAccTableConfig(suffix, description string, withThirdColumn bool) string {
	columns := `
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
    {
      name             = "email"
      data_type        = "VARCHAR"
      precision        = 255
      ordinal_position = 2
    },`
	if withThirdColumn {
		columns += `
    {
      name             = "created_at"
      data_type        = "TIMESTAMP"
      ordinal_position = 3
    },`
	}

	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Acceptance database service for table tests"
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
  description     = %q

  columns = [%s
  ]
}
`, suffix+"-service", suffix+"-db", suffix+"-schema", suffix+"-table", description, columns)
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

func testAccCheckTableDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_table" {
			continue
		}

		_, err := sdk.Tables.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("table %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking table %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
