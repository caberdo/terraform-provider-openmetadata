resource "openmetadata_document" "getting_started" {
  name                 = "getting-started"
  entity_type          = "KnowledgePanel"
  fully_qualified_name = "getting-started.acceptance"
  display_name         = "Getting started"
  description          = "Onboarding notes for the data platform."
  data = jsonencode({
    title = "Getting started"
    body  = "Read this first."
  })
}
