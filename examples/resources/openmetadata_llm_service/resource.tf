resource "openmetadata_llm_service" "openai" {
  name         = "openai"
  service_type = "OpenAI"
  display_name = "OpenAI"
  description  = "LLM service backing prompt and model metadata."
}
