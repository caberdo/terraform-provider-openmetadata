data "openmetadata_file" "quarterly" {
  fully_qualified_name = "google-drive.reports.quarterly-report"
}

output "quarterly_file_id" {
  value = data.openmetadata_file.quarterly.id
}
