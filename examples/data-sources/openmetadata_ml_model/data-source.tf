data "openmetadata_ml_model" "churn" {
  fully_qualified_name = "mlflow.ChurnModel"
}

output "churn_model_id" {
  value = data.openmetadata_ml_model.churn.id
}
