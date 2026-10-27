resource "openmetadata_api_service" "payments" {
  name         = "payments-api"
  service_type = "Rest"
  display_name = "Payments API"
  description  = "REST API that serves payment operations."

  connection_config = jsonencode({
    type = "Rest"
    openAPISchemaConnection = {
      openAPISchemaURL = "https://petstore3.swagger.io/api/v3/openapi.json"
    }
  })
}
