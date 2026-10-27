data "openmetadata_event_subscription" "finance_updates" {
  fully_qualified_name = "finance-updates"
}

output "finance_updates_subscription_id" {
  value = data.openmetadata_event_subscription.finance_updates.id
}
