package search_index_test

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

func TestAccSearchIndex_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-search-index")
	serviceName := name + "-service"
	resourceName := "openmetadata_search_index.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckSearchIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSearchIndexConfig(serviceName, name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.name", "message"),
					resource.TestCheckResourceAttr(resourceName, "fields.0.data_type", "TEXT"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccSearchIndexConfigUpdated(serviceName, name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Search Index"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "1"),
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

func testAccSearchIndexConfig(serviceName, name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_search_service" "test" {
  name         = %q
  service_type = "ElasticSearch"
}

resource "openmetadata_search_index" "test" {
  name        = %q
  service     = openmetadata_search_service.test.name
  description = %q

  fields = [
    {
      name      = "message"
      data_type = "TEXT"
    },
  ]
}
`, serviceName, name, description)
}

func testAccSearchIndexConfigUpdated(serviceName, name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_search_service" "test" {
  name         = %q
  service_type = "ElasticSearch"
}

resource "openmetadata_search_index" "test" {
  name         = %q
  service      = openmetadata_search_service.test.name
  display_name = "Acceptance Search Index"
  description  = %q

  fields = [
    {
      name      = "message"
      data_type = "TEXT"
    },
  ]
}
`, serviceName, name, description)
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

func testAccCheckSearchIndexDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_search_index" {
			continue
		}

		_, err := sdk.SearchIndexes.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("search index %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking search index %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
