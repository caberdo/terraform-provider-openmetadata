# Architecture

The provider is a thin translation layer between Terraform state and the
OpenMetadata REST API. It owns no truth of its own: every read is an API call,
every mutation is a create, patch or delete, and an attribute the server cannot
return is preserved rather than zeroed.

Every entity copies the glossary slice
(`internal/resources/glossary/`, `internal/datasources/glossary/`). There is no
second pattern.

## Request flow

```
Terraform CLI
     │ plugin protocol 6.0
     ▼
main.go ── providerserver.Serve ──► internal/provider
                                        │ schema, env fallback, Configure
                                        │ builds client.Config → client.Client
                                        ▼
                             internal/providerdata.ProviderData
                                   │                  │
                    ┌──────────────┘                  └─────────────┐
                    ▼                                               ▼
          internal/resources/<entity>                    internal/datasources/<entity>
          58 resources                                   58 data sources
                    │                                               │
                    └──────────────► internal/client ◄──────────────┘
                                     lazy login, SDK memoization,
                                     raw DELETE, error classification
                                            │
                              ┌─────────────┴─────────────┐
                              ▼                           ▼
                  openmetadata-go-client            net/http
                  (CRUD, retries, typed models)     (login, version, delete)
                              │                           │
                              └─────────────┬─────────────┘
                                            ▼
                                  OpenMetadata server
                                  /api/v1/<entity>
```

## Client

`internal/client.Client` is the only path to the server. `client.New` validates
configuration and returns a client that performs no network I/O; `Configure`
therefore succeeds without a reachable server, so `terraform validate` and
`terraform plan` work offline.

- **Configuration.** `New` trims the trailing slash from `host`, requires an
  `http://` or `https://` scheme, defaults `api_version` to `v1`, rejects a
  negative retry count, and defaults an empty `retry_wait` to 30s. The base URL
  is `<host>/api/<version>`.
- **Authentication.** `SDK(ctx)` takes a mutex and memoizes the SDK client.
  A configured `token` is used as a bearer token. Without a token the first
  call exchanges `username`/`password` at `POST /api/<version>/users/login`;
  the password is Base64-encoded because the endpoint rejects plaintext. The JWT
  is cached for the process lifetime. The login response is accepted as a JSON
  string, a plain-text body or an object carrying `accessToken`, `token`,
  `jwt` or `id_token`.
- **SDK reuse.** The wrapper does not reimplement REST semantics. The SDK
  supplies typed request and response models, the computed retry policy
  (`retry_count` after the initial attempt, retried after `retry_wait`) and
  the bearer header. The HTTP client is shared and bounded by a 2-minute
  timeout per call.
- **SDK gaps.** Six services expose no `Delete` in the Go SDK: domains, data
  products, documents, personas, queries and event subscriptions. Those
  resources call the narrow `Client.Delete`, which sends
  `DELETE <host>/api/<version>/<path>` with the bearer token and optional query
  parameters. A destroy is never a no-op.
- **Version probe.** `Client.Version` reads `GET /system/version`; the
  acceptance precheck and `OPENMETADATA_EXPECT_VERSION` use it to fail before a
  stale server can masquerade as a passing run.
- **Bounded reads.** Every raw response body is capped at 1 MiB, so an HTML
  error page or a proxy response cannot be buffered without limit.

## Provider configuration

`internal/provider/provider.go` resolves attribute → environment → default,
then validates the cross-field rules before any request is made:

| Attribute | Environment | Default | Rule |
|---|---|---|---|
| `host` | `OPENMETADATA_HOST` | none | Required; must include the scheme |
| `token` | `OPENMETADATA_TOKEN` | none | Takes precedence over username/password |
| `username` | `OPENMETADATA_USERNAME` | none | Required with `password` when `token` is empty |
| `password` | `OPENMETADATA_PASSWORD` | none | Required with `username` when `token` is empty |
| `api_version` | `OPENMETADATA_API_VERSION` | `v1` | Passed to the SDK |
| `retry_count` | `OPENMETADATA_RETRY_COUNT` | `3` | Zero or greater |
| `retry_wait` | `OPENMETADATA_RETRY_WAIT` | `30s` | Go duration, greater than zero |

`token` and `password` are marked `Sensitive`; they are used in memory and are
never written to state or a diagnostic. `Configure` hands one
`*providerdata.ProviderData` to every resource and data source through
`ResourceData` and `DataSourceData`.

## Resource pattern

Each entity owns one package with `resource.go`, `model.go` and tests. The
resource implements `resource.Resource`, `ResourceWithConfigure` and
`ResourceWithImportState`:

- **Metadata** returns `<provider type name>_<entity>`, where the provider type
  name is `openmetadata`.
- **Schema** declares computed server fields (`id`, `fully_qualified_name`,
  `version`, timestamps), user-managed fields, validators and plan modifiers.
  An attribute with `Default` is also `Computed`; Terraform rejects the
  combination otherwise.
- **Configure** type-asserts `req.ProviderData` to
  `*providerdata.ProviderData` and stores the client.
- **Create** builds the typed `Create<Entity>` body, calls the SDK and applies
  the returned object to state.
- **Read** resolves the identifier, removes the resource from state on HTTP 404,
  and applies the server representation. A missing API field is mapped to the
  schema default or to null, never to a zero value the user did not configure.
- **Update** sends a JSON Patch document only for fields that changed, through
  `ometa.JSONPatchOp`. Fields managed outside Terraform (tags, owners,
  reviewers, and entity-specific relationships) are never resent, so they
  survive an apply. Some fields are fixed at creation and silently keep the old
  value on the server; those attributes carry `RequiresReplace()` and emit no
  patch operation.
- **Delete** calls the SDK delete with the documented hard-delete and recursive
  flags, or the raw helper when the SDK has no method. A 404 is treated as
  already gone.
- **ImportState** passes the ID through. Import and Read accept either the UUID
  or the fully qualified name; the lookup tries the UUID first and falls back to
  the name on 404 or 400.

`model.go` holds the flat `tfsdk` model and an `apply` function that copies the
API object into it. Nested entities with their own Terraform shape keep local
model types; only genuinely shared conversions live in `internal/convert`.

Child entities derive their fully qualified name from their parents (service,
database, schema, table, and so on). The derivation is done from the configured
parent names, and reads use the same path so a server-side FQN is never guessed.

## Data source pattern

A data source is read-only: `Metadata`, `Schema`, `Configure` and `Read`. It
accepts either the UUID or the fully qualified name, resolves it with the same
lookup helper as the resource, and copies the server object into its own model.
It has no create, update, delete or import path.

## Error taxonomy

`internal/client` normalizes every non-2xx response into the SDK's
`*ometa.APIError`, so callers have one shape to reason about:

- `NotFound` — HTTP 404. A read removes the resource from state; a delete treats
  it as success.
- `BadRequest` — HTTP 400. Identifier lookups use it to fall back from UUID to
  fully qualified name.
- `Conflict` — HTTP 409. Available for callers that need to distinguish a
  duplicate create.
- `Unauthorized` — HTTP 401 or 403. Available for credential and permission
  classification.
- `ErrorDetail` / `ErrorDiagnostic` — render the status and server message for
  a Terraform diagnostic. Secrets are never included.

## Test layers

| Layer | What it proves |
|---|---|
| L1 — pure unit | provider configuration, login exchange, token shapes, error classification, version probe, raw delete, conversion helpers |
| L2 — entity logic | schema shape and required/computed flags, model mapping, patch-document construction |
| L3 — acceptance | real OpenMetadata in the compose stack, `TF_ACC=1`, full CRUD, update, destroy, and both import identifiers |

Unit tests live in the same package as the code they cover. Acceptance tests
live in `package <entity>_test` and import `internal/acctest`; importing
`internal/acctest` from the internal package would create an import cycle
through `internal/provider`.

`internal/acctest` holds the stable harness:

- `ProtoV6ProviderFactories` wires the framework provider into
  `terraform-plugin-testing`.
- `PreCheck` skips without `TF_ACC`, requires a host and credentials, probes
  `/system/version` with a 30-second deadline, and fails when
  `OPENMETADATA_EXPECT_VERSION` disagrees with the server.
- `NewClient` builds a client from the environment so `CheckDestroy` can prove
  an entity is gone instead of trusting Terraform state.

The compose stack (`docker-compose.yml`) runs PostgreSQL 15, a single-node
Elasticsearch, a one-shot `openmetadata-ops.sh migrate`, and the OpenMetadata
server. The server version is pinned once in
`.github/openmetadata-versions.txt`; the Makefile and CI pass the same value as
`OPENMETADATA_VERSION`, and the precheck compares it with `/system/version`.
There is no version matrix, so one pinned release is the supported acceptance
target at a time.

Acceptance tests must not skip once `TF_ACC` is set. A missing capability or an
unreachable server is an environment failure, not a pass. They must compile
under `go vet ./...` even when they are not run locally.

## Dependency rules

1. `internal/client` imports only the standard library and
   `openmetadata-go-client`; it knows nothing about Terraform.
2. `internal/convert` imports the framework's `types` and `diag` only.
3. `internal/providerdata` imports `internal/client` and is the carrier every
   entity depends on.
4. `internal/resources/*` and `internal/datasources/*` import `client`,
   `convert`, `providerdata` and their local models. Nothing under those trees
   imports `internal/provider`.
5. `internal/acctest` imports `client` and `provider`; it is only imported by
   external `_test` packages.
6. `internal/provider` imports every entity package and is the only place
   registration lives. Registration lists stay alphabetical so parallel
   contributors insert in predictable places.

## Documentation and releases

`examples/` and `templates/` are rendered into `docs/` by `tfplugindocs` through
`go generate ./...`; `docs/` is never hand-edited. CI regenerates the docs and
fails on drift, including new untracked files.

The `Create Release Tag` workflow signs a tag and calls `Release`, which runs
GoReleaser for linux, darwin and windows on amd64 and arm64, signs the checksum
file, attests build provenance, emits an SBOM and publishes the registry
manifest. `make release-snapshot` builds the same platform set locally.
