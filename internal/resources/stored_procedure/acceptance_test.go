package stored_procedure_test

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

func TestAccStoredProcedure_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-sp-service")
	databaseName := tfacctest.RandomWithPrefix("tf-acc-sp-database")
	schemaName := tfacctest.RandomWithPrefix("tf-acc-sp-schema")
	procedureName := tfacctest.RandomWithPrefix("tf-acc-sp-procedure")
	resourceName := "openmetadata_stored_procedure.test"
	fullyQualifiedName := fmt.Sprintf("%s.%s.%s.%s", serviceName, databaseName, schemaName, procedureName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckStoredProcedureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStoredProcedureConfig(serviceName, databaseName, schemaName, procedureName,
					"Initial description", "BEGIN SELECT 1; END"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", procedureName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", fullyQualifiedName),
					resource.TestCheckResourceAttr(resourceName, "database_schema", fmt.Sprintf("%s.%s.%s", serviceName, databaseName, schemaName)),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "code", "BEGIN SELECT 1; END"),
					resource.TestCheckResourceAttr(resourceName, "language", "SQL"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccStoredProcedureConfig(serviceName, databaseName, schemaName, procedureName,
					"Updated description", "BEGIN SELECT 2; END"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "code", "BEGIN SELECT 2; END"),
					resource.TestCheckResourceAttr(resourceName, "language", "SQL"),
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

func testAccStoredProcedureConfig(serviceName, databaseName, schemaName, procedureName, description, code string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
}

resource "openmetadata_database" "test" {
  name    = %q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name     = %q
  database = openmetadata_database.test.fully_qualified_name
}

resource "openmetadata_stored_procedure" "test" {
  name            = %q
  database_schema = openmetadata_database_schema.test.fully_qualified_name
  description     = %q
  code            = %q
  language        = "SQL"
}
`, serviceName, databaseName, schemaName, procedureName, description, code)
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

func testAccCheckStoredProcedureDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_stored_procedure" {
			continue
		}

		_, err := sdk.StoredProcedures.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("stored procedure %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking stored procedure %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
