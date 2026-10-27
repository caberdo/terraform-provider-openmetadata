resource "openmetadata_drive_service" "google_drive" {
  name         = "google-drive"
  service_type = "GoogleDrive"
}

resource "openmetadata_spreadsheet" "budget" {
  name    = "budget"
  service = openmetadata_drive_service.google_drive.name
}

resource "openmetadata_worksheet" "q1" {
  name         = "q1"
  spreadsheet  = openmetadata_spreadsheet.budget.fully_qualified_name
  display_name = "Q1 budget"
  description  = "First-quarter budget figures."
}
