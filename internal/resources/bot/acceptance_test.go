package bot_test

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

func TestAccBot_basic(t *testing.T) {
	botUserName := tfacctest.RandomWithPrefix("tf-acc-bot-user")
	botName := tfacctest.RandomWithPrefix("tf-acc-bot")
	resourceName := "openmetadata_bot.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig(botUserName, botName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", botName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", botName),
					resource.TestCheckResourceAttr(resourceName, "bot_user", botUserName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccBotConfigUpdated(botUserName, botName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Bot"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "bot_user", botUserName),
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

// testAccBotConfig creates the bot user the bot is bound to. Under basic auth
// the server keeps the user name equal to the email local part, so both derive
// from the same prefix.
func testAccBotConfig(botUserName, botName, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_user" "bot" {
  name             = %q
  email            = "%s@example.com"
  is_bot           = true
  jwt_token_expiry = "7"
}

resource "openmetadata_bot" "test" {
  name        = %q
  bot_user    = openmetadata_user.bot.name
  description = %q
}
`, botUserName, botUserName, botName, description)
}

func testAccBotConfigUpdated(botUserName, botName, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_user" "bot" {
  name             = %q
  email            = "%s@example.com"
  is_bot           = true
  jwt_token_expiry = "7"
}

resource "openmetadata_bot" "test" {
  name         = %q
  bot_user     = openmetadata_user.bot.name
  display_name = "Acceptance Bot"
  description  = %q
}
`, botUserName, botUserName, botName, description)
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

func testAccCheckBotDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_bot" {
			continue
		}

		_, err := sdk.Bots.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("bot %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking bot %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
