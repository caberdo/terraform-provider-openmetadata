resource "openmetadata_user" "jane" {
  name         = "jane.doe"
  email        = "jane.doe@example.com"
  display_name = "Jane Doe"
  description  = "Data engineer on the analytics platform team."
  is_admin     = false
  timezone     = "Europe/Amsterdam"

  teams = ["analytics"]
  roles = ["DataConsumer"]
}
