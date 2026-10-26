resource "openmetadata_classification" "sensitivity" {
  name        = "Sensitivity"
  description = "How sensitive an asset is."
}

resource "openmetadata_tag" "confidential" {
  classification = openmetadata_classification.sensitivity.name
  name           = "Confidential"
  display_name   = "Confidential"
  description    = "Access restricted to approved staff."
}
