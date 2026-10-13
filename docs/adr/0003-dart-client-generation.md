# ADR-0003 — Dart client generation: swagger_parser (GO, with two conditions)

| Field | Value |
|---|---|
| Status | ACCEPTED (P0-21 spike) |
| Date | 2026-10-13 |
| Related | [ADR-0001 O4 and mobile Context7 verification](0001-stack.md), [PLAN P0-21 / P0-22](../project-specs/PLAN-control-contable.md), [contract](../api/openapi.yaml) |
| Owner | frontend-sr |

## Context

ADR-0001 chose `swagger_parser` to generate the Flutter API client from
`docs/api/openapi.yaml` (OpenAPI **3.1.0**), with one open risk: does it handle our 3.1
constructs? The fallback was `openapi-generator` `dart-dio` (JVM) fed through an overlay
that downgrades the spec to 3.0, at about +4 h.

Constructs in the contract at the time of the spike (2439 lines, 35 paths):

| Construct | Where | Count |
|---|---|---|
| `oneOf: [$ref, {type: 'null'}]` | `TaxpayerRegime.end_period`, `RegimeEstimate.isr` / `iva_components`, `BreakdownLine.rate`, `TaxInputValue.value_*`, `HistoryMonth.iva` / `isr` | 9 |
| `oneOf: [inline object, {type: 'null'}]` | `SubscriptionStatus.plan` | 1 |
| `additionalProperties: {oneOf: [string, boolean, array]}` | `RegimeOptions` | 1 |
| `const` | `Legend`, `TokenPair.token_type`, `RegisterAccepted.status` | 3 |
| `type: [T, 'null']` (incl. `[string, boolean, 'null']`) | 20+ fields | — |
| Path-level `parameters` | `/v1/taxpayers/{taxpayerId}/...` | 17 path items |

## Spike (time-box 4.5 h, used about 1 h)

Throwaway Dart package in the session scratchpad, not committed. Exact pins:
swagger_parser 1.45.0, freezed 4.0.2, freezed_annotation 3.1.0, json_serializable 6.14.1,
json_annotation 4.12.0, retrofit 4.10.0, retrofit_generator 10.2.11 and 10.2.12,
build_runner 2.16.2, dio 5.11.1, Dart 3.13.5.

Config: `json_serializer: freezed`, `use_freezed3: true`, `root_client: true`,
`unknown_enum_value: true`, `put_clients_in_folder: true`.

Steps: `dart run swagger_parser` → `dart run build_runner build -d` → `dart analyze lib`
→ a round-trip decode test (4 cases) → a script that checks every component schema
generated with the right kind (enum / class / typedef).

## Evidence

### What broke

1. **Path-level `parameters` crash the parser.** Run 1 failed before writing anything:
   `type 'List<Object?>' is not a subtype of type 'Map<String, dynamic>' in type cast` at
   `open_api_parser.dart:634`. `parseRestClients` casts every path-item value to a Map
   *before* it skips the `parameters` key. Path-level parameters are valid OAS (3.0 and
   3.1), so this is an upstream bug, not a 3.1 gap.
   *Workaround proven:* a 25-line Dart pre-step copies path-level parameters into each
   operation (spec-equivalent) and writes a JSON copy for the generator. After it:
   **10 clients, 49 requests, 107 data classes, 119 files**; build_runner wrote 230 outputs;
   `dart analyze lib`: **No issues found**.
2. **Silent type collision (the dangerous one).** The inline enum
   `SubscriptionStatus.plan.period` (`weekly|monthly|yearly`) is generated as `Period` and
   **overwrites the component `Period`** (the `YYYY-MM` string). Everything still compiles,
   but `TaxpayerRegime.startPeriod`, `HistoryMonth.period`, etc. become that enum, and
   `'2026-01'` would decode as `$unknown` at runtime. The kind-check script found exactly this
   one mismatch across all components.
   *Remedy proven:* extract the inline object to a named component `SubscriptionPlan`
   (`plan: oneOf: [$ref SubscriptionPlan, null]`). The enum becomes
   `SubscriptionPlanPeriod`, `Period` is back to `typedef Period = String`, the kind check
   is clean, analysis clean. The wire format does not change.

### What generated correctly

| Construct | Dart result | Verified by |
|---|---|---|
| `oneOf [$ref, null]` | `Period? endPeriod`, `TaxResult? isr`, `DecimalRate? rate` | decode test: `null` and value both round-trip |
| `oneOf [inline object, null]` | `SubscriptionPlan? plan` (after remedy 2) | decode test: both branches |
| `additionalProperties: oneOf[...]` | `typedef RegimeOptions = Map<String, dynamic>` | decode test: string, bool and list values |
| `const` | plain `String` / `typedef Legend = String` | decode test. The constant is not enforced client-side; the server is the authority |
| `type: [T, 'null']` | `T?` (e.g. `DateTime? lastViewedAt`) | analyzer + generated code |
| `type: [string, boolean, 'null']` | `dynamic defaultValue` | generated code. Acceptable: `RegimeOptionDef.default` is read only to prefill a form |
| enums | `@JsonEnum` with `$unknown` fallback | decode test: an unknown `ErrorCode` does not throw |
| `204` responses | `Future<void>` | generated clients |
| path / query params, defaults | `@Path('taxpayerId') required Uuid`, `@Query('include_archived') bool? = false` | generated clients |

## Decision

**GO with swagger_parser 1.45.0.** The fallback (openapi-generator `dart-dio` + 3.0 overlay)
is not needed: both failures have small, proven remedies, and the 3.1 constructs we worried
about (`oneOf`+null, `const`, type lists) all work.

Conditions. Each is tracked in the PLAN: (1), (3) and the upstream issue in P0-23, CI wiring in
P0-25, (2) is done in this PR:

1. **Pre-step normalizer.** `mobile/tool/normalize_openapi.dart` (the spike's path-parameter
   inliner, with a unit test first) runs before `swagger_parser` and is the only way the
   client is generated (`make mobile-client` or equivalent). The contract keeps its
   path-level parameters. Upstream bug: already filed as
   [Carapacik/swagger_parser#374](https://github.com/Carapacik/swagger_parser/issues/374)
   (open; a commenter traced it to path-level `parameters`; still present in 1.45.0), so no
   duplicate was opened. Delete the pre-step once a release that fixes #374 is pinned.
2. **Contract rule: no inline object schemas inside `oneOf`/`anyOf`.** Name them as
   components (see `docs/api/CONTRACT.md`). Applied in this PR: `SubscriptionStatus.plan` is
   now `oneOf: [$ref SubscriptionPlan, null]`. Non-breaking: the wire format is identical, the
   node is not frozen, and no backend code is generated from it yet. `make contract-lint` passes.
3. **CI guard.** `mobile/tool/check_generated_kinds.dart` (the spike's kind-check script,
   test first against a fixture that reproduces the `Period` collision) runs inside
   `make mobile-client` and fails it when a generated type's kind (enum / class / typedef)
   differs from its component schema, or two schemas map to one Dart name. Mobile CI (P0-25)
   runs `make mobile-client`, so a future inline-enum collision cannot pass silently.
4. **Pins.** Exact versions as above, with **retrofit_generator 10.2.11** (10.2.12 was
   published on 2026-10-13 and is under the two-week rule; the spike passed on both).

## Implementation (P0-23)

| Piece | Where |
|---|---|
| Normalizer (condition 1), unit-tested | `mobile/tool/normalize_openapi.dart`, `mobile/tool/src/openapi_normalizer.dart`, `mobile/test/tool/normalize_openapi_test.dart` |
| Kind guard (condition 3) | `mobile/tool/check_generated_kinds.dart`, `mobile/tool/src/generated_kinds.dart`. Its test runs the real swagger_parser (in memory, with `mobile/swagger_parser.yaml`) over `test/tool/fixtures/inline_enum_collision.yaml` and must report the `Period` typedef/enum collision; the named-component fixture must pass |
| Generator config | `mobile/swagger_parser.yaml` (input: the normalized JSON in `mobile/build/openapi/`), `mobile/build.yaml` |
| Pipeline | `make mobile-client`: normalize → swagger_parser → build_runner → kind guard → `dart format`. Output committed in `mobile/lib/api/generated/` |
| Drift check | `make mobile-client-check`: regenerates and fails if `mobile/lib/api/generated/` differs from what is committed (P0-25 runs it in CI) |
| Wiring | `apiClientProvider` (`mobile/lib/core/network/api_client_provider.dart`) builds the generated `ApiClient` on the single `dioProvider` (per-flavor base URL) |

Expected kinds: a schema with `enum` → `enum`; with `properties` or `allOf` → class; anything
else (scalars, arrays, maps, `const`) → `typedef`. The guard also fails when two schemas map
to one Dart name, when one Dart name is declared twice, or when a component was not generated.

## Consequences

- (+) Pure-Dart toolchain, no JVM in mobile CI; idiomatic Freezed 3 + retrofit output;
  the client is regenerated from the contract, never hand-written.
- (+) The collision guard also protects future contract edits.
- (−) One pre-step script to own until upstream fixes the path-parameter bug.
- (−) `const` and multi-type unions lose precision in Dart (`String`, `dynamic`). Server
  validation stays the authority; the UI must not branch on these values without a typed
  wrapper.
- Fallback stays documented: if a future contract construct breaks swagger_parser without a
  small remedy, switch to openapi-generator `dart-dio` + 3.0 overlay (ADR-0001 O4) under a new ADR.
