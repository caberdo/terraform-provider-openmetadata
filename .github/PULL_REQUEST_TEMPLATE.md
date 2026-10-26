## Description

<!-- What changes and why. -->

## Linked Issues

<!-- Fixes #123 -->

## Scope / owned paths

<!--
State the task's exclusive write paths and confirm the diff stays inside them.
For an entity task, copy the paths from TASKS.md.
-->

- Owned paths:
- [ ] No writes outside the owned paths (`git diff --name-only <base>...HEAD`)

## Testing

- [ ] `make build` succeeds
- [ ] `make vet` is clean
- [ ] `make test` passes (unit tests)
- [ ] `make lint` is clean
- [ ] `make generate` run and `docs/` committed, if a schema changed
- [ ] `make validate-examples` passes, if an example changed
- [ ] New or changed resources/data sources have unit tests for schema, create, update, delete and import (valid + invalid)
- [ ] A regression test covers this change, if it is a bug fix
- [ ] Acceptance tests pass in CI (`Acceptance`), if CRUD behaviour through Terraform Core changed

## Documentation

- [ ] `make generate` run (if a schema changed; never edit `docs/` by hand)
- [ ] Examples added or updated under `examples/` (if user-facing behaviour changed)
- [ ] README/CHANGELOG updated, if the change affects usage or releases

## Entity checklist

- [ ] Exactly one alphabetical insertion in `internal/provider/provider.go` for each of `Resources()` and `DataSources()`
- [ ] Acceptance tests are in `package <entity>_test` and import `internal/acctest`
- [ ] Every attribute with `Default` also sets `Computed`
- [ ] Updates send only changed fields through `ometa.JSONPatchOp`; no full replace
- [ ] Read preserves state for attributes the server cannot return

See [AGENTS.md](../AGENTS.md) for the parallel work contract and
[CONTRIBUTING.md](../CONTRIBUTING.md) for the development setup.
