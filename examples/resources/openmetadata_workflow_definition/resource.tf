resource "openmetadata_workflow_definition" "contract_approval" {
  name               = "contract_approval"
  display_name       = "Contract approval"
  description        = "Approval workflow used by the data contract review process."
  type               = "noOp"
  store_stage_status = false
}
