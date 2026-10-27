resource "openmetadata_api_service" "payments" {
  name         = "payments-api"
  service_type = "Rest"
}

resource "openmetadata_api_collection" "pets" {
  name         = "pets"
  service      = openmetadata_api_service.payments.name
  endpoint_url = "https://api.example.com/v1"
}

resource "openmetadata_api_endpoint" "list_pets" {
  name           = "list-pets"
  api_collection = openmetadata_api_collection.pets.fully_qualified_name
  request_method = "GET"
  endpoint_url   = "https://api.example.com/v1/pets"
  display_name   = "List pets"
  description    = "Returns every pet known to the service."
}
