package prompt_template_test

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

func TestAccPromptTemplate_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-prompt-template")
	resourceName := "openmetadata_prompt_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckPromptTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptTemplateConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "template_content", "Classify the ticket: {{ticket}}"),
					resource.TestCheckResourceAttr(resourceName, "template_type", "Classification"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccPromptTemplateConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Prompt Template"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
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

func testAccPromptTemplateConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_prompt_template" "test" {
  name             = %q
  template_content = "Classify the ticket: {{ticket}}"
  template_type    = "Classification"
  description      = "Initial description"
}
`, name)
}

func testAccPromptTemplateConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_prompt_template" "test" {
  name             = %q
  display_name     = "Acceptance Prompt Template"
  template_content = "Classify the ticket: {{ticket}}"
  template_type    = "Classification"
  description      = "Updated description"
}
`, name)
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

func testAccCheckPromptTemplateDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_prompt_template" {
			continue
		}

		_, err := sdk.PromptTemplates.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("prompt template %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking prompt template %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
