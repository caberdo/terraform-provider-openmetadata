data "openmetadata_search_index" "messages" {
  fully_qualified_name = "elastic.messages"
}

output "messages_search_index_id" {
  value = data.openmetadata_search_index.messages.id
}

output "messages_search_index_fields" {
  value = [for field in data.openmetadata_search_index.messages.fields : field.name]
}
