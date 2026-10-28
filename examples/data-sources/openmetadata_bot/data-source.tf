data "openmetadata_bot" "reporting" {
  fully_qualified_name = "reporting"
}

output "reporting_bot_id" {
  value = data.openmetadata_bot.reporting.id
}

output "reporting_bot_user" {
  value = data.openmetadata_bot.reporting.bot_user
}
