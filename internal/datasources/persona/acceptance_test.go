package persona_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccPersonaDataSource_basic(t *testing.T) {
	personaName := tfacctest.RandomWithPrefix("tf-acc-persona-ds")
	userName := tfacctest.RandomWithPrefix("tf-acc-persona-ds-user")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccPersonaDataSourceConfig(personaName, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_persona.by_name", "name", personaName),
					resource.TestCheckResourceAttr("data.openmetadata_persona.by_name", "fully_qualified_name", personaName),
					resource.TestCheckResourceAttr("data.openmetadata_persona.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_persona.by_name", "default", "false"),
					resource.TestCheckResourceAttr("data.openmetadata_persona.by_name", "users.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.openmetadata_persona.by_name", "users.*", userName),
					resource.TestCheckResourceAttrPair("data.openmetadata_persona.by_id", "id", "openmetadata_persona.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_persona.by_id", "fully_qualified_name", "openmetadata_persona.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccPersonaDataSourceConfig(personaName, userName string) string {
	return fmt.Sprintf(`
resource "openmetadata_user" "test" {
  name        = %[2]q
  email       = "%[2]s@example.com"
  description = "Persona data source acceptance user"
}

resource "openmetadata_persona" "test" {
  name        = %[1]q
  description = "Looked up by name"
  users       = [openmetadata_user.test.fully_qualified_name]
}

data "openmetadata_persona" "by_name" {
  fully_qualified_name = openmetadata_persona.test.fully_qualified_name
}

data "openmetadata_persona" "by_id" {
  id = openmetadata_persona.test.id
}
`, personaName, userName)
}
