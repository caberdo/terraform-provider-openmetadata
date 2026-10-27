data "openmetadata_ai_governance_policy" "bias_review" {
  fully_qualified_name = "bias-review"
}

output "bias_review_policy_id" {
  value = data.openmetadata_ai_governance_policy.bias_review.id
}
