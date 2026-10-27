data "openmetadata_test_case" "users_row_count" {
  fully_qualified_name = "warehouse.analytics.public.users.users_row_count"
}

output "users_row_count_id" {
  value = data.openmetadata_test_case.users_row_count.id
}

output "users_row_count_status" {
  value = data.openmetadata_test_case.users_row_count.test_case_status
}

output "users_row_count_parameters" {
  value = {
    for parameter in data.openmetadata_test_case.users_row_count.parameter_values :
    parameter.name => parameter.value
  }
}
