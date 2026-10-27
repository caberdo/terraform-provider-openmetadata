data "openmetadata_learning_resource" "data_governance_101" {
  fully_qualified_name = "data_governance_101"
}

output "data_governance_101_categories" {
  value = data.openmetadata_learning_resource.data_governance_101.categories
}

output "data_governance_101_source_url" {
  value = data.openmetadata_learning_resource.data_governance_101.source[0].url
}
