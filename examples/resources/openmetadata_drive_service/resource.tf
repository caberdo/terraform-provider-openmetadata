resource "openmetadata_drive_service" "team_drive" {
  name         = "team-drive"
  service_type = "GoogleDrive"
  display_name = "Team drive"
  description  = "Drive service for the analytics team."
}
