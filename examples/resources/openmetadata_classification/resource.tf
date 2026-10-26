resource "openmetadata_classification" "sensitivity" {
  name               = "Sensitivity"
  display_name       = "Data sensitivity"
  description        = "How sensitive an asset is."
  mutually_exclusive = true
}
