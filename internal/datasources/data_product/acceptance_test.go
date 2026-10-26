package data_product_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDataProductDataSource_basic(t *testing.T) {
	domainName := tfacctest.RandomWithPrefix("tf-acc-product-ds-domain")
	name := tfacctest.RandomWithPrefix("tf-acc-product-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProductDataSourceConfig(domainName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_data_product.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_data_product.by_name", "domains.#", "1"),
					resource.TestCheckResourceAttr("data.openmetadata_data_product.by_name", "domains.0", domainName),
					resource.TestCheckResourceAttrPair("data.openmetadata_data_product.by_id", "id", "openmetadata_data_product.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_data_product.by_id", "fully_qualified_name", "openmetadata_data_product.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccDataProductDataSourceConfig(domainName, name string) string {
	return fmt.Sprintf(`
resource "openmetadata_domain" "parent" {
  name        = %q
  domain_type = "Aggregate"
}

resource "openmetadata_data_product" "test" {
  name        = %q
  description = "Looked up by name"
  domains     = [openmetadata_domain.parent.fully_qualified_name]
}

data "openmetadata_data_product" "by_name" {
  fully_qualified_name = openmetadata_data_product.test.fully_qualified_name
}

data "openmetadata_data_product" "by_id" {
  id = openmetadata_data_product.test.id
}
`, domainName, name)
}
