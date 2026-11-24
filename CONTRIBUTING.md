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
make coverage
make coverage-gate
make lint
make vet
make validate-examples
make generate
```

## Acceptance tests

Acceptance tests run against a real OpenMetadata server in Docker Compose. The
stack is PostgreSQL, Elasticsearch, a one-shot schema migration and the
OpenMetadata server; the supported window is listed in
[`.github/openmetadata-versions.txt`](.github/openmetadata-versions.txt) and
passed to the images and to `OPENMETADATA_EXPECT_VERSION`.

```bash
make testacc-docker                          # newest version in the file
make testacc-docker OPENMETADATA_VERSION=2.0.8
make testacc-matrix                          # every version in the file
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
export OPENMETADATA_VERSION=$(tail -n 1 .github/openmetadata-versions.txt | tr -d '[:space:]')
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

**Changing the supported window:** edit
`.github/openmetadata-versions.txt` (the floor first, then the stable releases
in ascending order), update the `OpenMetadata 1.13` floor in `README.md` and
`SECURITY.md` if it moved, and run `make testacc-matrix`. The Makefile and CI
pass the selected entry to `docker-compose.yml` as `OPENMETADATA_VERSION`; only
update the compose fallbacks when they no longer mirror the newest entry, and
never hand-edit the workflow's expected version.

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
| `ci / Lint & Security` | `ci` → `lint` | golangci-lint and gosec |
| `ci / Test` | `ci` → `test` | build, vet, `go mod tidy`, cross-platform build, example validation and the race-enabled unit suite |
| `ci / Generate Docs` | `ci` → `docs` | `go generate ./...` leaves `docs/` unchanged |
| `acceptance / Acceptance` | `acceptance` → roll-up | the compose stack against the supported OpenMetadata window, plus combined unit and acceptance coverage |

`main` is protected by the `protect-main` repository ruleset: those four
contexts, a linear history, no force pushes or deletions, a merge queue and no
bypass actors. The job shapes live in
`caberdo/terraform-provider-kit/.github/workflows/provider-ci.yml` and
`provider-acceptance.yml`; the caller jobs in `.github/workflows/test.yml` are
named `ci` and `acceptance`, which is the prefix of every context above. Renaming
a caller job or a called job's display name is a breaking CI change: update the
workflows, `required-contexts` in `.github/workflows/protection-drift.yml`, this
table and the ruleset together.

This repository's ruleset is named `protect-main`. tfkit defaults to
`main-protection`, which is not this repository's ruleset, so pass
`--ruleset protect-main` (or set `RULESET_NAME=protect-main`) on every command:
without it `check` reports that `main-protection` does not exist, and `setup`
creates a second ruleset named `main-protection` beside the real one.

```sh
tfkit branch-protection check --ruleset protect-main \
  --contexts "ci / Lint & Security,ci / Test,ci / Generate Docs,acceptance / Acceptance"
tfkit branch-protection setup --ruleset protect-main \
  --contexts "ci / Lint & Security,ci / Test,ci / Generate Docs,acceptance / Acceptance"
```

`setup` needs admin on the repository. The same commands run through `go run
github.com/caberdo/terraform-provider-kit/cmd/tfkit@v0.8.0`. The weekly
**Branch Protection Drift** workflow runs the check and opens an issue on
drift; it needs a `BRANCH_PROTECTION_TOKEN` secret with administration read
permission.

The Markdown link check runs in `ci / Validate Markdown Files`, which is not
one of the four required contexts above. It is therefore advisory: a broken
link shows up on the pull request but no longer blocks a merge unless that
context is added to the ruleset. The four contexts listed here are the merge
gate; if `ci / Validate Markdown Files` should be required, add it to
`required-contexts` and the ruleset together.

## Coverage floors

`make coverage-gate` (that is, `tfkit coverage gate`) enforces two things on the
unit profile (`coverage.out`):

- the module total against `.github/coverage-baseline.txt`, which the README
  badge tracks;
- every package returned by `go list ./internal/...` against its entry in
  `.github/coverage-floors.txt`, or that file's `default` floor.

The package list is discovered, so a new `internal/resources/*` or
`internal/datasources/*` package is gated the moment it exists; a package with
no tests measures 0% and fails. The gate prints every package it checked and
names each one below its floor with the delta.

To **raise** a floor: improve coverage, run `make coverage`, then edit the
package's line. To **lower** one: do it in the same pull request as the change
that dropped coverage and explain why in the pull request. Lowering a floor
without that justification is a review finding. `make coverage-gate` checks the
current `coverage.out`, and `make coverage-readme-check` fails when the README
badge has drifted from the baseline.

## Release

A release is a signed `v*` tag cut after the release pull request merges: the
newest `CHANGELOG.md` section names the version, and there is no `VERSION` file.
The release pull request assembles the `CHANGELOG.md` `Unreleased` section from
the `changelog.d/` fragments and updates the README install constraint when
the minor line moves. Merging it publishes nothing.

Cut the tag from a clone of `main` and push it; that push runs **Release**, which
publishes the tag with GoReleaser, signed checksums, the registry manifest, build
provenance and an SBOM:

```bash
git tag -s vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

Dispatch **Release** with `tag: vX.Y.Z` to publish or republish a tag that
already exists; a dispatch cannot choose a new version, because nothing but a
tag names one.

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
