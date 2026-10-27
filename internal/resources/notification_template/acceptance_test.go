package notification_template_test

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

func TestAccNotificationTemplate_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-notification-template")
	resourceName := "openmetadata_notification_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckNotificationTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTemplateConfig(name, "Initial description", "Asset created", "Body before"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "fully_qualified_name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttr(resourceName, "template_subject", "Asset created"),
					resource.TestCheckResourceAttr(resourceName, "template_body", "Body before"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "is_modified_from_default"),
				),
			},
			{
				Config: testAccNotificationTemplateConfigUpdated(name, "Updated description", "Body after {{ entity.name }}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Notification Template"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "template_body", "Body after {{ entity.name }}"),
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

func testAccNotificationTemplateConfig(name, description, subject, body string) string {
	return fmt.Sprintf(`
resource "openmetadata_notification_template" "test" {
  name             = %q
  description      = %q
  template_subject = %q
  template_body    = %q
}
`, name, description, subject, body)
}

func testAccNotificationTemplateConfigUpdated(name, description, body string) string {
	return fmt.Sprintf(`
resource "openmetadata_notification_template" "test" {
  name             = %q
  display_name     = "Acceptance Notification Template"
  description      = %q
  template_subject = "Asset created"
  template_body    = "%s"
}
`, name, description, body)
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

func testAccCheckNotificationTemplateDestroy(state *terraform.State) error {
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
		if rs.Type != "openmetadata_notification_template" {
			continue
		}

		_, err := sdk.NotificationTemplates.GetByID(ctx, rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("notification template %s still exists", rs.Primary.ID)
		}
		if !client.NotFound(err) {
			return fmt.Errorf("checking notification template %s: %w", rs.Primary.ID, err)
		}
	}

	return nil
}
