package glossary_term_test

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

func TestAccGlossaryTerm_basic(t *testing.T) {
	glossaryName := tfacctest.RandomWithPrefix("tf-acc-term-glossary")
	name := tfacctest.RandomWithPrefix("tf-acc-term")
	resourceName := "openmetadata_glossary_term.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckGlossaryTermDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlossaryTermConfig(glossaryName, name, "Initial description", `["alias-one"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "glossary", glossaryName),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", glossaryName+"."+name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "synonyms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "synonyms.0", "alias-one"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccGlossaryTermConfig(glossaryName, name, "Updated description", `["alias-one", "alias-two"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Term"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "synonyms.#", "2"),
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

func testAccGlossaryTermConfig(glossaryName, name, description, synonyms string) string {
	return fmt.Sprintf(`
resource "openmetadata_glossary" "parent" {
  name = %q
}

resource "openmetadata_glossary_term" "test" {
  glossary     = openmetadata_glossary.parent.name
  name         = %q
  display_name = "Acceptance Term"
  description  = %q
  synonyms     = %s
}
`, glossaryName, name, description, synonyms)
}

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

func testAccCheckGlossaryTermDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_glossary_term" {
			continue
		}

		_, err := sdk.GlossaryTerms.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("glossary term %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking glossary term %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
