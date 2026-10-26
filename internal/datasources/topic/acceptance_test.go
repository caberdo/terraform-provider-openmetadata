package topic_test

import (
	"fmt"
	"testing"

	tfacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/caberdo/terraform-provider-openmetadata/internal/acctest"
)

func TestAccTopicDataSource_basic(t *testing.T) {
	suffix := tfacctest.RandomWithPrefix("tf-acc-topic-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicDataSourceConfig(suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openmetadata_topic.by_name", "name", suffix+"-topic"),
					resource.TestCheckResourceAttrPair("data.openmetadata_topic.by_name", "fully_qualified_name", "openmetadata_topic.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_topic.by_name", "partitions", "3"),
					resource.TestCheckResourceAttr("data.openmetadata_topic.by_name", "cleanup_policies.#", "1"),
					resource.TestCheckResourceAttrPair("data.openmetadata_topic.by_id", "id", "openmetadata_topic.test", "id"),
					resource.TestCheckResourceAttrPair("data.openmetadata_topic.by_id", "fully_qualified_name", "openmetadata_topic.test", "fully_qualified_name"),
					resource.TestCheckResourceAttr("data.openmetadata_topic.by_id", "partitions", "3"),
				),
			},
		},
	})
}

func testAccTopicDataSourceConfig(suffix string) string {
	return fmt.Sprintf(`
resource "openmetadata_messaging_service" "test" {
  name         = %q
  service_type = "CustomMessaging"
}

resource "openmetadata_topic" "test" {
  name        = %q
  service     = openmetadata_messaging_service.test.fully_qualified_name
  description = "Looked up by name and id"
  partitions  = 3

  cleanup_policies = ["delete"]
}

data "openmetadata_topic" "by_name" {
  fully_qualified_name = openmetadata_topic.test.fully_qualified_name
}

data "openmetadata_topic" "by_id" {
  id = openmetadata_topic.test.id
}
`, suffix+"-service", suffix+"-topic")
}
