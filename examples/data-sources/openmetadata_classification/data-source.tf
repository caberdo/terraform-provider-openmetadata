data "openmetadata_classification" "sensitivity" {
  fully_qualified_name = "Sensitivity"
}

output "sensitivity_classification_id" {
  value = data.openmetadata_classification.sensitivity.id
}
