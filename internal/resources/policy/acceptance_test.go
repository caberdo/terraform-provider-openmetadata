package policy_test

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

func TestAccPolicy_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-policy")
	resourceName := "openmetadata_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"name":   "view-tables",
						"effect": "allow",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.operations.*", "ViewAll"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.resources.*", "table"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccPolicyConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"name":        "view-basic",
						"effect":      "allow",
						"condition":   "isOwner",
						"description": "View owned assets",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rules.*", map[string]string{
						"name":   "deny-delete",
						"effect": "deny",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.operations.*", "ViewBasic"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.operations.*", "ViewQueries"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.operations.*", "Delete"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.resources.*", "table"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rules.*.resources.*", "dashboard"),
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

func testAccPolicyConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_policy" "test" {
  name        = %q
  description = %q
  enabled     = true

  rules = [{
    name       = "view-tables"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}
`, name, description)
}

func testAccPolicyConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_policy" "test" {
  name         = %q
  display_name = "Acceptance Policy"
  description  = %q
  enabled      = false

  rules = [
    {
      name        = "view-basic"
      effect      = "allow"
      operations  = ["ViewBasic", "ViewQueries"]
      resources   = ["table", "dashboard"]
      condition   = "isOwner"
      description = "View owned assets"
    },
    {
      name       = "deny-delete"
      effect     = "deny"
      operations = ["Delete"]
      resources  = ["table"]
    },
  ]
}
`, name, description)
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

func testAccCheckPolicyDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_policy" {
			continue
		}

		_, err := sdk.Policies.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("policy %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking policy %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
