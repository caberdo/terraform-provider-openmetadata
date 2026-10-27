data "openmetadata_directory" "reports" {
  fully_qualified_name = "google-drive.root.reports"
}

output "reports_directory_id" {
  value = data.openmetadata_directory.reports.id
}
