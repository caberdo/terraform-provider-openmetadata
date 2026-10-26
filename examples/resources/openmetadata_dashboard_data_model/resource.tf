resource "openmetadata_dashboard_service" "warehouse" {
  name         = "warehouse"
  service_type = "CustomDashboard"
}

resource "openmetadata_dashboard_data_model" "sales" {
  name            = "sales_model"
  display_name    = "Sales Model"
  description     = "Model behind the sales dashboard."
  service         = openmetadata_dashboard_service.warehouse.fully_qualified_name
  data_model_type = "PowerBIDataModel"
  sql             = "SELECT region, sum(amount) AS total_sales FROM sales GROUP BY region"

  columns = [
    {
      name         = "region"
      display_name = "Region"
      data_type    = "VARCHAR"
      data_length  = 256
      description  = "Sales region."
    },
    {
      name         = "total_sales"
      display_name = "Total Sales"
      data_type    = "DOUBLE"
      description  = "Sum of all sales in the region."
    },
  ]
}
