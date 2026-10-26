data "openmetadata_tag" "confidential" {
  fully_qualified_name = "Sensitivity.Confidential"
}

output "confidential_tag_id" {
  value = data.openmetadata_tag.confidential.id
}
