data "openmetadata_glossary" "finance" {
  fully_qualified_name = "finance"
}

output "finance_glossary_id" {
  value = data.openmetadata_glossary.finance.id
}
