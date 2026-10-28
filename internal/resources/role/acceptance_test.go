package role_test

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

func TestAccRole_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-role")
	policyA := tfacctest.RandomWithPrefix("tf-acc-role-policy-a")
	policyB := tfacctest.RandomWithPrefix("tf-acc-role-policy-b")
	resourceName := "openmetadata_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig(name, policyA, policyB, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "policies.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policies.*", "openmetadata_policy.a", "fully_qualified_name"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccRoleConfigUpdated(name, policyA, policyB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Role"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "policies.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policies.*", "openmetadata_policy.b", "fully_qualified_name"),
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

func testAccPolicyConfig(policyA, policyB string) string {
	return fmt.Sprintf(`
resource "openmetadata_policy" "a" {
  name = %q

  rules = [{
    name       = "AllowAll"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}

resource "openmetadata_policy" "b" {
  name = %q

  rules = [{
    name       = "AllowAll"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}
`, policyA, policyB)
}

func testAccRoleConfig(name, policyA, policyB, description string) string {
	return testAccPolicyConfig(policyA, policyB) + fmt.Sprintf(`
resource "openmetadata_role" "test" {
  name        = %q
  description = %q
  policies    = [openmetadata_policy.a.fully_qualified_name]
}
`, name, description)
}

func testAccRoleConfigUpdated(name, policyA, policyB string) string {
	return testAccPolicyConfig(policyA, policyB) + fmt.Sprintf(`
resource "openmetadata_role" "test" {
  name         = %q
  display_name = "Acceptance Role"
  description  = "Updated description"
  policies     = [openmetadata_policy.b.fully_qualified_name]
}
`, name)
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

func testAccCheckRoleDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_role" {
			continue
		}

		_, err := sdk.Roles.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("role %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking role %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
