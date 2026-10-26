resource "openmetadata_domain" "engineering" {
  name         = "Engineering"
  display_name = "Engineering"
  description  = "Engineering-owned data assets."
  domain_type  = "Source-aligned"
}

resource "openmetadata_domain" "platform" {
  name        = "Platform"
  domain_type = "Aggregate"
  parent      = openmetadata_domain.engineering.fully_qualified_name
}
