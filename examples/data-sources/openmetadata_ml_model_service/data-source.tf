data "openmetadata_ml_model_service" "churn_tracking" {
  fully_qualified_name = "churn-tracking"
}

output "churn_tracking_ml_model_service_id" {
  value = data.openmetadata_ml_model_service.churn_tracking.id
}
