package domain_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDomainDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-domain-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_domain.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_domain.by_name", "domain_type", "Aggregate"),
					resource.TestCheckResourceAttr("data.openmetadata_domain.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttrPair("data.openmetadata_domain.by_id", "id", "openmetadata_domain.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_domain.by_id", "fully_qualified_name", "openmetadata_domain.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDomainDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "test" {
  name        = %q
  domain_type = "Aggregate"
  description = "Looked up by name"
}

data "openmetadata_domain" "by_name" {
  fully_qualified_name = openmetadata_domain.test.fully_qualified_name
}

data "openmetadata_domain" "by_id" {
  id = openmetadata_domain.test.id
}
`, name)
}
