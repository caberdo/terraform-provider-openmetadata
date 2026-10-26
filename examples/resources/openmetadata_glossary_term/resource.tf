resource "openmetadata_glossary" "finance" {
  name = "finance"
}

resource "openmetadata_glossary_term" "petty_cash" {
  glossary     = openmetadata_glossary.finance.name
  name         = "PettyCash"
  display_name = "Petty cash"
  description  = "Small cash expenses."
  synonyms     = ["cash on hand", "float"]
}
