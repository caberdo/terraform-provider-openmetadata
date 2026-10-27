package test_case_test

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

func TestAccTestCase_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-test-case")
	resourceName := "openmetadata_test_case.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckTestCaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTestCaseConfig(suffix, "", "Initial test case description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", suffix+"-case"),
					resource.TestCheckResourceAttrSet(resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial test case description"),
					resource.TestCheckResourceAttrSet(resourceName, "test_definition"),
					resource.TestCheckResourceAttr(resourceName, "compute_passed_failed_row_count", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_dynamic_assertion", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "entity_fqn"),
					resource.TestCheckResourceAttrSet(resourceName, "test_suite"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccTestCaseConfig(suffix, "Acceptance Test Case", "Updated test case description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Test Case"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test case description"),
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

func TestAccTestCaseDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-test-case-ds")
	resourceName := "openmetadata_test_case.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTestCaseDataSourceConfigByID(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_id", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_id", "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_id", "fully_qualified_name", resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_test_case.by_id", "description", "Data source test case"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_case.by_id", "test_definition"),
					resource.TestCheckResourceAttrSet("data.openmetadata_test_case.by_id", "entity_fqn"),
				),
			},
			{
				Config: testAccTestCaseDataSourceConfigByName(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_name", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_test_case.by_name", "fully_qualified_name", resourceName, "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_test_case.by_name", "description", "Data source test case"),
				),
			},
		},
	})
}

// testAccTestCaseConfig builds a Postgres table and a reusable test definition
// so the test case has a real entity link to attach to.
func testAccTestCaseConfig(suffix, displayName, description string) string {
	displayNameBlock := ""
	if displayName != "" {
		displayNameBlock = fmt.Sprintf("\n  display_name = %q", displayName)
	}

	return fmt.Sprintf(`
resource "openmetadata_database_service" "test" {
  name         = %q
  service_type = "Postgres"
  description  = "Acceptance database service for test case tests"
}

resource "openmetadata_database" "test" {
  name    = %q
  service = openmetadata_database_service.test.name
}

resource "openmetadata_database_schema" "test" {
  name     = %q
  database = openmetadata_database.test.fully_qualified_name
}

resource "openmetadata_table" "test" {
  name            = %q
  database_schema = openmetadata_database_schema.test.fully_qualified_name

  columns = [
    {
      name             = "id"
      data_type        = "BIGINT"
      ordinal_position = 1
    },
  ]
}

resource "openmetadata_test_definition" "test" {
  name           = %q
  entity_type    = "TABLE"
  test_platforms = ["Other"]
}

resource "openmetadata_test_case" "test" {
  name            = %q
  entity_link     = "<#E::table::${openmetadata_table.test.fully_qualified_name}>"
  test_definition = openmetadata_test_definition.test.fully_qualified_name%s
  description     = %q
}
`, suffix+"-service", suffix+"-db", suffix+"-schema", suffix+"-table",
		suffix+"-definition", suffix+"-case", displayNameBlock, description)
}

func testAccTestCaseDataSourceConfigByID(suffix string) string {
	return testAccTestCaseConfig(suffix, "", "Data source test case") + `
data "openmetadata_test_case" "by_id" {
  id = openmetadata_test_case.test.id
}
`
}

func testAccTestCaseDataSourceConfigByName(suffix string) string {
	return testAccTestCaseConfig(suffix, "", "Data source test case") + `
data "openmetadata_test_case" "by_name" {
  fully_qualified_name = openmetadata_test_case.test.fully_qualified_name
}
`
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

func testAccCheckTestCaseDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_test_case" {
			continue
		}

		_, err := sdk.TestCases.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("test case %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking test case %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
