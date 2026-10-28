data "openmetadata_role" "data_steward" {
  fully_qualified_name = "DataSteward"
}

output "data_steward_role_id" {
  value = data.openmetadata_role.data_steward.id
}

output "data_steward_role_policies" {
  value = data.openmetadata_role.data_steward.policies
}
