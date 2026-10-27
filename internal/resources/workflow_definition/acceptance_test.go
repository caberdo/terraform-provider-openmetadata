package workflow_definition_test

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

func TestAccWorkflowDefinition_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-workflow-definition")
	resourceName := "openmetadata_workflow_definition.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckWorkflowDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowDefinitionConfig(name, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "noOp"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "store_stage_status", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "deployed"),
					resource.TestCheckResourceAttrSet(resourceName, "suspended"),
				),
			},
			{
				Config: testAccWorkflowDefinitionConfigUpdated(name, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Workflow Definition"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "type", "noOp"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type"},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"type"},
				ImportStateIdFunc:       importStateIDByFullyQualifiedName(resourceName),
			},
		},
	})
}

func testAccWorkflowDefinitionConfig(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_workflow_definition" "test" {
  name        = %q
  type        = "noOp"
  description = %q
}
`, name, description)
}

func testAccWorkflowDefinitionConfigUpdated(name, description string) string {
	return fmt.Sprintf(`
resource "openmetadata_workflow_definition" "test" {
  name         = %q
  display_name = "Acceptance Workflow Definition"
  type         = "noOp"
  description  = %q
}
`, name, description)
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

func testAccCheckWorkflowDefinitionDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_workflow_definition" {
			continue
		}

		_, err := sdk.WorkflowDefinitions.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("workflow definition %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking workflow definition %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
