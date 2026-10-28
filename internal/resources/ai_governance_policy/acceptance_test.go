package ai_governance_policy_test

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

func TestAccAIGovernancePolicy_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-ai-governance-policy")
	resourceName := "openmetadata_ai_governance_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckAIGovernancePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAIGovernancePolicyConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "BiasThreshold"),
					resource.TestCheckResourceAttr(resourceName, "enforcement_level", "Advisory"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAIGovernancePolicyConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Bias Policy"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "policy_type", "BiasThreshold"),
					resource.TestCheckResourceAttr(resourceName, "enforcement_level", "Advisory"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
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

func testAccAIGovernancePolicyConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_ai_governance_policy" "test" {
  name              = %q
  policy_type       = "BiasThreshold"
  enforcement_level = "Advisory"
  enabled           = true
  description       = %q
}
`, name, description)
}

func testAccAIGovernancePolicyConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_ai_governance_policy" "test" {
  name              = %q
  policy_type       = "BiasThreshold"
  enforcement_level = "Advisory"
  enabled           = true
  display_name      = "Acceptance Bias Policy"
  description       = %q
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

func testAccCheckAIGovernancePolicyDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_ai_governance_policy" {
			continue
		}

		_, err := sdk.AIGovernancePolicies.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("AI governance policy %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking AI governance policy %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
