package persona_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
)

func TestAccPersona_basic(t *testing.T) {
	personaName := tfacctest.RandomWithPrefix("tf-acc-persona")
	firstUser := tfacctest.RandomWithPrefix("tf-acc-persona-user")
	secondUser := tfacctest.RandomWithPrefix("tf-acc-persona-user2")
	resourceName := "openmetadata_persona.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckPersonaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPersonaConfig(personaName, firstUser, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", personaName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", personaName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "default", "false"),
					resource.TestCheckResourceAttr(resourceName, "users.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "users.*", firstUser),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccPersonaConfigUpdated(personaName, firstUser, secondUser, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Persona"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "default", "true"),
					testAccCheckPersonaUsers(resourceName, firstUser, secondUser),
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

func testAccPersonaUserConfig(label, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_user" %[1]q {
  name        = %[2]q
  email       = "%[2]s@example.com"
  description = "Persona acceptance user"
}
`, label, name)
}

func testAccPersonaConfig(personaName, userName, description string) string {
	return testAccPersonaUserConfig("first", userName) + fmt.Sprintf(`
resource "openmetadata_persona" "test" {
  name        = %q
  description = %q
  users       = [openmetadata_user.first.fully_qualified_name]
}
`, personaName, description)
}

func testAccPersonaConfigUpdated(personaName, firstUser, secondUser, description string) string {
	return testAccPersonaUserConfig("first", firstUser) + testAccPersonaUserConfig("second", secondUser) + fmt.Sprintf(`
resource "openmetadata_persona" "test" {
  name         = %q
  display_name = "Acceptance Persona"
  description  = %q
  default      = true
  users = [
    openmetadata_user.first.fully_qualified_name,
    openmetadata_user.second.fully_qualified_name,
  ]
}
`, personaName, description)
}

// testAccCheckPersonaUsers compares the assigned users as a set so the check
// does not depend on the order the server returns them in.
func testAccCheckPersonaUsers(resourceName string, want ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		got := make(map[string]bool, len(want))
		count := 0
		for key, value := range rs.Primary.Attributes {
			if strings.HasPrefix(key, "users.") && key != "users.#" {
				got[value] = true
				count++
			}
		}

		if count != len(want) {
			return fmt.Errorf("persona %s has %d users, want %d: %v", resourceName, count, len(want), got)
		}
		for _, name := range want {
			if !got[name] {
				return fmt.Errorf("persona %s is missing user %q: %v", resourceName, name, got)
			}
		}
		return nil
	}
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

func testAccCheckPersonaDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_persona" {
			continue
		}

		_, err := sdk.Personas.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("persona %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking persona %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
