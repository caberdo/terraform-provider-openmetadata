package domain_test

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

func TestAccDomain_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-domain")
	childName := tfacctest.RandomWithPrefix("tf-acc-subdomain")
	resourceName := "openmetadata_domain.test"
	childResourceName := "openmetadata_domain.child"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig(name, "Aggregate", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "domain_type", "Aggregate"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccDomainConfig(name, "Consumer-aligned", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Domain"),
					resource.TestCheckResourceAttr(resourceName, "domain_type", "Consumer-aligned"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
			{
				Config: testAccNestedDomainConfig(name, childName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(childResourceName, "parent", name),
					resource.TestCheckResourceAttr(childResourceName, "fully_qualified_name", name+"."+childName),
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

func testAccDomainConfig(name, domainType, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "test" {
  name         = %q
  display_name = "Acceptance Domain"
  domain_type  = %q
  description  = %q
}
`, name, domainType, description)
}

func testAccNestedDomainConfig(name, childName string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "test" {
  name         = %q
  display_name = "Acceptance Domain"
  domain_type  = "Consumer-aligned"
  description  = "Updated description"
}

resource "openmetadata_domain" "child" {
  name        = %q
  domain_type = "Source-aligned"
  parent      = openmetadata_domain.test.fully_qualified_name
}
`, name, childName)
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

func testAccCheckDomainDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_domain" {
			continue
		}

		_, err := sdk.Domains.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("domain %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking domain %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
