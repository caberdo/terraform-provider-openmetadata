resource "openmetadata_database_service" "warehouse" {
  name         = "warehouse"
  service_type = "Postgres"
  description  = "Warehouse database service."
}

resource "openmetadata_query" "recent_orders" {
  service      = openmetadata_database_service.warehouse.fully_qualified_name
  name         = "recent_orders"
  display_name = "Recent orders"
  description  = "Orders placed in the last day."
  query        = "SELECT * FROM orders WHERE created_at > now() - interval '1 day'"
}

output "recent_orders_query_id" {
  value = openmetadata_query.recent_orders.id
}
