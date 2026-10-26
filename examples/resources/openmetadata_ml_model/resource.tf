resource "openmetadata_ml_model_service" "mlflow" {
  name         = "mlflow"
  service_type = "Mlflow"
  display_name = "MLflow"
  description  = "ML model service for the experiment tracking server."
}

resource "openmetadata_ml_model" "churn" {
  name         = "ChurnModel"
  service      = openmetadata_ml_model_service.mlflow.fully_qualified_name
  algorithm    = "linear_regression"
  display_name = "Churn model"
  description  = "Predicts whether a customer will churn."
}
