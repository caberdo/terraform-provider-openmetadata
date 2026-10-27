# Import by UUID.
terraform import openmetadata_test_case.users_row_count 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, assigned by the server when the test case
# is created.
terraform import openmetadata_test_case.users_row_count warehouse.analytics.public.users.users_row_count
