# Terraform Provider for OpenMetadata

A Terraform provider for the [OpenMetadata](https://open-metadata.org/) metadata
platform. It manages glossaries, classifications, tags, services and the entities
that hang off them through the OpenMetadata REST API.

The provider is a thin translation layer between Terraform state and
OpenMetadata. It owns no truth of its own: every read is an API call, every
mutation is a create, patch or delete, and attributes the server cannot return
are preserved rather than zeroed.

> **Status:** early development. The module, tooling, CI, acceptance stack and
> release path are in place, and all ten waves in [TASKS.md](TASKS.md) are
> implemented: 58 resources and 58 matching data sources, from the glossary
> reference slice through taxonomy, databases, BI, pipelines, platform
> services, drives, API assets, quality, access and AI. The API surface still
> moves with the OpenMetadata server, so pin the provider version.

## Requirements

- Terraform 1.9 or later (plugin protocol 6.0).
- OpenMetadata **1.13** or later. CI runs the acceptance suite against the
  version pinned in
  [`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt).

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
make coverage          # unit coverage profile and summary (coverage.out)
make coverage-gate     # fail below .github/coverage-baseline.txt or a package floor
make coverage-readme   # rewrite the coverage badge from the baseline
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
to the contents of `.github/openmetadata-versions.txt`. The suite asserts the
server's `/system/version` matches that pin before running, so a stale image
fails at the precheck rather than in a confusing capability error.

## Coverage

<!-- coverage:start -->
![Coverage floor](https://img.shields.io/badge/coverage%20floor-33%25-green)

The coverage gate fails below **33%** (`./scripts/coverage-gate.sh`).
<!-- coverage:end -->

`make coverage` writes the unit profile to `coverage.out`; `make coverage-gate`
enforces both the module floor in `.github/coverage-baseline.txt` and the
per-package floors in `.github/coverage-floors.txt`. Acceptance coverage is
merged with the unit profile by `make coverage-merge` and reported by CI, but it
is not gated because it needs a live server.

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
