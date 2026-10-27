data "openmetadata_drive_service" "team_drive" {
  fully_qualified_name = "team-drive"
}

output "team_drive_service_id" {
  value = data.openmetadata_drive_service.team_drive.id
}
