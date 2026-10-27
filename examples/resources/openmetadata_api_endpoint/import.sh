# Import by UUID.
terraform import openmetadata_api_endpoint.list_pets 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, in the form service.collection.endpoint.
terraform import openmetadata_api_endpoint.list_pets payments-api.pets.list-pets
