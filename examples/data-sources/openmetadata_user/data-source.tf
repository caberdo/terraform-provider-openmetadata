data "openmetadata_user" "jane" {
  fully_qualified_name = "jane.doe"
}

output "jane_user_id" {
  value = data.openmetadata_user.jane.id
}
