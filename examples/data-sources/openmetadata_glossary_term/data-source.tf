data "openmetadata_glossary_term" "petty_cash" {
  fully_qualified_name = "finance.PettyCash"
}

output "petty_cash_term_id" {
  value = data.openmetadata_glossary_term.petty_cash.id
}
