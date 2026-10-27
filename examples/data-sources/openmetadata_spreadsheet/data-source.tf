data "openmetadata_spreadsheet" "revenue" {
  fully_qualified_name = "google-drive.reports.quarterly-revenue"
}

output "revenue_spreadsheet_parent" {
  value = data.openmetadata_spreadsheet.revenue.parent
}
