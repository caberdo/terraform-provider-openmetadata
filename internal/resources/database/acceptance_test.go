package database_test

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

func TestAccDatabase_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-db-service")
	databaseName := tfacctest.RandomWithPrefix("tf-acc-db")
	resourceName := "openmetadata_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig(serviceName, databaseName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", databaseName),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+"."+databaseName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccDatabaseConfigUpdated(serviceName, databaseName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Database"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+"."+databaseName),
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

func testAccDatabaseConfig(serviceName, databaseName, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
}

resource "openmetadata_database" "test" {
  name        = %q
  service     = openmetadata_database_service.test.name
  description = %q
}
`, serviceName, databaseName, description)
}

func testAccDatabaseConfigUpdated(serviceName, databaseName, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
}

resource "openmetadata_database" "test" {
  name         = %q
  service      = openmetadata_database_service.test.name
  display_name = "Acceptance Database"
  description  = %q
}
`, serviceName, databaseName, description)
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

func testAccCheckDatabaseDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_database" {
			continue
		}

		_, err := sdk.Databases.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("database %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking database %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
