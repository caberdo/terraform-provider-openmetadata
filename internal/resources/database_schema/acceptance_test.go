package database_schema_test

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

func TestAccDatabaseSchema_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-db-service")
	databaseName := tfacctest.RandomWithPrefix("tf-acc-db")
	schemaName := tfacctest.RandomWithPrefix("tf-acc-schema")
	resourceName := "openmetadata_database_schema.test"
	wantFQN := serviceName + "." + databaseName + "." + schemaName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDatabaseSchemaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseSchemaConfig(serviceName, databaseName, schemaName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", wantFQN),
					resource.TestCheckResourceAttr(resourceName, "database", serviceName+"."+databaseName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccDatabaseSchemaConfigUpdated(serviceName, databaseName, schemaName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Schema"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", wantFQN),
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

func testAccDatabaseSchemaConfig(serviceName, databaseName, schemaName, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %[1]q
  service_type = "Postgres"
  description  = "Acceptance database service"
}

resource "openmetadata_database" "test" {
  name    = %[2]q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name        = %[3]q
  database    = openmetadata_database.test.fully_qualified_name
  description = %[4]q
}
`, serviceName, databaseName, schemaName, description)
}

func testAccDatabaseSchemaConfigUpdated(serviceName, databaseName, schemaName, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %[1]q
  service_type = "Postgres"
  description  = "Acceptance database service"
}

resource "openmetadata_database" "test" {
  name    = %[2]q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name         = %[3]q
  database     = openmetadata_database.test.fully_qualified_name
  display_name = "Acceptance Schema"
  description  = %[4]q
}
`, serviceName, databaseName, schemaName, description)
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

func testAccCheckDatabaseSchemaDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_database_schema" {
			continue
		}

		_, err := sdk.DatabaseSchemas.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("database schema %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking database schema %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
