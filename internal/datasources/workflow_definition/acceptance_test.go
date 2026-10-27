package workflow_definition_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccWorkflowDefinitionDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-workflow-definition-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowDefinitionDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_workflow_definition.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_workflow_definition.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_workflow_definition.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_workflow_definition.by_name", "store_stage_status", "false"),
					resource.TestCheckResourceAttrPair("data.openmetadata_workflow_definition.by_id", "id", "openmetadata_workflow_definition.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_workflow_definition.by_id", "fully_qualified_name", "openmetadata_workflow_definition.test", "fully_qualified_name"),
				),
			},
		},
	})
}

func testAccWorkflowDefinitionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_workflow_definition" "test" {
  name        = %q
  type        = "noOp"
  description = "Looked up by name"
}

data "openmetadata_workflow_definition" "by_name" {
  fully_qualified_name = openmetadata_workflow_definition.test.fully_qualified_name
}

data "openmetadata_workflow_definition" "by_id" {
  id = openmetadata_workflow_definition.test.id
}
`, name)
}
