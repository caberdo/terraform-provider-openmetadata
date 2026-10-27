data "openmetadata_test_suite" "quality_gate" {
  fully_qualified_name = "quality_gate"
}

output "quality_gate_test_suite_id" {
  value = data.openmetadata_test_suite.quality_gate.id
}
