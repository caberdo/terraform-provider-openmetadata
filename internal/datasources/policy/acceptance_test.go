package policy_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccPolicyDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-policy-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_policy.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_policy.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_policy.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_policy.by_name", "enabled", "true"),
					resource.TestCheckResourceAttr("data.openmetadata_policy.by_name", "rules.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.openmetadata_policy.by_name", "rules.*", map[string]string{
						"name":   "view-tables",
						"effect": "allow",
					}),
					resource.TestCheckTypeSetElemAttr("data.openmetadata_policy.by_name", "rules.*.operations.*", "ViewAll"),
					resource.TestCheckTypeSetElemAttr("data.openmetadata_policy.by_name", "rules.*.resources.*", "table"),
					resource.TestCheckResourceAttrPair("data.openmetadata_policy.by_id", "id", "openmetadata_policy.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_policy.by_id", "fully_qualified_name", "openmetadata_policy.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccPolicyDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_policy" "test" {
  name        = %q
  description = "Looked up by name"
  enabled     = true

  rules = [{
    name       = "view-tables"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}

data "openmetadata_policy" "by_name" {
  fully_qualified_name = openmetadata_policy.test.fully_qualified_name
}

data "openmetadata_policy" "by_id" {
  id = openmetadata_policy.test.id
}
`, name)
}
