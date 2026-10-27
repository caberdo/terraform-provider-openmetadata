data "openmetadata_workflow_definition" "contract_approval" {
  fully_qualified_name = "contract_approval"
}

output "contract_approval_workflow_id" {
  value = data.openmetadata_workflow_definition.contract_approval.id
}
