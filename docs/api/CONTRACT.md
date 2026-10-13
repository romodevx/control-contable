# API contract rules

`docs/api/openapi.yaml` is the single source of truth. It is linted on every PR that touches
`docs/api/**`, `redocly.yaml` or `tools/contract/**` (`.github/workflows/contract.yml`), and locally with:

```sh
make contract-lint        # lint the contract
make contract-lint-test   # prove each rule rejects what it should (fixtures)
```

Tooling is pinned: Redocly CLI `2.62.1` in `tools/contract/package.json` + `package-lock.json`,
installed with `npm ci`. Ruleset: [`redocly.yaml`](../../redocly.yaml); custom rules:
[`tools/contract/plugin.js`](../../tools/contract/plugin.js).

## Enforced rules

| Rule | What it enforces |
|---|---|
| `recommended-strict` | Redocly's recommended set, every rule as an error |
| `operation-operationId` (+ `-unique`, `-url-safe`) | Every operation has a unique, URL-safe `operationId` (codegen and the operationID→flag map depend on it) |
| `rule/operation-x-module` | Every operation declares `x-module`: `platform` or a `module_<name>` flag key |
| `contract/error-response-has-body` | Every `4xx`/`5xx`/`default` response (inline or `$ref`) has at least one media type |
| `rule/error-response-envelope` | Every media type of that body has a schema, and it is `$ref: '#/components/schemas/ErrorResponse'` (single error envelope) |
| `contract/body-example` | Every request/response body has an example, on the media type or on the schema it references |
| `no-invalid-media-type-examples`, `no-invalid-schema-examples` | Every example validates against its schema (no extra properties) |
| `security-defined` | Every operation is covered by a declared security scheme (global or its own) |
| `contract/public-operation-explicit` | `security: []` (public) only together with `x-public: true`, and vice versa |
| `contract/frozen-unchanged` | Any node marked `x-frozen: sha256:<hash>`, anywhere in the document, still hashes to that value |

## Reviewed conventions (not lint-enforced yet)

| Convention | Why | Backstop |
|---|---|---|
| No inline object schemas inside `oneOf`/`anyOf`: name them as components and `$ref` them | swagger_parser names an inline object's enums after the bare property, so `SubscriptionStatus.plan.period` silently overwrote the `Period` component in the Dart client ([ADR-0003](../adr/0003-dart-client-generation.md)) | Mobile kind guard in `make mobile-client` (PLAN P0-23, run in CI by P0-25) |

## Frozen contract

A frozen node is a published promise: the backend, the generated Dart client and the
contract tests already depend on it. It is marked in the contract itself:

```yaml
x-frozen: sha256:<sha256 of the node's canonical JSON, keys sorted, x-frozen excluded>
```

The hash covers the node as written (a `$ref` is hashed as the reference string, not the
target), so every schema a frozen operation exposes is frozen on its own.

### Frozen today

| Node | Frozen by | Why |
|---|---|---|
| `GET /healthz` (`getHealth`) | P0-05 | Liveness/readiness probe used by deploy health checks |
| `GET /v1/flags` (`listEnabledFlags`) | P0-05 | Feeds the mobile FlagsProvider |
| `components.schemas.ErrorResponse` | P0-05 | The single error envelope every endpoint returns |
| `components.schemas.FieldError` | P0-05 | Item of `ErrorResponse.error.details` |
| `components.schemas.HealthStatus` | P0-05 | `/healthz` body |
| `components.schemas.EnabledFlags`, `ModuleKey` | P0-05 | `/v1/flags` body |
| M1 `module_auth` operations: `register`, `verifyEmail`, `resendVerification`, `login`, `refreshSession`, `logout`, `getMe`, `updateMe`, `deleteMe`, `listLegalDocuments` | M1-01 | [MODULE-auth](../project-specs/MODULE-auth.md) (D1-D6); the backend (M1-05..M1-09) and the Flutter auth screens (M1-11..M1-13) build on it |
| `components.schemas.RegisterRequest`, `DocumentAcceptance`, `VerificationSent`, `VerifyEmailRequest`, `ResendVerificationRequest`, `LoginRequest`, `RefreshRequest`, `TokenPair`, `AuthSession`, `User`, `UserUpdate`, `DeleteAccountRequest`, `LegalDocumentKind`, `LegalDocument`, `LegalDocumentList` | M1-01 | Bodies of the M1 operations |
| `components.headers.NoStore` | M1-01 | `Cache-Control: no-store` on every response that carries tokens or account data |

**`FieldError.field` naming (one convention).** A dotted path relative to the request body
(`password`, `accepted_documents.1.kind`), `body` for the body root, and the bare parameter
name for path/query/header errors (`taxpayerId`, `period`). It is never a `/`-prefixed JSON
Pointer, in examples, specs or handlers. `FieldError.code` is the failed JSON Schema keyword
(`required`, `minLength`, `enum`, `additionalProperties`, …) for request validation, or a
snake_case domain code from the handler (`password_policy`). The router produces both from
either kin-openapi engine (`backend/adapter/http/router/fields.go`).

`ErrorCode` is **append-only**, not frozen: modules add codes as they land. Removing or
renaming a code is a breaking change (rules below). Clients must treat unknown codes as
`internal_error`-like generic failures.

Module contracts are frozen the same way at each module's "contract freeze" task in the PLAN
(mark its operations and body schemas with `x-frozen`).

### Changing a frozen node

1. **Breaking** (removing/renaming a field, path, enum value; tightening validation; changing
   a type or status code): needs an ADR and a version bump (`info.version`; a new path
   version when shipped clients would break). Prefer adding a new field/operation instead.
2. **Non-breaking** (new optional response field, wording of a description, new example):
   allowed; the PR says why it is non-breaking.
3. In both cases, the lint error prints the new hash. Set `x-frozen` to it in the same PR;
   the tech-lead gate reviews every `x-frozen` diff.

To freeze a new node, add `x-frozen: sha256:pending`, run `make contract-lint` and copy the
`actual` hash from the error.

## Contract tests (P0-18)

Real responses from the real router are validated against this file with
[libopenapi-validator](https://github.com/pb33f/libopenapi-validator) `v0.16.1` (pinned per
ADR-0001), independently of the kin-openapi copy the server is generated from. The harness is
`backend/adapter/http/contracttest`:

```go
c := contracttest.Load(t)                                         // parses this file once per test binary
rec := c.ServeOp(t, contracttest.Op("createIncome"), handler, req) // serve + validate; counts as coverage
rec = c.Serve(t, handler, req)                                     // serve + validate; no coverage claim (e.g. flag-OFF 404)
```

- Status, declared response headers and body are checked in **strict mode**: an undeclared
  property fails, and so does any response header the matched response does not declare,
  except the transport headers listed (with the reason for each) in `transportHeaders`
  (`backend/adapter/http/contracttest/headers.go`): `Content-Type`, `Content-Length`, `Date`,
  `X-Request-Id`, `X-Content-Type-Options`. The header check is the harness's own:
  libopenapi-validator's strict mode skips responses that declare no headers and ignores
  `Set-Cookie`, `Cache-Control`, `ETag`, `Vary`, ... by default. A response that needs one of
  those declares it in this file. Formats (`uuid`, `date`, ...) are asserted.
- A failure prints the operation, the declared statuses, each violation with its JSON path and
  the response received (status, headers, pretty-printed body).
- `ServeOp` also fails when the request does not route to the named operation.

**Coverage of the contract.** `TestContractCoverage` scans the backend's Go files (every build
tag) for `contracttest.Op("<operationId>")` literals. Every operation must be either covered or
listed in `pendingOperations` (`backend/adapter/http/contracttest/coverage_test.go`) with the
PLAN task that owns its tests; never both. The list only shrinks: a module's contract-test task
covers its operations and removes them from it in the same PR. A new operation is added to the
list (or covered) in the PR that adds it to the contract.

Run them with `make test-integration` (they need Docker), or only the contract suites with
`make test-contract`. CI runs them in the backend workflow's integration job.
