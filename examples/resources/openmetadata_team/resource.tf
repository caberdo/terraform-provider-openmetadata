resource "openmetadata_policy" "team_policy" {
  name = "platform-team-policy"

  rules = [{
    name       = "view-all"
    effect     = "allow"
    operations = ["ViewAll"]
    resources  = ["table"]
  }]
}

resource "openmetadata_role" "data_steward" {
  name     = "platform-data-steward"
  policies = [openmetadata_policy.team_policy.name]
}

# The root of the hierarchy. A team created without parents is placed under
# Organization, so declare the business unit first and nest everything below it.
resource "openmetadata_team" "engineering" {
  name      = "engineering"
  team_type = "BusinessUnit"
}

resource "openmetadata_team" "platform" {
  name         = "platform"
  display_name = "Platform"
  description  = "Owns the platform services and their metadata."
  team_type    = "Group"
  email        = "platform@example.com"
  is_joinable  = false

  parents       = [openmetadata_team.engineering.name]
  policies      = [openmetadata_policy.team_policy.name]
  default_roles = [openmetadata_role.data_steward.name]
  users         = ["jane.doe"]
}
