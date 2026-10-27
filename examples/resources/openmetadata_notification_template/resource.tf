resource "openmetadata_notification_template" "asset_created" {
  name         = "asset-created-email"
  display_name = "Asset created (email)"
  description  = "Subject and body used by the data platform team when an asset is created."

  template_subject = "New asset: {{ entity.name }}"

  template_body = <<-EOT
    Hello {{ user.displayName }},

    A new {{ entity.entityType }} named {{ entity.name }} was created.
    Open it here: {{ entity.href }}
  EOT
}
