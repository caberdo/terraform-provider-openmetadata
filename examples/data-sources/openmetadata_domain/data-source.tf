data "openmetadata_domain" "engineering" {
  fully_qualified_name = "Engineering"
}

output "engineering_domain_id" {
  value = data.openmetadata_domain.engineering.id
}
