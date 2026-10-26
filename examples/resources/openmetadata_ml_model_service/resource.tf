resource "openmetadata_ml_model_service" "churn_tracking" {
  name         = "churn-tracking"
  service_type = "Mlflow"
  display_name = "Churn model tracking"
  description  = "ML model service backing the churn models."
}
