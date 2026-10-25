terraform {
  required_providers {
    openmetadata = {
      source = "caberdo/openmetadata"
    }
  }
}

# Credentials are read from OPENMETADATA_TOKEN, or from
# OPENMETADATA_USERNAME together with OPENMETADATA_PASSWORD, when the
# attributes below are omitted. Prefer the environment over committed values.
provider "openmetadata" {
  host = "https://openmetadata.example.com"
}
