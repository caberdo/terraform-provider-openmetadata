resource "openmetadata_search_service" "elastic" {
  name         = "elastic"
  service_type = "ElasticSearch"
}

resource "openmetadata_search_index" "messages" {
  name         = "messages"
  service      = openmetadata_search_service.elastic.name
  display_name = "Messages"
  description  = "Search index for application messages."

  fields = [
    {
      name      = "message"
      data_type = "TEXT"
    },
  ]
}
