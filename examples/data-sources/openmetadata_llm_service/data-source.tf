data "openmetadata_llm_service" "openai" {
  fully_qualified_name = "openai"
}

output "openai_llm_service_id" {
  value = data.openmetadata_llm_service.openai.id
}
