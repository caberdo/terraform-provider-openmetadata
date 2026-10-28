data "openmetadata_policy" "viewer" {
  fully_qualified_name = "viewer"
}

output "viewer_policy_id" {
  value = data.openmetadata_policy.viewer.id
}

output "viewer_policy_rules" {
  value = data.openmetadata_policy.viewer.rules
}
