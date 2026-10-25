resource "openmetadata_glossary" "finance" {
  name               = "finance"
  display_name       = "Finance"
  description        = "Glossary owned by the finance data team."
  mutually_exclusive = false
}
