resource "openmetadata_llm_service" "openai" {
  name         = "openai"
  service_type = "OpenAI"
  display_name = "OpenAI"
  description  = "LLM service for the OpenAI API."
}

resource "openmetadata_llm_model" "chat" {
  name           = "gpt-4o"
  service        = openmetadata_llm_service.openai.fully_qualified_name
  base_model     = "gpt-4o"
  model_provider = "openai"
  display_name   = "GPT-4o"
  description    = "General-purpose chat model used by the support assistant."
}
