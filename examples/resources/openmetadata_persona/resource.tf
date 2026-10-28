resource "openmetadata_user" "alice" {
  name  = "alice"
  email = "alice@example.com"
}

resource "openmetadata_persona" "data_steward" {
  name         = "DataSteward"
  display_name = "Data Steward"
  description  = "Owns data quality and governance for the organisation."
  default      = true
  users        = [openmetadata_user.alice.fully_qualified_name]
}
