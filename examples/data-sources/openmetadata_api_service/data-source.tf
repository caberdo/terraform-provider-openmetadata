data "openmetadata_api_service" "payments" {
  fully_qualified_name = "payments-api"
}

output "payments_api_service_id" {
  value = data.openmetadata_api_service.payments.id
}
