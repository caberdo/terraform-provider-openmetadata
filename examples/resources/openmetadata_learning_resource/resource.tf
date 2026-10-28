resource "openmetadata_learning_resource" "data_governance_101" {
  name          = "data_governance_101"
  resource_type = "Article"
  display_name  = "Data Governance 101"
  description   = "Introduction to the data governance programme."
  categories    = ["AI", "DataGovernance"]
  difficulty    = "Intro"
  status        = "Active"

  estimated_duration   = 15
  completion_threshold = 0.8

  contexts = [
    {
      page_id      = "home"
      component_id = "learning"
      priority     = 1
    },
  ]

  source = [
    {
      url      = "https://example.com/academy/data-governance-101"
      provider = "user"
    },
  ]
}
