data "openmetadata_api_endpoint" "list_pets" {
  fully_qualified_name = "payments-api.pets.list-pets"
}

output "list_pets_api_endpoint_id" {
  value = data.openmetadata_api_endpoint.list_pets.id
}
