data "openmetadata_llm_model" "chat" {
  fully_qualified_name = "openai.gpt-4o"
}

output "chat_model_id" {
  value = data.openmetadata_llm_model.chat.id
}
