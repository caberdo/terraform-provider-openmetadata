resource "openmetadata_ai_governance_policy" "bias_review" {
  name              = "bias-review"
  display_name      = "Bias review"
  description       = "Flags AI models whose bias score exceeds the review threshold."
  policy_type       = "BiasThreshold"
  enforcement_level = "Advisory"
  enabled           = true
}
