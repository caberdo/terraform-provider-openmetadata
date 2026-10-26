# Import by UUID.
terraform import openmetadata_dashboard_data_model.sales 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, which is <service>.model.<data model name>.
terraform import openmetadata_dashboard_data_model.sales warehouse.model.sales_model
