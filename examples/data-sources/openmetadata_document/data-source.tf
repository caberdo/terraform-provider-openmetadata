data "openmetadata_document" "getting_started" {
  fully_qualified_name = "getting-started.acceptance"
}

output "getting_started_document_id" {
  value = data.openmetadata_document.getting_started.id
}
