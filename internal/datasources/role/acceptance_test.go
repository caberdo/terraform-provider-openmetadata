package role_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccRoleDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-role-ds")
	policyName := tfacctest.RandomWithPrefix("tf-acc-role-ds-policy")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDataSourceConfig(name, policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_role.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_role.by_name", "fully_qualified_name", name),
					resource.TestCheckTypeSetElemAttrPair("data.openmetadata_role.by_name", "policies.*", "openmetadata_policy.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_role.by_id", "id", "openmetadata_role.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_role.by_id", "fully_qualified_name", "openmetadata_role.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_role.by_id", "policies.#", "openmetadata_role.test", "policies.#"),
				),
			},
		},
	})
}

func testAccRoleDataSourceConfig(name, policyName string) string {
	return fmt.Sprintf(`
resource "openmetadata_policy" "test" {
  name = %q

  rules = [{
    name       = "AllowAll"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}

resource "openmetadata_role" "test" {
  name        = %q
  description = "Looked up by name"
  policies    = [openmetadata_policy.test.fully_qualified_name]
}

data "openmetadata_role" "by_name" {
  fully_qualified_name = openmetadata_role.test.fully_qualified_name
}

data "openmetadata_role" "by_id" {
  id = openmetadata_role.test.id
}
`, policyName, name)
}
