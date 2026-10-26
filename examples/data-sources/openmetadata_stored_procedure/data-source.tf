data "openmetadata_stored_procedure" "customer_report" {
  fully_qualified_name = "warehouse.analytics.public.customer_report"
}

output "customer_report_id" {
  value = data.openmetadata_stored_procedure.customer_report.id
}

output "customer_report_code" {
  value = data.openmetadata_stored_procedure.customer_report.code
}
