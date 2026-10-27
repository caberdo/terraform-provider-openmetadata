package llm_model_test

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

func TestAccLlmModel_basic(t *testing.T) {
	serviceName := tfacctest.RandomWithPrefix("tf-acc-llm-service")
	name := tfacctest.RandomWithPrefix("tf-acc-llm-model")
	resourceName := "openmetadata_llm_model.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckLlmModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLlmModelConfig(serviceName, name, "", "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
					resource.TestCheckResourceAttr(resourceName, "base_model", "gpt-4o"),
					resource.TestCheckResourceAttr(resourceName, "model_provider", "openai"),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", serviceName+"."+name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccLlmModelConfig(serviceName, name, "Acceptance LLM Model", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance LLM Model"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "base_model", "gpt-4o"),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
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

func testAccLlmModelConfig(serviceName, name, displayName, description string) string {
	displayLine := ""
	if displayName != "" {
		displayLine = fmt.Sprintf("  display_name = %q\n", displayName)
	}
	return fmt.Sprintf(`
resource "openmetadata_llm_service" "test" {
  name         = %q
  service_type = "OpenAI"
}

resource "openmetadata_llm_model" "test" {
  name           = %q
  service        = openmetadata_llm_service.test.fully_qualified_name
  base_model     = "gpt-4o"
  model_provider = "openai"
%[4]s  description    = %[5]q
}
`, serviceName, name, displayName, displayLine, description)
}

// importStateIDByFullyQualifiedName exercises the name-based import path rather
// than the UUID path used by the default import step. The LLM model fully
// qualified name is service.name.
func importStateIDByFullyQualifiedName(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		service := rs.Primary.Attributes["service"]
		name := rs.Primary.Attributes["name"]
		if service == "" || name == "" {
			return "", fmt.Errorf("resource %s has no service or name", resourceName)
		}
		return service + "." + name, nil
	}
}

func testAccCheckLlmModelDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_llm_model" {
			continue
		}

		_, err := sdk.LLMModels.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("LLM model %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking LLM model %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
