# AGENTS.md

Guidance for coding agents working in this repository.

## Read these before changing behaviour

1. [TASKS.md](TASKS.md) — the entity list, wave order and the parallel work
   contract. It is the source of truth for what one task may touch.
2. [CONTRIBUTING.md](CONTRIBUTING.md) — commands, review gates and the pull
   request process.
3. The reference slice:
   `internal/resources/glossary/resource.go`,
   `internal/resources/glossary/model.go`,
   `internal/datasources/glossary/datasource.go` and their tests.

Every entity copies the glossary slice. Do not invent a second architecture.

## Package layout

```
main.go                     provider entry point (providerserver.Serve)
internal/provider/          provider schema, configuration and registration
internal/providerdata/      typed data handed to resources
internal/client/            SDK wrapper, login exchange, error classification
internal/convert/           shared conversion helpers
internal/acctest/           TF_ACC guards and acceptance helpers
internal/resources/<name>/  one package per resource, tests co-located
internal/datasources/...    one package per data source, tests co-located
examples/                   HCL examples rendered into docs/
templates/                  tfplugindocs templates
docs/                       generated provider documentation
```

The per-entity packages are what make parallel work safe: no shared files, and
tests travel with the code they cover. Do not create a flat
`internal/provider/resource_*.go` layout.

## Parallel work contract

An agent working one entity owns exactly these paths:

```
internal/resources/<entity>/        (resource.go, model.go, tests)
internal/datasources/<entity>/      (datasource.go, model.go, tests)
examples/resources/openmetadata_<entity>/
examples/data-sources/openmetadata_<entity>/
docs/                               (regenerated, not hand-edited)
internal/provider/provider.go       (one alphabetical line, one slice)
```

Rules that keep parallel pull requests mergeable:

1. **One insertion in `provider.go`.** Add exactly one line to `Resources()` and
   one to `DataSources()`, in alphabetical position. Do not reorder or reformat
   other lines.
2. **No shared-file edits.** `internal/client`, `internal/convert`,
   `internal/providerdata`, `internal/acctest`, `go.mod`, `go.sum`, CI files and
   `GNUmakefile` are off limits. If you need a change there, open a separate
   issue.
3. **Acceptance tests in an external package.** Use `package <entity>_test`.
   Importing `internal/acctest` from the internal package creates an import
   cycle through `internal/provider`.
4. **`Default` needs `Computed`.** Terraform rejects an attribute `Default` on a
   non-computed attribute.
5. **Patch, don't replace.** The SDK exposes no `Update`; send only changed
   fields through `ometa.JSONPatchOp` so fields managed outside Terraform
   survive.
6. **Replace what the server cannot patch.** Some fields are fixed at creation
   (the glossary's `mutually_exclusive`, for example: the server silently keeps
   the old value). Give those `RequiresReplace()` and do not emit a patch op;
   otherwise Terraform fails with "Provider produced inconsistent result after
   apply".
7. **Regenerate docs.** Run `go generate ./...`; never hand-edit `docs/`.
8. **Read preserves state.** Never zero an attribute the server does not return.

## Commands

| Command | Purpose |
|---|---|
| `make build` | build `bin/terraform-provider-openmetadata` |
| `make test` | unit tests |
| `make test-race` | unit tests with the race detector |
| `make lint` | golangci-lint (config at `.golangci.yml`) |
| `make fmt` / `make vet` | formatting and vet |
| `make validate-examples` | `terraform fmt`/`validate` every example |
| `make generate` | regenerate `docs/` with `tfplugindocs` |
| `make testacc-docker` | acceptance suite against OpenMetadata in Docker |
| `make testacc-down` | remove the acceptance stack and its volumes |
| `make release-check` | validate the GoReleaser configuration |
| `make release-snapshot` | build every release platform into `dist/` |

## Rules

- Only the lead edits `go.mod`, `go.sum`, CI files and files under
  `internal/provider/` other than the two registration insertions described
  above. Entity work stays inside its owned paths; `docs/` is regenerated, never
  hand-edited.
- No new third-party dependency without the lead's approval.
- No `panic`, `log.Fatal` or `os.Exit` outside `main`.
- No secret in a diagnostic, log field, test failure message or committed file.
- Docs under `docs/` are generated. Never hand-edit them.
- Acceptance tests must compile even when they are not run locally, and they
  must pass in CI before a CRUD change merges.

## Review gates

A change is ready when `make build`, `make vet`, `make lint` and the relevant
`go test` scope pass, `make generate` leaves `docs/` unchanged, examples validate
when they changed, the `Acceptance` job is green for CRUD changes, and no write
escaped the task's owned paths.
