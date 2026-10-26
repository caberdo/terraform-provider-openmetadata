# Import by UUID.
terraform import openmetadata_domain.engineering 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, which for a top-level domain is its name.
terraform import openmetadata_domain.engineering Engineering

# Nested domains use the parent.child form.
terraform import openmetadata_domain.platform Engineering.Platform
