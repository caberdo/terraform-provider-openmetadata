package data_product_test

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

func TestAccDataProduct_basic(t *testing.T) {
	domainName := tfacctest.RandomWithPrefix("tf-acc-product-domain")
	secondDomainName := tfacctest.RandomWithPrefix("tf-acc-product-domain-2")
	name := tfacctest.RandomWithPrefix("tf-acc-data-product")
	resourceName := "openmetadata_data_product.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDataProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataProductConfig(domainName, name, "Initial description", "openmetadata_domain.parent.fully_qualified_name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domains.0", domainName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccDataProductConfigWithDisplay(domainName, name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Data Product"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
			{
				// OpenMetadata's default entity rule allows one domain per
				// data product, so exercise the patch path by switching to a
				// different domain rather than adding a second one.
				Config: testAccDataProductConfigSecondDomain(domainName, secondDomainName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "domains.0", secondDomainName),
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

func testAccDataProductConfig(domainName, name, description, domains string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "parent" {
  name        = %q
  domain_type = "Aggregate"
}

resource "openmetadata_data_product" "test" {
  name        = %q
  description = %q
  domains     = [%s]
}
`, domainName, name, description, domains)
}

func testAccDataProductConfigWithDisplay(domainName, name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "parent" {
  name        = %q
  domain_type = "Aggregate"
}

resource "openmetadata_data_product" "test" {
  name         = %q
  display_name = "Acceptance Data Product"
  description  = %q
  domains      = [openmetadata_domain.parent.fully_qualified_name]
}
`, domainName, name, description)
}

func testAccDataProductConfigSecondDomain(domainName, secondDomainName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "parent" {
  name        = %q
  domain_type = "Aggregate"
}

resource "openmetadata_domain" "second" {
  name        = %q
  domain_type = "Aggregate"
}

resource "openmetadata_data_product" "test" {
  name         = %q
  display_name = "Acceptance Data Product"
  description  = "Updated description"
  domains      = [openmetadata_domain.second.fully_qualified_name]
}
`, domainName, secondDomainName, name)
}

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

func testAccCheckDataProductDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_data_product" {
			continue
		}

		_, err := sdk.DataProducts.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("data product %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking data product %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
