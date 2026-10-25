# Task breakdown

58 OpenMetadata entities, one work item each, sized so a single agent can land
one entity in one PR without touching another agent's files.

`scripts/entities.tsv` is the source of truth (entity, wave, SDK base path,
whether the SDK exposes `Delete`). `scripts/create-issues.sh` turns it into
GitHub issues and is idempotent, so re-running it never duplicates work.

```sh
scripts/create-issues.sh                      # dry run
scripts/create-issues.sh --apply              # create everything
scripts/create-issues.sh --only wave-1        # one wave
scripts/create-issues.sh --entity tag --apply # one entity

scripts/create-epics.sh                       # dry run
scripts/create-epics.sh --apply               # 10 epics + sub-issue links
```

Both scripts are idempotent and safe to re-run: they match on exact issue titles
and skip what already exists, so a partial failure is fixed by running them
again rather than cleaned up by hand.

## GitHub tracking

Each wave is an epic issue, and every entity issue is attached to its epic as a
real GitHub **sub-issue** (via the `addSubIssue` mutation, since `gh` has no
first-class command for it). Progress per wave is therefore visible in the
issue UI instead of being inferred from labels.

| Wave | Epic | Children |
|------|------|----------|
| 1 | #58 Taxonomy and governance | 6 |
| 2 | #59 Database stack | 6 |
| 3 | #60 BI stack | 4 |
| 4 | #61 Pipelines and messaging | 4 |
| 5 | #62 Platform services | 8 |
| 6 | #63 Drives and documents | 6 |
| 7 | #64 API assets | 3 |
| 8 | #65 Data quality and operations | 8 |
| 9 | #66 Access and organisation | 6 |
| 10 | #67 AI and learning | 6 |

The `wave-NN` labels remain on the children so issues can still be filtered
without going through the epic.

## Current state

| Item | Status |
|------|--------|
| Module, provider skeleton, client wrapper | done |
| `openmetadata_glossary` resource + data source (reference slice) | done |
| Docs generation from `examples/` + `templates/` | done |
| Unit tests (`go test ./...`) | green |
| Acceptance tests | CI only — no local Docker |
| GitHub issues | 57 entity issues + 10 wave epics |
| CI, GoReleaser, lint config, `README`, `LICENSE` | not yet written |

**The reference slice is `glossary`.** Every other entity copies it. Read
`internal/resources/glossary/resource.go` and
`internal/datasources/glossary/datasource.go` before starting; do not invent a
second architecture.

## Waves

Waves are dependency order, not importance order. Everything in wave 1 is
referenced by later resources through tags, classifications, domains, or data
products, so it lands first and shrinks merge conflicts.

| Wave | Theme | Entities | Why this order |
|------|-------|----------|----------------|
| 0 | Reference slice | glossary | Shipped. The pattern all others copy. |
| 1 | Taxonomy and governance | 6 | `classification`, `tag`, `glossary_term`, `domain`, `data_product`, `metric` are referenced by almost every later entity. |
| 2 | Database stack | 6 | `database_service` → `database` → `database_schema` → `table` → `stored_procedure`, plus `query`. Deepest hierarchy; validates the nested-lookup design. |
| 3 | BI stack | 4 | `dashboard_service`, `dashboard`, `dashboard_data_model`, `chart`. |
| 4 | Pipelines and messaging | 4 | `pipeline_service`, `pipeline`, `messaging_service`, `topic`. |
| 5 | Platform services | 8 | `storage_service`, `container`, `ml_model_service`, `ml_model`, `search_service`, `search_index`, `metadata_service`, `security_service`. Service-registration patterns shared here. |
| 6 | Drives and documents | 6 | `drive_service`, `directory`, `file`, `spreadsheet`, `worksheet`, `document`. Tree-shaped, like the database stack. |
| 7 | API assets | 3 | `api_service`, `api_collection`, `api_endpoint`. |
| 8 | Quality and operations | 8 | `test_definition`, `test_case`, `test_suite`, `data_contract`, `ingestion_pipeline`, `event_subscription`, `notification_template`, `workflow_definition`. Test cases point at tables, so this follows wave 2. |
| 9 | Access and organisation | 6 | `user`, `team`, `role`, `policy`, `bot`, `persona`. |
| 10 | AI and learning | 6 | `llm_service`, `llm_model`, `ai_application`, `ai_governance_policy`, `prompt_template`, `learning_resource`. Newest surface, least stable API. |

## Known blocker: entities without `Delete`

Six SDK services expose no `Delete` method:

`data_product`, `document`, `domain`, `event_subscription`, `persona`, `query`

A Terraform resource needs a working destroy. These issues are labelled
`sdk-blocker` and must resolve the delete path first — either a REST `DELETE`
the SDK omits, a soft delete via
`PATCH {"op":"add","path":"/deleted","value":true}`, or a narrow raw-HTTP helper
in `internal/client`. Shipping a no-op `Delete` is not acceptable: it orphans
metadata and silently breaks `terraform destroy`.

## Per-entity contract

An agent working one entity owns exactly these paths:

```
internal/resources/<entity>/        (resource.go, model.go, tests)
internal/datasources/<entity>/      (datasource.go, model.go, tests)
examples/resources/openmetadata_<entity>/
examples/data-sources/openmetadata_<entity>/
docs/                               (regenerated, not hand-edited)
internal/provider/provider.go       (one alphabetical line, one slice)
```

Rules that keep parallel PRs mergeable:

1. **One insertion in `provider.go`.** Add exactly one line to `Resources()` and
   one to `DataSources()`, in alphabetical position. Do not reorder other lines,
   do not reformat the file.
2. **No shared-file edits.** `internal/client`, `internal/convert`,
   `internal/providerdata`, `internal/acctest`, `go.mod`, and `GNUmakefile` are
   off limits. If you need a change there, open a separate issue.
3. **Acceptance tests in an external package.** `package <entity>_test`.
   Importing `internal/acctest` from the internal package creates an import
   cycle through `internal/provider`.
4. **`Default` needs `Computed`.** Terraform rejects an attribute `Default` on a
   non-computed attribute.
5. **Patch, don't replace.** The SDK has no `Update`; send only changed fields
   through `ometa.JSONPatchOp` so fields managed outside Terraform survive.
6. **Regenerate docs.** `go generate ./...`; never hand-edit `docs/`.

## Definition of done

```sh
go build ./...
go vet ./...
go test ./...
go generate ./...
git diff --exit-code docs
```

Plus a green acceptance run (CI only) and a PR that touches no file outside the
list above.
