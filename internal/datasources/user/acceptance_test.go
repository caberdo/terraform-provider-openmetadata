package user_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccUserDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-user-ds")
	email := name + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_user.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_user.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_user.by_name", "email", email),
					resource.TestCheckResourceAttr("data.openmetadata_user.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_user.by_id", "id", "openmetadata_user.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_user.by_id", "fully_qualified_name", "openmetadata_user.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig(name, email string) string {
	return fmt.Sprintf(`
resource "openmetadata_user" "test" {
  name        = %q
  email       = %q
  description = "Looked up by name"
}

data "openmetadata_user" "by_name" {
  fully_qualified_name = openmetadata_user.test.fully_qualified_name
}

data "openmetadata_user" "by_id" {
  id = openmetadata_user.test.id
}
`, name, email)
}
