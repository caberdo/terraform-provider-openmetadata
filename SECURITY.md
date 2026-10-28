# Security Policy

## Supported versions

Security fixes are issued for the latest published provider release. The
provider targets OpenMetadata 1.13 and later; the version the acceptance suite
runs against is pinned in
[`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt).

## Threat model

The provider holds credentials for an OpenMetadata server. Treat a Terraform run
as equivalent to a privileged OpenMetadata session:

- The provider can create, update and delete metadata across every supported
  entity: glossaries, classifications, tags, services, databases, tables, teams,
  policies, roles and the rest. A destroy is permanent unless the server's soft
  delete applies, and it is visible to every platform user.
- The provider can change governance and access objects (`policy`, `role`,
  `team`, `user`, `bot`, `persona`), so a mistaken apply can widen permissions.
- Credentials live in Terraform configuration or environment variables. The
  provider does not write `token` or `password` to state, but state does contain
  the metadata it manages, including attributes marked `Sensitive` such as
  service connector configuration. Terraform does not encrypt state; protect the
  backend and restrict who can read it.
- A token or personal access token is used until it expires. Rotate it through
  the provider configuration; the provider caches the JWT only for the process
  lifetime.

## Secret handling

- `token` and `password` are marked `Sensitive` in the provider schema.
- Attributes that carry service credentials or connection strings are marked
  `Sensitive` in the resource and data source schemas.
- The login exchange and debug logs never include the token, the password or the
  login response body. Log fields are limited to the host, API version and
  authentication mode.
- Raw response bodies are bounded to 1 MiB, so an unexpected HTML error page
  cannot exhaust memory or be logged verbatim.
- No secret may appear in a diagnostic, log field, test failure message or
  committed file. Acceptance tests read credentials from the environment and run
  against a disposable local stack; do not point them at a production server.
- Release checksums are signed and build provenance is attested; verify a
  download before running it.

## Reporting a vulnerability

Report vulnerabilities privately through GitHub's
[security advisories](https://github.com/caberdo/terraform-provider-openmetadata/security/advisories/new)
for this repository. Do not open a public issue for a suspected vulnerability.

Please include:

- the affected provider version and the OpenMetadata version,
- a description of the issue and its impact,
- the configuration involved, with tokens, passwords and hostnames removed,
- reproduction steps or a proof of concept,
- any suggested remediation.

You will receive an acknowledgement, and a coordinated disclosure date once the
issue is understood.
