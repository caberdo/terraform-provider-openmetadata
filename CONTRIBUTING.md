# Contributing

Thanks for contributing to the Terraform provider for OpenMetadata.

## Before you start

Read, in this order:

1. [TASKS.md](TASKS.md) — the entity list, wave order and the per-entity
   contract. One entity per pull request.
2. [AGENTS.md](AGENTS.md) — the package layout and the rules that keep parallel
   work mergeable.
3. The reference slice under `internal/resources/glossary/` and
   `internal/datasources/glossary/`.

The glossary slice is the architecture. If a later entity needs a shared
interface that does not exist yet, open a separate issue against the shared
packages instead of editing them in an entity pull request.

## Development setup

Requirements:

- Go 1.25.8 (the version in `go.mod`).
- Terraform 1.9 or later for example validation.
- Docker Compose v2 for acceptance tests. Podman works when Docker Compose
  targets the Podman socket (`DOCKER_HOST=unix://…`).

```bash
make build
make test
make lint
make vet
make validate-examples
make generate
```

## Pull requests

- One logical change per pull request; keep the diff inside the paths the task
  owns.
- Add or update tests with the change. Unit tests live next to the code;
  acceptance tests live in `package <entity>_test` and import `internal/acctest`.
- Do not hand-edit `docs/`; run `make generate` and commit the result.
- `go.mod`/`go.sum` changes need the lead's approval.
- Fill in the scope / owned paths section of the pull request template.
- CI must be green: `Lint`, `Test` (build, vet, cross-platform build, generated
  docs diff, example validation and the race-enabled unit suite) and
  `Acceptance` (the compose stack against the pinned OpenMetadata version).

`main` should be protected with those three contexts required, a linear history
and no force pushes. A job `name:` is the required context, so renaming one is a
breaking CI change: update `.github/workflows/test.yml` and the repository
protection together.

## Release

Dispatch **Create Release Tag** from the Actions tab. It calculates or accepts a
version, creates a GPG-signed tag, and calls the **Release** workflow, which runs
GoReleaser with signed checksums, the registry manifest, build provenance and an
SBOM. To republish an existing tag, dispatch **Release** with the tag instead of
re-running the tag workflow.

## Commits

Use conventional commit prefixes:

- `feat(<area>): ...` for user-visible behaviour
- `fix(<area>): ...` for bug fixes
- `chore(<area>): ...` for tooling, CI and repository hygiene
