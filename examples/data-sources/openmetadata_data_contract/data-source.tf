data "openmetadata_data_contract" "users" {
  fully_qualified_name = "warehouse.analytics.public.users.users-monthly-reporting"
}

output "users_data_contract_id" {
  value = data.openmetadata_data_contract.users.id
}

output "users_data_contract_status" {
  value = data.openmetadata_data_contract.users.entity_status
}
