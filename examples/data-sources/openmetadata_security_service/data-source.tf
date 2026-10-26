data "openmetadata_security_service" "ranger" {
  fully_qualified_name = "ranger"
}

output "ranger_security_service_id" {
  value = data.openmetadata_security_service.ranger.id
}
