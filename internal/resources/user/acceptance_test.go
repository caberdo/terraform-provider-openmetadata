package user_test

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

func TestAccUser_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-user")
	email := name + "@example.com"
	teamName := tfacctest.RandomWithPrefix("tf-acc-team")
	policyName := tfacctest.RandomWithPrefix("tf-acc-policy")
	roleName := tfacctest.RandomWithPrefix("tf-acc-role")
	personaName := tfacctest.RandomWithPrefix("tf-acc-persona")
	resourceName := "openmetadata_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(name, email, teamName, policyName, roleName, personaName, "", "", false, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "is_bot", "false"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "teams.*", "openmetadata_team.test", "fully_qualified_name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "roles.*", "openmetadata_role.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "default_persona", personaName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccUserConfig(name, email, teamName, policyName, roleName, personaName, "Acceptance User", "Updated description", true, "Europe/Amsterdam"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance User"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "true"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "Europe/Amsterdam"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "roles.*", "openmetadata_role.test", "fully_qualified_name"),
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

func TestAccUser_botUser(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-bot-user")
	email := name + "@example.com"
	resourceName := "openmetadata_user.bot"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBotUserConfig(name, email, "7"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "is_bot", "true"),
					resource.TestCheckResourceAttr(resourceName, "jwt_token_expiry", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// The server may inherit a system-wide default persona; bot users do not
				// set one, so the imported value is not deterministic.
				ImportStateVerifyIgnore: []string{"jwt_token_expiry", "default_persona"},
			},
		},
	})
}

func testAccUserConfig(name, email, teamName, policyName, roleName, personaName, displayName, description string, isAdmin bool, timezone string) string {
	displayNameLine := ""
	if displayName != "" {
		displayNameLine = fmt.Sprintf("  display_name = %q\n", displayName)
	}
	timezoneLine := ""
	if timezone != "" {
		timezoneLine = fmt.Sprintf("  timezone     = %q\n", timezone)
	}

	return fmt.Sprintf(`
resource "openmetadata_team" "test" {
  name      = %q
  team_type = "Group"
}

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
  policies = [openmetadata_policy.test.fully_qualified_name]
}

resource "openmetadata_persona" "test" {
  name = %q
}

resource "openmetadata_user" "test" {
  name        = %q
  email       = %q
  description = %q
  is_admin    = %t
%s%s  teams = [openmetadata_team.test.fully_qualified_name]
  roles           = [openmetadata_role.test.fully_qualified_name]
  default_persona = openmetadata_persona.test.name
}
`, teamName, policyName, roleName, personaName, name, email, description, isAdmin, displayNameLine, timezoneLine)
}

func testAccBotUserConfig(name, email, jwtTokenExpiry string) string {
	return fmt.Sprintf(`
resource "openmetadata_user" "bot" {
  name             = %q
  email            = %q
  is_bot           = true
  jwt_token_expiry = %q
}
`, name, email, jwtTokenExpiry)
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

func testAccCheckUserDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_user" {
			continue
		}

		_, err := sdk.Users.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("user %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking user %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
