package document_test

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

func TestAccDocument_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-document")
	resourceName := "openmetadata_document.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckDocumentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocumentConfig(name, "Initial", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "entity_type", "KnowledgePanel"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name+".acceptance"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "data", `{"body":"Initial","title":"Acceptance"}`),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccDocumentConfigUpdated(name, "Updated", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Document"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "data", `{"body":"Updated","title":"Acceptance"}`),
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

func testAccDocumentConfig(name, body, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_document" "test" {
  name                 = %q
  entity_type          = "KnowledgePanel"
  fully_qualified_name = %q
  description          = %q
  data = jsonencode({
    title = "Acceptance"
    body  = %q
  })
}
`, name, name+".acceptance", description, body)
}

func testAccDocumentConfigUpdated(name, body, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_document" "test" {
  name                 = %q
  entity_type          = "KnowledgePanel"
  fully_qualified_name = %q
  display_name         = "Acceptance Document"
  description          = %q
  data = jsonencode({
    title = "Acceptance"
    body  = %q
  })
}
`, name, name+".acceptance", description, body)
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

func testAccCheckDocumentDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_document" {
			continue
		}

		_, err := sdk.Documents.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("document %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking document %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
