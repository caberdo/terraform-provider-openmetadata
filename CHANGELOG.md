# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Module, provider configuration and the `openmetadata_glossary` resource and
  data source, the reference slice for every other entity.
- Wave 1 taxonomy and governance: `classification`, `tag`, `glossary_term`,
  `domain`, `data_product` and `metric` resources and data sources.
- Wave 2 database stack: `database_service`, `database`, `database_schema`,
  `table`, `stored_procedure` and `query` resources and data sources, including
  nested fully-qualified-name handling from service down to columns.
- `openmetadata_query` uses the REST `DELETE /api/v1/queries/{id}` endpoint via
  the raw delete helper because the SDK exposes no `Delete`; destroy is a soft
  delete on the server.
- Wave 3 BI stack: `chart`, `dashboard`, `dashboard_data_model` and
  `dashboard_service` resources and data sources, including dashboard/chart
  relationships, data model columns and JSON connector configuration.
- Wave 4 pipelines and messaging: `pipeline_service`, `pipeline`,
  `messaging_service` and `topic` resources and data sources, including pipeline
  tasks and topic message schemas, retention and cleanup policies.
- A narrow raw `DELETE` helper in `internal/client` for domains and data
  products. The Go SDK exposes no `Delete` for those services even though the
  REST API supports a permanent delete.
- Build tooling: `GNUmakefile`, `.golangci.yml` and generated documentation from
  `examples/` and `templates/`.
- CI: lint, race-enabled unit tests, cross-platform build, generated-doc drift,
  example validation and the acceptance suite in Docker Compose.
- Acceptance stack: postgres, Elasticsearch, one-shot migration and the
  OpenMetadata server, pinned by `.github/openmetadata-version.txt`.
- Release pipeline: GPG-signed GoReleaser archives for linux, darwin and windows
  on amd64 and arm64, registry manifest, build provenance and SBOM.
- Community files: README, AGENTS, CONTRIBUTING, SECURITY, CODE_OF_CONDUCT,
  LICENSE, NOTICE, dependabot, CODEOWNERS and pull-request/issue templates.

### Fixed

- Username/password login now Base64-encodes the password, as the OpenMetadata
  login endpoint requires.
- `openmetadata_glossary.mutually_exclusive` forces replacement instead of a
  patch: OpenMetadata keeps the value set at creation, so an in-place update
  produced an inconsistent result.
