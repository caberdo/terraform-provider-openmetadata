package bot_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccBotDataSource_basic(t *testing.T) {
	botUserName := tfacctest.RandomWithPrefix("tf-acc-bot-user-ds")
	botName := tfacctest.RandomWithPrefix("tf-acc-bot-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccBotDataSourceConfig(botUserName, botName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_bot.by_name", "name", botName),
					resource.TestCheckResourceAttr("data.openmetadata_bot.by_name", "fully_qualified_name", botName),
					resource.TestCheckResourceAttr("data.openmetadata_bot.by_name", "bot_user", botUserName),
					resource.TestCheckResourceAttr("data.openmetadata_bot.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_bot.by_id", "id", "openmetadata_bot.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_bot.by_id", "fully_qualified_name", "openmetadata_bot.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccBotDataSourceConfig(botUserName, botName string) string {
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
  description = "Looked up by name"
}

data "openmetadata_bot" "by_name" {
  fully_qualified_name = openmetadata_bot.test.fully_qualified_name
}

data "openmetadata_bot" "by_id" {
  id = openmetadata_bot.test.id
}
`, botUserName, botUserName, botName)
}
