data "openmetadata_query" "recent_orders" {
  fully_qualified_name = "warehouse.recent_orders"
}

output "recent_orders_query_text" {
  value = data.openmetadata_query.recent_orders.query
}
