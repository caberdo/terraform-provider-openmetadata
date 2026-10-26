# Import by UUID.
terraform import openmetadata_chart.revenue_by_region 1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c

# Or import by fully qualified name, which combines the dashboard service name
# with the chart name.
terraform import openmetadata_chart.revenue_by_region looker.revenue_by_region
