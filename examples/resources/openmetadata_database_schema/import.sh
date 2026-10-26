# Import by UUID.
terraform import openmetadata_database_schema.public 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, in the form database.schema.
terraform import openmetadata_database_schema.public local_postgres.sales.public
