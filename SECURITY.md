# Security Policy

## Supported versions

Security fixes are issued for the latest published provider release. The
provider targets OpenMetadata 1.13 and later; the version the acceptance suite
runs against is pinned in
[`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt).

## Reporting a vulnerability

Report vulnerabilities privately through GitHub's
[security advisories](https://github.com/caberdo/terraform-provider-openmetadata/security/advisories/new)
for this repository. Do not open a public issue for a suspected vulnerability.

Please include the provider version, the OpenMetadata version, the configuration
involved (with tokens, passwords and hostnames removed), and the steps to
reproduce.

## Credential handling

- The provider accepts a bearer token, or a username and password, through
  attributes or environment variables. `token` and `password` are marked
  `Sensitive` in the schema.
- Provider configuration is used in memory and is not written to Terraform
  state.
- Diagnostics and debug logs never include the token, password or the response
  body of the login exchange.
- Acceptance tests read credentials from the environment and run against a
  disposable local stack. Do not point them at a production server.
