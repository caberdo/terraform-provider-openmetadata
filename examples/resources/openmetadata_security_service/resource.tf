resource "openmetadata_security_service" "ranger" {
  name         = "ranger"
  service_type = "Ranger"
  display_name = "Ranger"
  description  = "Security service for the Ranger policy source."
}
