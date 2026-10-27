data "openmetadata_worksheet" "q1" {
  fully_qualified_name = "google-drive.budget.q1"
}

output "q1_worksheet_id" {
  value = data.openmetadata_worksheet.q1.id
}
