package api_endpoint_test

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

func TestAccApiEndpoint_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-api-endpoint")
	resourceName := "openmetadata_api_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckApiEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApiEndpointConfig(name, "GET", "https://api.example.com/v1/pets", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-list"),
					resource.TestCheckResourceAttr(resourceName, "service", name),
					resource.TestCheckResourceAttr(resourceName, "service_type", "Rest"),
					resource.TestCheckResourceAttrPair(resourceName, "api_collection", "openmetadata_api_collection.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "request_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_url", "https://api.example.com/v1/pets"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name+"."+name+"-pets."+name+"-list"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccApiEndpointConfigUpdated(name, "POST", "https://api.example.com/v2/pets", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance API Endpoint"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "request_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_url", "https://api.example.com/v2/pets"),
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

func testAccApiEndpointConfig(name, method, endpointURL, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_api_service" "test" {
  name         = %q
  service_type = "Rest"
}

resource "openmetadata_api_collection" "test" {
  name    = "%s-pets"
  service = openmetadata_api_service.test.name
}

resource "openmetadata_api_endpoint" "test" {
  name           = "%s-list"
  api_collection = openmetadata_api_collection.test.fully_qualified_name
  request_method = %q
  endpoint_url   = %q
  description    = %q
}
`, name, name, name, method, endpointURL, description)
}

func testAccApiEndpointConfigUpdated(name, method, endpointURL, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_api_service" "test" {
  name         = %q
  service_type = "Rest"
}

resource "openmetadata_api_collection" "test" {
  name    = "%s-pets"
  service = openmetadata_api_service.test.name
}

resource "openmetadata_api_endpoint" "test" {
  name           = "%s-list"
  api_collection = openmetadata_api_collection.test.fully_qualified_name
  request_method = %q
  endpoint_url   = %q
  display_name   = "Acceptance API Endpoint"
  description    = %q
}
`, name, name, name, method, endpointURL, description)
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

func testAccCheckApiEndpointDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_api_endpoint" {
			continue
		}

		_, err := sdk.APIEndpoints.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("API endpoint %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking API endpoint %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
