data "openmetadata_dashboard_data_model" "sales" {
  fully_qualified_name = "warehouse.model.sales_model"
}

output "sales_data_model_id" {
  value = data.openmetadata_dashboard_data_model.sales.id
}

output "sales_data_model_columns" {
  value = data.openmetadata_dashboard_data_model.sales.columns
}
