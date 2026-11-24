# Terraform Provider for OpenMetadata

[![CI](https://img.shields.io/github/actions/workflow/status/caberdo/terraform-provider-openmetadata/test.yml?branch=main&label=CI)](https://github.com/caberdo/terraform-provider-openmetadata/actions/workflows/test.yml)
<!-- coverage-badge:start -->
[![Coverage floor](https://img.shields.io/badge/coverage%20floor-33%25-green)](#testing-and-coverage)
<!-- coverage-badge:end -->
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/caberdo/terraform-provider-openmetadata/badge)](https://scorecard.dev/viewer/?uri=github.com/caberdo/terraform-provider-openmetadata)
[![Release](https://img.shields.io/github/v/release/caberdo/terraform-provider-openmetadata?label=release)](https://github.com/caberdo/terraform-provider-openmetadata/releases)
[![Terraform Registry](https://img.shields.io/badge/registry-caberdo%2Fopenmetadata-7b42bc)](https://registry.terraform.io/providers/caberdo/openmetadata)
[![OpenTofu](https://img.shields.io/badge/OpenTofu-available-ffda18)](https://search.opentofu.org/provider/caberdo/openmetadata)
[![Go](https://img.shields.io/github/go-mod/go-version/caberdo/terraform-provider-openmetadata)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/github/license/caberdo/terraform-provider-openmetadata)](LICENSE)
[![Last commit](https://img.shields.io/github/last-commit/caberdo/terraform-provider-openmetadata)](https://github.com/caberdo/terraform-provider-openmetadata/commits/main)

Manage the **[OpenMetadata](https://open-metadata.org/)** metadata platform as code. The provider
covers glossaries, classifications, tags, domains, services and the entities that hang off them —
databases, tables, dashboards, pipelines, topics, drives, API assets, quality objects, access
objects and AI assets — through the OpenMetadata REST API.

**58 resources and 58 data sources.**

| | |
|---|---|
| 🎯 **Broad OpenMetadata coverage** | 58 resources and 58 matching data sources across ten waves: the glossary reference slice, taxonomy and governance, the database stack, BI, pipelines and messaging, platform services, drives and documents, API assets, quality and operations, access and organisation, and AI and learning |
| 🧪 **Tested against a real server** | Unit tests for provider configuration, the login exchange, error classification, conversion helpers and every entity's schema and patch document, plus an acceptance suite that drives full CRUD, update, destroy and both import identifiers against OpenMetadata in the compose stack |
| 🔁 **Patch, don't replace** | Updates send only changed fields through a JSON patch, so fields managed outside Terraform survive; a field the server cannot patch carries `RequiresReplace` instead of being silently dropped |
| 🧭 **The server stays the source of truth** | Every read is an API call, an attribute the server does not return is preserved rather than zeroed, and a read that gets a 404 removes the resource from state instead of failing |
| ⚙️ **Modern internals** | terraform-plugin-framework v1.19.0 (plugin protocol 6.0), the official OpenMetadata Go SDK 1.13.0, Go 1.27.4 |

> **Status:** early development. The module, tooling, CI, acceptance stack and release path are in
> place, and all ten waves in [TASKS.md](TASKS.md) are implemented: 58 resources and 58 matching
> data sources, from the glossary reference slice through taxonomy, databases, BI, pipelines,
> platform services, drives, API assets, quality, access and AI. The API surface still moves with
> the OpenMetadata server, so pin the provider version.

## Contents

- [Quick start](#quick-start)
- [Requirements and compatibility](#requirements-and-compatibility)
- [Repository layout](#repository-layout)
- [Provider configuration](#provider-configuration)
- [Resources](#resources)
- [Data sources](#data-sources)
- [Examples](#examples)
- [Verifying a release](#verifying-a-release)
- [Testing and coverage](#testing-and-coverage)
- [Development](#development)
- [Troubleshooting and FAQ](#troubleshooting-and-faq)
- [Security, contributing and support](#security-contributing-and-support)
- [License](#license)

## Quick start

```terraform
terraform {
  required_providers {
    openmetadata = {
      source  = "caberdo/openmetadata"
      version = "~> 0.2"
    }
  }
}

provider "openmetadata" {
  host = "https://openmetadata.example.com"
}

resource "openmetadata_glossary" "finance" {
  name               = "finance"
  display_name       = "Finance"
  description        = "Glossary owned by the finance data team."
  mutually_exclusive = false
}
```

```bash
terraform init
terraform apply
```

Credentials are read from `OPENMETADATA_TOKEN`, or from `OPENMETADATA_USERNAME` together with
`OPENMETADATA_PASSWORD`, when the attributes are omitted. The provider does not contact the server
during `terraform validate` or `terraform plan`; the login exchange happens on the first API call.

## Requirements and compatibility

| | |
|---|---|
| Terraform | 1.9 or later (plugin protocol 6.0); also available for [OpenTofu](https://search.opentofu.org/provider/caberdo/openmetadata) |
| OpenMetadata | **1.13** or later. The acceptance window lives in [`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt) — currently `1.13.6`, `2.0.5`, `2.0.6`, `2.0.7`, `2.0.8` and `2.0.9` |
| Go (to build from source) | 1.27.4 |

### Supported OpenMetadata versions

- The minimum supported server is **OpenMetadata 1.13.6**. One provider build supports 1.13.6 and
  newer; a server outside the window is untested rather than rejected, because the provider does
  not enforce a server version up front and surfaces the server's own error when an endpoint it
  needs is missing.
- [`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt) is the single source of
  truth for the CI matrix and `make testacc-matrix`. It starts at the declared floor and lists the
  last stable releases — currently `1.13.6`, `2.0.5`, `2.0.6`, `2.0.7`, `2.0.8` and `2.0.9`.
  A pull request runs the floor and the newest release; `main` and the merge queue run the whole
  window. A weekly workflow opens an issue when the window goes stale.
- A plain `make testacc-docker` runs the newest version in the file; set
  `OPENMETADATA_VERSION=2.0.8` to pin another one for a single run.
- The acceptance precheck compares the server's `/system/version` with the selected version before
  running, so a stale image fails at the precheck rather than in a confusing capability error.

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

## Provider configuration

| Provider attribute | Environment variable | Description |
|---|---|---|
| `host` | `OPENMETADATA_HOST` | OpenMetadata server root including the scheme, for example `https://openmetadata.example.com` |
| `token` | `OPENMETADATA_TOKEN` | JWT or personal access token sent as a bearer token; takes precedence over `username` and `password` |
| `username` | `OPENMETADATA_USERNAME` | User email used to obtain a JWT when `token` is unset |
| `password` | `OPENMETADATA_PASSWORD` | Password for `username` (sensitive) |
| `api_version` | `OPENMETADATA_API_VERSION` | OpenMetadata REST API version; defaults to `v1` |
| `retry_count` | `OPENMETADATA_RETRY_COUNT` | Retries after the initial attempt for retryable responses (429 and 504); defaults to 3 |
| `retry_wait` | `OPENMETADATA_RETRY_WAIT` | Delay between retries as a Go duration, for example `30s`; defaults to 30s |

A value set in the provider block wins over its environment variable, which is the fallback used
when the attribute is omitted. A missing host fails `terraform plan` with an explicit diagnostic, and
so does a missing credential unless `token` or both `username` and `password` are set.

## Resources

Every resource links to its documentation on the Terraform Registry from [`docs/resources/`](docs/resources).
Their type names follow the OpenMetadata entity they manage:

| Group | Resources |
|---|---|
| Glossary and taxonomy | `openmetadata_glossary`, `openmetadata_glossary_term`, `openmetadata_classification`, `openmetadata_tag`, `openmetadata_domain`, `openmetadata_data_product`, `openmetadata_metric` |
| Database stack | `openmetadata_database_service`, `openmetadata_database`, `openmetadata_database_schema`, `openmetadata_table`, `openmetadata_stored_procedure`, `openmetadata_query` |
| BI stack | `openmetadata_dashboard_service`, `openmetadata_dashboard`, `openmetadata_dashboard_data_model`, `openmetadata_chart` |
| Pipelines and messaging | `openmetadata_pipeline_service`, `openmetadata_pipeline`, `openmetadata_messaging_service`, `openmetadata_topic` |
| Platform services | `openmetadata_storage_service`, `openmetadata_container`, `openmetadata_ml_model_service`, `openmetadata_ml_model`, `openmetadata_search_service`, `openmetadata_search_index`, `openmetadata_metadata_service`, `openmetadata_security_service` |
| Drives and documents | `openmetadata_drive_service`, `openmetadata_directory`, `openmetadata_file`, `openmetadata_spreadsheet`, `openmetadata_worksheet`, `openmetadata_document` |
| API assets | `openmetadata_api_service`, `openmetadata_api_collection`, `openmetadata_api_endpoint` |
| Quality and operations | `openmetadata_test_definition`, `openmetadata_test_case`, `openmetadata_test_suite`, `openmetadata_data_contract`, `openmetadata_ingestion_pipeline`, `openmetadata_event_subscription`, `openmetadata_notification_template`, `openmetadata_workflow_definition` |
| Access and organisation | `openmetadata_user`, `openmetadata_team`, `openmetadata_policy`, `openmetadata_role`, `openmetadata_bot`, `openmetadata_persona` |
| AI and learning | `openmetadata_ai_application`, `openmetadata_ai_governance_policy`, `openmetadata_learning_resource`, `openmetadata_llm_model`, `openmetadata_llm_service`, `openmetadata_prompt_template` |

Every resource creates, reads, updates and deletes through the OpenMetadata REST API. Updates are
JSON patches, not replacements, so a field another team manages outside Terraform is left alone.

## Data sources

Each of the 58 resources has a matching data source with the same type name — `openmetadata_glossary`
has `data.openmetadata_glossary`, and so on — so a configuration can look an entity up without
managing it. Every data source is read-only, takes the entity's identifying attributes and lives
under [`docs/data-sources/`](docs/data-sources).

## Examples

| Path | What it shows |
|---|---|
| [`examples/resources/<resource>`](examples/resources) | One `resource.tf` per resource, used to generate the docs |
| [`examples/data-sources/<data source>`](examples/data-sources) | One `data-source.tf` per data source |
| [`examples/provider`](examples/provider) | Provider configuration |

`make validate-examples` checks every example against the provider schema in CI.

## Verifying a release

Releases are cut after the release pull request merges by pushing a GPG-signed `v*` tag, or by
dispatching **Release** with an existing tag to republish it. The kit's tag job is skipped without
a version file, so no workflow invents a version on `main`.
GoReleaser then builds linux, darwin and windows binaries for amd64 and arm64 from the
GPG-signed tag and publishes:

| Asset | Purpose |
|---|---|
| `terraform-provider-openmetadata_<version>_<os>_<arch>.zip` | The provider binaries |
| `terraform-provider-openmetadata_<version>_SHA256SUMS` | Checksums for all artifacts |
| `terraform-provider-openmetadata_<version>_SHA256SUMS.sig` | GPG signature over the checksums |
| `terraform-provider-openmetadata_<version>_manifest.json` | Terraform registry manifest (plugin protocol versions) |

The release workflow also attests build provenance for the artifacts and emits an SBOM. Replace
`<version>` with the release you are verifying and check where an archive came from with the GitHub
CLI:

```bash
gh attestation verify terraform-provider-openmetadata_<version>_linux_amd64.zip \
    --repo caberdo/terraform-provider-openmetadata
```

That answers which repository, tag, commit and workflow produced the archive, and it fails if the
file was rebuilt or modified after the fact. The public half of the release signing key is not
committed to this repository, so the attestation record — not a key import — is the check that
needs no out-of-band material. Terraform itself verifies the checksums when it installs the
provider and records them in `.terraform.lock.hcl`; commit that file so every `terraform init`
re-checks the same artifacts.

`make release-snapshot` produces the same platform set locally without publishing or signing.

## Testing and coverage

| Command | What it runs |
|---|---|
| `make test` | Unit tests (`TestAcc*` skips itself without `TF_ACC`) |
| `make test-race` | Unit tests with the race detector |
| `make coverage` | Unit coverage profile (`coverage.out`) and a per-function summary |
| `make coverage-html` | The coverage profile as a browsable HTML report |
| `make coverage-gate` | Fails below the module floor or a per-package floor |
| `make coverage-readme-check` | Fails when the README coverage block has drifted from the baseline |
| `make testacc-docker` | The acceptance suite against the newest OpenMetadata release in the compose stack |
| `make testacc-matrix` | The acceptance suite against every version in the supported window |
| `make coverage-merge` | Merges the unit and acceptance profiles into one statement-coverage table |

The suite is layered: pure unit tests cover provider configuration, the login exchange, token
shapes, error classification and conversion helpers; entity tests cover schema shape, model
mapping and patch-document construction; the acceptance layer runs real CRUD, update, destroy and
both import identifiers against OpenMetadata in the compose stack.

`make coverage` writes the unit profile to `coverage.out`; `make coverage-gate` enforces both the
module floor in `.github/coverage-baseline.txt` and the per-package floors in
`.github/coverage-floors.txt`. Acceptance coverage is merged with the unit profile by
`make coverage-merge` and reported by CI, but it is not gated because it needs a live server. A
package with no tests measures 0% and fails the gate, so a new entity package is gated the moment
it exists.

<!-- coverage:start -->
The coverage gate fails below **33%** (`make coverage-gate`).
<!-- coverage:end -->

## Development

```bash
make build              # build bin/terraform-provider-openmetadata
make install            # install into ~/.terraform.d/plugins for local dev_overrides
make test coverage      # unit tests + coverage profile
make test-race          # unit tests with the race detector
make coverage-readme    # regenerate the README coverage badge and floor statement
make lint vet fmt       # golangci-lint (.golangci.yml), go vet, gofmt
make generate           # regenerate docs/ with tfplugindocs
make validate-examples  # check examples/ against the provider schema
make release-check      # validate .github/goreleaser.yml
make release-snapshot   # build every release platform into dist/
```

Acceptance tests need Docker Compose v2 (the stack uses health- and completion-based
dependencies). With Podman, point Docker Compose at the Podman socket and override `COMPOSE`:

```bash
export DOCKER_HOST="unix://$(podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}')"
make testacc-docker COMPOSE=podman-compose   # postgres, Elasticsearch, migrate, server, then the suite
make testacc-matrix COMPOSE=podman-compose   # the same suite for every supported version
make testacc-down COMPOSE=podman-compose     # remove the stack and its volumes
```

The compose stack selects the server through `OPENMETADATA_VERSION`, which defaults to the newest
entry in `.github/openmetadata-versions.txt`; `make testacc-matrix` walks the whole file. The suite
asserts the server's `/system/version` matches that version before running, so a stale image fails
at the precheck rather than in a confusing capability error.

To use a local build with Terraform, add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "caberdo/openmetadata" = "/path/to/terraform-provider-openmetadata/bin"
  }
}
```

`main` is protected: every change lands through a pull request with green CI, and
[`AGENTS.md`](AGENTS.md) describes the repository conventions (the parallel work contract, entity
packages, patch semantics and documentation rules).

## Troubleshooting and FAQ

**`terraform plan` fails with "Missing OpenMetadata host" or "Missing OpenMetadata credentials".**
Set `host` in the provider block or `OPENMETADATA_HOST`, and set either `token` or both `username`
and `password` (or the matching `OPENMETADATA_*` variables). The provider does not guess a default
host.

**Does the provider contact the server during `terraform validate` or `terraform plan`?**
No. Authentication is lazy: the login exchange happens on the first API call, which is why
validation works without a reachable server.

**A request fails with 429 or 504.**
Those are the retryable responses. The provider retries `retry_count` times (default 3) with
`retry_wait` between attempts (default `30s`); raise either when the server is under load.

**A resource vanished from state after a refresh.**
A read that gets a 404 removes the resource from state, and a delete of an object that is already
gone is treated as success. That is deliberate: an object removed out-of-band does not fail the
plan.

**Changing an attribute forces a replacement.**
Some fields are fixed at creation — the glossary's `mutually_exclusive` is the canonical example,
because the server silently keeps the old value. Those fields carry `RequiresReplace` instead of
being patched, so the change is visible in the plan rather than failing with "Provider produced
inconsistent result after apply".

**`make testacc-docker` fails without Docker.**
Install Docker Compose v2, or use Podman: `make testacc-docker COMPOSE=podman-compose` with Podman
4.6+ and podman-compose 1.6+, which implement the health and completion conditions the stack uses.

**The acceptance precheck fails on a version mismatch.**
The suite compares the server's `/system/version` with `OPENMETADATA_EXPECT_VERSION`, which the
compose file sets from the selected entry in `.github/openmetadata-versions.txt`. Rebuild the stack
after changing the version instead of mixing a stale image with a new expectation.

**CI fails with docs drift.**
Run `make generate` and commit the result. `docs/` is generated from `examples/` and `templates/`
and is never hand-edited; the docs job also catches a new untracked file that `git diff` misses.

**Can I use OpenTofu?**
Yes — this is a plugin-protocol 6.0 provider and it is listed in the
[OpenTofu registry](https://search.opentofu.org/provider/caberdo/openmetadata) as well.

## Security, contributing and support

- [Generated provider docs](docs/index.md) — every resource and data source page, rendered from
  [`examples/`](examples) and [`templates/`](templates) by `make generate`.
- [`SECURITY.md`](SECURITY.md) — supported versions, the threat model and how to report a
  vulnerability privately.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — development setup, acceptance tests, review gates and the
  release procedure.
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) — the ground rules for the community.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — request flow, client, patterns, test layers and releases.
- [`TASKS.md`](TASKS.md) — the entity list, wave order and the parallel work contract.
- [`AGENTS.md`](AGENTS.md) — the contract for agent-assisted entity work.
- [`CHANGELOG.md`](CHANGELOG.md) — what changed in every release.
- [`.github/ISSUE_TEMPLATE`](.github/ISSUE_TEMPLATE) — bug reports and feature requests; include
  the provider version, Terraform version, OpenMetadata version and the authentication mode, never
  the credential itself.
- [`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md) — the checklist every pull
  request is reviewed against.

## License

MIT — see the [LICENSE](LICENSE) file, and [`NOTICE`](./NOTICE) for the upstream
works this project derives from and the licence carve-outs that apply to them.
