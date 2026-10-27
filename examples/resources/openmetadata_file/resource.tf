resource "openmetadata_drive_service" "google_drive" {
  name         = "google-drive"
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "reports" {
  name        = "reports"
  service     = openmetadata_drive_service.google_drive.name
  description = "Shared finance reports."
}

resource "openmetadata_file" "quarterly" {
  name         = "quarterly-report"
  service      = openmetadata_drive_service.google_drive.name
  directory    = openmetadata_directory.reports.fully_qualified_name
  display_name = "Quarterly report"
  description  = "Quarterly finance report as a PDF."
}
