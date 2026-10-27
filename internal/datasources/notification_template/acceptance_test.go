package notification_template_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccNotificationTemplateDataSource_basic(t *testing.T) {
	name := tfacctest.RandomWithPrefix("tf-acc-notification-template-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTemplateDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_notification_template.by_name", "name", name),
					resource.TestCheckResourceAttr("data.openmetadata_notification_template.by_name", "fully_qualified_name", name),
					resource.TestCheckResourceAttr("data.openmetadata_notification_template.by_name", "description", "Looked up by name"),
					resource.TestCheckResourceAttr("data.openmetadata_notification_template.by_name", "template_subject", "Data source subject"),
					resource.TestCheckResourceAttrPair("data.openmetadata_notification_template.by_id", "id", "openmetadata_notification_template.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_notification_template.by_id", "fully_qualified_name", "openmetadata_notification_template.test", "fully_qualified_name"),
					resource.TestCheckResourceAttrPair("data.openmetadata_notification_template.by_id", "template_body", "openmetadata_notification_template.test", "template_body"),
				),
			},
		},
	})
}

func testAccNotificationTemplateDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openmetadata_notification_template" "test" {
  name             = %q
  description      = "Looked up by name"
  template_subject = "Data source subject"
  template_body    = "Data source body for {{ entity.name }}"
}

data "openmetadata_notification_template" "by_name" {
  fully_qualified_name = openmetadata_notification_template.test.fully_qualified_name
}

data "openmetadata_notification_template" "by_id" {
  id = openmetadata_notification_template.test.id
}
`, name)
}
