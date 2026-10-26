# Import by UUID.
terraform import openmetadata_topic.order_created 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, in the form service.topic.
terraform import openmetadata_topic.order_created events.order_created
