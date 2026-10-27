package document_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccDocumentDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-document-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_document.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_document.by_name", "entity_type", "KnowledgePanel"),
					resource.TestCheckResourceAttr("data.openmetadata_document.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_document.by_name", "data", `{"body":"Initial","title":"Acceptance"}`),
					resource.TestCheckResourceAttrPair("data.openmetadata_document.by_id", "id", "openmetadata_document.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_document.by_id", "fully_qualified_name", "openmetadata_document.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_document.by_id", "entity_type", "KnowledgePanel"),
					resource.TestCheckResourceAttr("data.openmetadata_document.by_id", "data", `{"body":"Initial","title":"Acceptance"}`),
				),
			},
		},
	})
}

func testAccDocumentDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_document" "test" {
  name                 = %q
  entity_type          = "KnowledgePanel"
  fully_qualified_name = %q
  description          = "Looked up by name"
  data = jsonencode({
    title = "Acceptance"
    body  = "Initial"
  })
}

data "openmetadata_document" "by_name" {
  fully_qualified_name = openmetadata_document.test.fully_qualified_name
}

data "openmetadata_document" "by_id" {
  id = openmetadata_document.test.id
}
`, name, name+".acceptance")
}
