package ai_governance_policy_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccAIGovernancePolicyDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-ai-governance-policy-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccAIGovernancePolicyDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_ai_governance_policy.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_ai_governance_policy.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_ai_governance_policy.by_name", "policy_type", "BiasThreshold"),
					resource.TestCheckResourceAttr("data.openmetadata_ai_governance_policy.by_name", "enforcement_level", "Advisory"),
					resource.TestCheckResourceAttr("data.openmetadata_ai_governance_policy.by_name", "enabled", "true"),
					resource.TestCheckResourceAttr("data.openmetadata_ai_governance_policy.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ai_governance_policy.by_id", "id", "openmetadata_ai_governance_policy.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_ai_governance_policy.by_id", "fully_qualified_name", "openmetadata_ai_governance_policy.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccAIGovernancePolicyDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_ai_governance_policy" "test" {
  name              = %q
  policy_type       = "BiasThreshold"
  enforcement_level = "Advisory"
  enabled           = true
  description       = "Looked up by name"
}

data "openmetadata_ai_governance_policy" "by_name" {
  fully_qualified_name = openmetadata_ai_governance_policy.test.fully_qualified_name
}

data "openmetadata_ai_governance_policy" "by_id" {
  id = openmetadata_ai_governance_policy.test.id
}
`, name)
}
