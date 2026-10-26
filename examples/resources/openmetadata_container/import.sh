# Import by UUID.
terraform import openmetadata_container.events 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, in the form service.parent.container.
terraform import openmetadata_container.events s3-bucket.raw.events
