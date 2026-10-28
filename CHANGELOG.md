# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.1.0] - 2026-10-28

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
- Wave 5 platform services: `storage_service`, `container`,
  `ml_model_service`, `ml_model`, `search_service`, `search_index`,
  `metadata_service` and `security_service` resources and data sources,
  including container parent resolution and search index field management.
- Wave 6 drives and documents: `drive_service`, `directory`, `file`,
  `spreadsheet`, `worksheet` and `document` resources and data sources,
  including directory/file parent resolution and worksheet spreadsheet
  derivation.
- Wave 7 API assets: `api_service`, `api_collection` and `api_endpoint`
  resources and data sources, including collection/endpoint fully-qualified-name
  derivation and sensitive JSON connector configuration.
- Wave 8 data quality and operations: `test_definition`, `test_case`,
  `test_suite`, `data_contract`, `ingestion_pipeline`, `event_subscription`,
  `notification_template` and `workflow_definition` resources and data sources,
  including table and service UUID references for test cases and ingestion
  pipelines, and the default `noOp` workflow trigger.
- Wave 9 access and organisation: `user`, `team`, `policy`, `role`, `bot` and
  `persona` resources and data sources, including bot-user JWT provisioning,
  team hierarchy and default roles, policy rules, and the reference
  relationships between users, teams, roles, policies and personas. Reference
  collections are sets because OpenMetadata returns them sorted.
- `openmetadata_persona` uses the REST `DELETE /api/v1/personas/{id}` endpoint
  via the raw delete helper because the SDK exposes no `Delete`; destroy is a
  hard delete on the server.
- `openmetadata_event_subscription` uses the REST
  `DELETE /api/v1/events/subscriptions/{id}` endpoint with `hardDelete` and
  `recursive` because the SDK exposes no `Delete`; destroy is a hard delete on
  the server.
- `openmetadata_document` uses the REST `DELETE /api/v1/docStore/{id}` endpoint
  via the raw delete helper because the SDK exposes no `Delete`; destroy is a
  hard delete on the server. Its `data` attribute is canonicalized JSON so
  `jsonencode()` configuration does not drift.
- A narrow raw `DELETE` helper in `internal/client` for domains and data
  products. The Go SDK exposes no `Delete` for those services even though the
  REST API supports a permanent delete.
- Build tooling: `GNUmakefile`, `.golangci.yml` and generated documentation from
  `examples/` and `templates/`.
- CI: lint, race-enabled unit tests, cross-platform build, generated-doc drift,
  example validation and the acceptance suite in Docker Compose.
- Acceptance stack: postgres, Elasticsearch, one-shot migration and the
  OpenMetadata server, pinned by `.github/openmetadata-versions.txt`.
- Release pipeline: GPG-signed GoReleaser archives for linux, darwin and windows
  on amd64 and arm64, registry manifest, build provenance and SBOM.
- Community files: README, AGENTS, CONTRIBUTING, SECURITY, CODE_OF_CONDUCT,
  ARCHITECTURE, LICENSE, NOTICE, dependabot, CODEOWNERS and pull-request/issue
  templates.
- Coverage tooling ported from the trino provider: `scripts/coverage-summary.sh`,
  `scripts/merge-coverage.sh`, `scripts/coverage-gate.sh` and
  `scripts/coverage-readme.sh`, with the module baseline and per-package floors
  in `.github/coverage-baseline.txt` and `.github/coverage-floors.txt`. The unit
  gate runs in the `Test` job, the acceptance container writes
  `coverage-acceptance.out` for the combined report, and the README badge is
  checked against the baseline.

### Fixed

- Username/password login now Base64-encodes the password, as the OpenMetadata
  login endpoint requires.
- `openmetadata_glossary.mutually_exclusive` forces replacement instead of a
  patch: OpenMetadata keeps the value set at creation, so an in-place update
  produced an inconsistent result.

[0.1.0]: https://github.com/caberdo/terraform-provider-openmetadata/releases/tag/v0.1.0
