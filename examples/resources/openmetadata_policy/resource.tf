resource "openmetadata_policy" "viewer" {
  name         = "viewer"
  display_name = "Viewer"
  description  = "Read-only access to tables and dashboards."
  enabled      = true

  rules = [
    {
      name        = "view-tables"
      effect      = "allow"
      operations  = ["ViewAll"]
      resources   = ["table"]
      description = "View every table."
    },
    {
      name       = "view-dashboards"
      effect     = "allow"
      operations = ["ViewAll", "ViewQueries"]
      # "All" is the canonical resource for every entity type. It cannot be
      # combined with other resources, and lowercase "all" is rejected.
      resources = ["All"]
    },
  ]
}
