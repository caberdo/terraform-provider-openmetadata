resource "openmetadata_policy" "data_steward" {
  name        = "DataStewardPolicy"
  description = "Allows data stewards to curate data assets."

  rules = [{
    name       = "curate-assets"
    effect     = "allow"
    operations = ["ViewAll", "EditDescription"]
    resources  = ["table"]
  }]
}

resource "openmetadata_role" "data_steward" {
  name         = "DataSteward"
  display_name = "Data Steward"
  description  = "Curates and documents data assets."
  policies     = [openmetadata_policy.data_steward.fully_qualified_name]
}
