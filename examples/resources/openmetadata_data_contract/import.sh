# Import by UUID.
terraform import openmetadata_data_contract.users 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, in the form service.database.schema.table.contract.
terraform import openmetadata_data_contract.users warehouse.analytics.public.users.users-monthly-reporting
