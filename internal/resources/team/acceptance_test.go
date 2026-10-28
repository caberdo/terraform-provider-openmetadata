package team_test

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

func TestAccTeam_basic(t *testing.T) {
	prefix := tfacctest.RandomWithPrefix("tf-acc-team")
	resourceName := "openmetadata_team.group"
	departmentName := prefix + "-department"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTeamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamConfig(prefix, "Initial description", prefix+"-group@example.com", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", prefix+"-group"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", prefix+"-group"),
					resource.TestCheckResourceAttr(resourceName, "team_type", "Group"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "email", prefix+"-group@example.com"),
					resource.TestCheckResourceAttr(resourceName, "is_joinable", "true"),
					resource.TestCheckResourceAttr(resourceName, "parents.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "parents.*", departmentName),
					resource.TestCheckResourceAttr(resourceName, "policies.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "policies.*", prefix+"-policy"),
					resource.TestCheckResourceAttr(resourceName, "default_roles.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_roles.*", prefix+"-role"),
					resource.TestCheckResourceAttr(resourceName, "default_persona", prefix+"-persona"),
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "users.*", prefix+"-user"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr("openmetadata_team.department", "parents.#", "1"),
					resource.TestCheckTypeSetElemAttr("openmetadata_team.department", "parents.*", prefix+"-business-unit"),
					resource.TestCheckResourceAttr("openmetadata_team.department", "team_type", "Department"),
				),
			},
			{
				Config: testAccTeamConfig(prefix, "Updated description", prefix+"-group-updated@example.com", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "email", prefix+"-group-updated@example.com"),
					resource.TestCheckResourceAttr(resourceName, "is_joinable", "false"),
					resource.TestCheckResourceAttr(resourceName, "team_type", "Group"),
					resource.TestCheckResourceAttr(resourceName, "parents.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
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

// testAccTeamConfig builds a full hierarchy plus every relationship a team
// supports. The BusinessUnit is the root, the Department proves explicit
// parents, and the Group carries policies, default roles, a default persona and
// users.
func testAccTeamConfig(prefix, description, email string, joinable bool) string {
	user := prefix + "-user"

	return fmt.Sprintf(`
resource "openmetadata_policy" "test" {
  name = %q

  rules = [{
    name       = "view-tables"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}

resource "openmetadata_role" "test" {
  name     = %q
  policies = [openmetadata_policy.test.name]
}

resource "openmetadata_persona" "test" {
  name = %q
}

resource "openmetadata_user" "test" {
  name  = %q
  email = %q
}

resource "openmetadata_team" "business_unit" {
  name      = %q
  team_type = "BusinessUnit"
}

resource "openmetadata_team" "department" {
  name      = %q
  team_type = "Department"
  parents   = [openmetadata_team.business_unit.name]
}

resource "openmetadata_team" "group" {
  name            = %q
  display_name    = "Acceptance group"
  description     = %q
  team_type       = "Group"
  email           = %q
  is_joinable     = %t
  parents         = [openmetadata_team.department.name]
  policies        = [openmetadata_policy.test.name]
  default_roles   = [openmetadata_role.test.name]
  default_persona = openmetadata_persona.test.name
  users           = [openmetadata_user.test.name]
}
`, prefix+"-policy", prefix+"-role", prefix+"-persona", user, user+"@example.com",
		prefix+"-business-unit", prefix+"-department", prefix+"-group", description, email, joinable)
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

func testAccCheckTeamDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_team" {
			continue
		}

		_, err := sdk.Teams.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("team %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking team %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
