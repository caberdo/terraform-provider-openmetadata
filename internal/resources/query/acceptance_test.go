package query_test

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

func TestAccQuery_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-query-svc")
	name := tfacctest.RandomWithPrefix("tf-acc-query")
	resourceName := "openmetadata_query.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryConfig(serviceName, name, "SELECT * FROM orders", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+"."+name),
					resource.TestCheckResourceAttr(resourceName, "query", "SELECT * FROM orders"),
					resource.TestCheckResourceAttr(resourceName, "query_type", "SELECT"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "service"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccQueryConfig(serviceName, name, "SELECT * FROM orders WHERE id > 100", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "query", "SELECT * FROM orders WHERE id > 100"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+"."+name),
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

func testAccQueryConfig(serviceName, name, query, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Service for query acceptance tests"
}

resource "openmetadata_query" "test" {
  service     = openmetadata_database_service.test.fully_qualified_name
  name        = %q
  query       = %q
  query_type  = "SELECT"
  description = %q
}
`, serviceName, name, query, description)
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

func testAccCheckQueryDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_query" {
			continue
		}

		// Destroy is a soft delete, so the server hides the query from normal
		// reads instead of removing the row.
		_, err := sdk.Queries.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("query %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking query %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
