resource "openmetadata_api_service" "payments" {
  name         = "payments-api"
  service_type = "Rest"
}

resource "openmetadata_api_collection" "pets" {
  name         = "pets"
  service      = openmetadata_api_service.payments.name
  endpoint_url = "https://api.example.com/v1"
  display_name = "Pets"
  description  = "Pet operations exposed by the payments API."
}
