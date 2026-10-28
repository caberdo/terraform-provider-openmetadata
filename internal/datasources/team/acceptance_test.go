package team_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTeamDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-team-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTeamDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "team_type", "Group"),
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "email", name+"@example.com"),
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "is_joinable", "true"),
					// A team created without parents is placed under Organization.
					resource.TestCheckResourceAttr("data.openmetadata_team.by_name", "parents.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.openmetadata_team.by_name", "parents.*", "Organization"),
					resource.TestCheckResourceAttrPair("data.openmetadata_team.by_id", "id", "openmetadata_team.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_team.by_id", "fully_qualified_name", "openmetadata_team.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_team.by_id", "team_type", "openmetadata_team.test", "team_type"),
				),
			},
		},
	})
}

func testAccTeamDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_team" "test" {
  name        = %q
  team_type   = "Group"
  description = "Looked up by name"
  email       = %q
}

data "openmetadata_team" "by_name" {
  fully_qualified_name = openmetadata_team.test.fully_qualified_name
}

data "openmetadata_team" "by_id" {
  id = openmetadata_team.test.id
}
`, name, name+"@example.com")
}
