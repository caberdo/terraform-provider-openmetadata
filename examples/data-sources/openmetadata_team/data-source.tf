data "openmetadata_team" "platform" {
  fully_qualified_name = "platform"
}

output "platform_team_id" {
  value = data.openmetadata_team.platform.id
}

output "platform_team_users" {
  value = data.openmetadata_team.platform.users
}
