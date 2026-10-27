resource "openmetadata_drive_service" "shared" {
  name         = "google-drive"
  service_type = "GoogleDrive"
}

resource "openmetadata_directory" "root" {
  name        = "root"
  service     = openmetadata_drive_service.shared.name
  description = "Org-wide shared drive root."
}

resource "openmetadata_directory" "reports" {
  name         = "reports"
  service      = openmetadata_drive_service.shared.name
  parent       = openmetadata_directory.root.fully_qualified_name
  display_name = "Reports"
  description  = "Quarterly and annual reports."
}
