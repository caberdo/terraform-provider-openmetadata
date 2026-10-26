# Terraform Provider for OpenMetadata

A Terraform provider for the [OpenMetadata](https://open-metadata.org/) metadata
platform. It manages glossaries, classifications, tags, services and the entities
that hang off them through the OpenMetadata REST API.

The provider is a thin translation layer between Terraform state and
OpenMetadata. It owns no truth of its own: every read is an API call, every
mutation is a create, patch or delete, and attributes the server cannot return
are preserved rather than zeroed.

> **Status:** early development. The module, tooling, CI and acceptance stack are
> in place. The glossary reference slice, the wave 1 taxonomy entities
> (`classification`, `tag`, `glossary_term`, `domain`, `data_product`, `metric`),
> the wave 2 database entities (`database_service`, `database`, `database_schema`,
> `table`, `stored_procedure`, `query`), the wave 3 BI entities (`chart`,
> `dashboard`, `dashboard_data_model`, `dashboard_service`), the wave 4
> pipeline and messaging entities (`pipeline_service`, `pipeline`,
> `messaging_service`, `topic`) and the wave 5 platform services
> (`storage_service`, `container`, `ml_model_service`, `ml_model`,
> `search_service`, `search_index`, `metadata_service`,
> `security_service`) are implemented. The remaining entities are tracked in
> [TASKS.md](TASKS.md) and in the wave epics on GitHub.

## Requirements

- Terraform 1.9 or later (plugin protocol 6.0).
- OpenMetadata **1.13** or later. CI runs the acceptance suite against the
  version pinned in
  [`.github/openmetadata-version.txt`](.github/openmetadata-version.txt).

## Quick start

```hcl
terraform {
  required_providers {
    openmetadata = {
      source = "caberdo/openmetadata"
    }
  }
}

provider "openmetadata" {
  host = "https://openmetadata.example.com"
}
```

Credentials are read from `OPENMETADATA_TOKEN`, or from
`OPENMETADATA_USERNAME` together with `OPENMETADATA_PASSWORD`, when the
attributes are omitted. The provider does not contact the server during
`terraform validate` or `terraform plan`; the login exchange happens on the first
API call.

Provider and resource documentation is generated into [`docs/`](docs/index.md).

## Development

```bash
make build             # build bin/terraform-provider-openmetadata
make test              # unit tests
make test-race         # unit tests with the race detector
make lint              # golangci-lint (config at .golangci.yml)
make vet               # go vet
make generate          # regenerate docs/ with tfplugindocs
make validate-examples # terraform fmt/validate every example
make release-check     # validate .github/goreleaser.yml
```

Acceptance tests need Docker Compose v2 (the stack uses health- and
completion-based dependencies). With Podman, point Docker Compose at the Podman
socket:

```bash
export DOCKER_HOST="unix://$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}')"
make testacc-docker    # postgres, Elasticsearch, migrate, server, then the suite
make testacc-down      # remove the stack and its volumes
```

The compose stack pins the server through `OPENMETADATA_VERSION`, which defaults
to the contents of `.github/openmetadata-version.txt`. The suite asserts the
server's `/system/version` matches that pin before running, so a stale image
fails at the precheck rather than in a confusing capability error.

## Releases

Releases are cut from the **Create Release Tag** workflow, which signs the tag
and invokes the **Release** workflow. GoReleaser builds linux, darwin and
windows binaries for amd64 and arm64, signs the checksum file with the
repository's GPG key, and publishes the Terraform registry manifest alongside
the archives. `make release-snapshot` produces the same platform set locally
without publishing.

## Repository layout

```
main.go                     provider entry point (providerserver.Serve)
internal/provider/          provider schema, configuration and registration
internal/providerdata/      typed data handed to resources
internal/client/            SDK wrapper, login, error classification
internal/convert/           shared conversion helpers
internal/acctest/           TF_ACC guards and acceptance helpers
internal/resources/<name>/  one package per resource, tests co-located
internal/datasources/...    one package per data source, tests co-located
examples/                   HCL examples rendered into docs/
templates/                  tfplugindocs templates
docs/                       generated provider documentation
```

## Documentation

- [Tasks and waves](TASKS.md) — every entity, its owning paths and the parallel
  work contract.
- [Contributing](CONTRIBUTING.md) — development workflow and review gates.
- [Agents](AGENTS.md) — the contract for agent-assisted entity work.
- [Security](SECURITY.md) — supported versions and reporting.
- [Changelog](CHANGELOG.md).
- [Generated provider docs](docs/index.md).

## Licence

Mozilla Public License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
