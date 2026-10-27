resource "openmetadata_test_suite" "quality_gate" {
  name         = "quality_gate"
  display_name = "Data Quality Gate"
  description  = "Logical test suite grouping the data quality checks promoted to production."
}
