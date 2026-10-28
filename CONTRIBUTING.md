# Contributing

Thanks for contributing to the Terraform provider for OpenMetadata.

## Before you start

Read, in this order:

1. [TASKS.md](TASKS.md) — the entity list, wave order and the per-entity
   contract. One entity per pull request.
2. [AGENTS.md](AGENTS.md) — the package layout and the rules that keep parallel
   work mergeable.
3. [ARCHITECTURE.md](ARCHITECTURE.md) — the request flow, client, resource
   pattern and test layers.
4. The reference slice under `internal/resources/glossary/` and
   `internal/datasources/glossary/`.

The glossary slice is the architecture. If a later entity needs a shared
interface that does not exist yet, open a separate issue against the shared
packages instead of editing them in an entity pull request.

## Development setup

Requirements:

- Go 1.25.8 (the version in `go.mod`).
- Terraform 1.9 or later for example validation.
- Docker Compose v2 for acceptance tests. Podman works either through
  `COMPOSE=podman-compose` or by pointing Docker Compose at the Podman socket
  (`DOCKER_HOST=unix://…`).

```bash
make build
make test
make lint
make vet
make validate-examples
make generate
```

## Acceptance tests

Acceptance tests run against a real OpenMetadata server in Docker Compose. The
stack is PostgreSQL, Elasticsearch, a one-shot schema migration and the
OpenMetadata server; the server version is pinned once in
[`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt) and
passed to the images and to `OPENMETADATA_EXPECT_VERSION`.

```bash
make testacc-docker                          # docker compose
make testacc-docker COMPOSE=podman-compose   # Podman
make testacc-down                            # remove the stack and volumes
```

With Docker Compose v2 and Podman, either point Compose at the Podman socket
(`DOCKER_HOST=unix://…`) or pass `COMPOSE=podman-compose`. The stack uses
health- and completion-based dependencies, so Compose v2 or podman-compose 1.6+
is required.

To run from the host against the published port instead of the `test`
container:

```bash
export OPENMETADATA_VERSION=$(tr -d '[:space:]' < .github/openmetadata-versions.txt)
docker compose up -d postgresql elasticsearch
docker compose run --rm migrate
docker compose up -d openmetadata-server

TF_ACC=1 OPENMETADATA_HOST=http://127.0.0.1:8585 \
  OPENMETADATA_USERNAME=admin@open-metadata.org \
  OPENMETADATA_PASSWORD=admin \
  OPENMETADATA_EXPECT_VERSION=$OPENMETADATA_VERSION \
  go test -count=1 -timeout 30m ./internal/...

docker compose down -v
```

`PreCheck` probes `/system/version` and fails when the server does not match the
pin, so a stale image cannot masquerade as a green run. Acceptance tests must
not skip once `TF_ACC` is set: an unreachable server or a missing capability is
an environment failure, not a pass. They must compile under `go vet ./...` even
when they are not run locally.

**Changing the supported server:** edit
`.github/openmetadata-versions.txt`, update the `OpenMetadata 1.13` floor in
`README.md` and `SECURITY.md` if it moved, and run `make testacc-docker`. Never
hand-edit `docker-compose.yml` image tags or the workflow's expected version;
both read the pin.

## Pull requests

- One logical change per pull request; keep the diff inside the paths the task
  owns.
- Add or update tests with the change. Unit tests live next to the code;
  acceptance tests live in `package <entity>_test` and import `internal/acctest`.
- Do not hand-edit `docs/`; run `make generate` and commit the result.
- `go.mod`/`go.sum` changes need the lead's approval.
- Fill in the scope / owned paths section of the pull request template.

CI must be green. The required contexts on `main` are:

| Context | Job | Covers |
|---|---|---|
| `Lint & Security` | `lint` | golangci-lint and gosec |
| `Test` | `test` | build, vet, `go mod tidy`, cross-platform build, example validation and the race-enabled unit suite |
| `Docs` | `docs` | `go generate ./...` leaves `docs/` unchanged |
| `Acceptance` | `acceptance` | the compose stack against the pinned OpenMetadata version |

`main` is protected by the `protect-main` repository ruleset: those four
contexts, a linear history, no force pushes or deletions, a merge queue and no
bypass actors. A job `name:` is the required context, so renaming one is a
breaking CI change: update `.github/workflows/test.yml`, the context list in
`scripts/setup-branch-protection.sh` and `scripts/check-branch-protection.sh`,
and the ruleset together.

Run `./scripts/check-branch-protection.sh` to verify the live ruleset and
`./scripts/setup-branch-protection.sh` to install or repair it (needs admin on
the repository). The weekly **Branch Protection Drift** workflow runs the same
check and opens an issue on drift; it needs a `BRANCH_PROTECTION_TOKEN` secret
with administration read permission.

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

## Reporting bugs and requesting features

Use the issue templates under `.github/ISSUE_TEMPLATE/`. For a bug, include the
provider version, Terraform version, OpenMetadata version, the authentication
mode (never the credential itself), the configuration that triggered it, and the
relevant diagnostic with tokens, passwords and hostnames removed.

## Security

Do not open a public issue for a suspected vulnerability. Follow
[`SECURITY.md`](SECURITY.md): report through GitHub's private security
advisories for this repository.
