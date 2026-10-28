data "openmetadata_persona" "data_steward" {
  fully_qualified_name = "DataSteward"
}

output "data_steward_users" {
  value = data.openmetadata_persona.data_steward.users
}
