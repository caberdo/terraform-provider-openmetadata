resource "openmetadata_drive_service" "reports" {
  name         = "google-drive"
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "quarterly" {
  name    = "reports"
  service = openmetadata_drive_service.reports.fully_qualified_name
}

resource "openmetadata_spreadsheet" "revenue" {
  name         = "quarterly-revenue"
  service      = openmetadata_drive_service.reports.fully_qualified_name
  parent       = openmetadata_directory.quarterly.fully_qualified_name
  display_name = "Quarterly revenue"
  description  = "Revenue workbook refreshed every quarter."
}
