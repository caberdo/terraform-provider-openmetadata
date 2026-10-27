data "openmetadata_api_collection" "pets" {
  fully_qualified_name = "payments-api.pets"
}

output "pets_api_collection_id" {
  value = data.openmetadata_api_collection.pets.id
}
