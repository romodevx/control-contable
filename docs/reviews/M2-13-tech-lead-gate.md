# M2-13: Tech-lead gate of module M2 Taxpayers & Onboarding

- **Date:** 2026-10-13
- **Branch:** `feature/taxpayers-qa-fixes` (PR #68 → `dev`), which carries M2-11 and M2-12 on top of all of M2 (M2-01 … M2-10).
- **Scope:** A) PR #68 (M2-11 flag tests, M2-12 QA report). B) Holistic review of M2:
  `backend/internal/taxpayers/**`, the regimes catalog and taxpayers migrations, the handlers and
  composition root, `mobile/lib/features/taxpayers/**`. Contract, spec and code checked against
  each other ([MODULE-taxpayers](../project-specs/MODULE-taxpayers.md), all OQ DECIDED, D-RFC).
  C) The three QA-deferred UX items D1, D2, D3 of [M2-12](M2-12-qa.md).
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language
  (idiomatic Go wins ties), SOLID, `~/.claude/rules/`.
- **Not covered here:** the security gate (M2-14) runs separately.

## Verdict: **PASS**

One blocker and two High defects were found in this gate. All three are fixed on this branch, and
each fix has a failing test committed before it. The three QA items are also fixed test-first.
No blocker is open.

Before `module_taxpayers` is enabled in a shared environment, FU1 (spec ↔ seed drift in the
compatibility matrix) must be settled. It is a product decision, not a code defect.

PR #68: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `make -C backend check` (gofmt, golangci-lint, `go test -race`) | green, 0 lint issues |
| `go test -tags=integration ./internal/taxpayers/... ./cmd/api/` (testcontainers) | green |
| `make mobile-check` (format, analyze, tests + coverage gate) | analyze clean, 744 passed, 17 skipped |
| B1 fix | `TestNewRFCRejects` (2 new cases) red at `eea7110`, green at `ea4ebf1` |
| H1/H2 fixes | `active_taxpayer_test` "stale list answers" (2) and `settings_screen_test` "a failed reload…" red at `d1a08f5`, green at `2519539` |
| D1/D2/D3 | failing tests at `8ac22a7`, `7ffed91`, `9a5fc39`; green at `14905ed`, `cc9c27f`, `f8dc421` |
| Gate scan (test relaxations in the branch diff) | 3 eligible hunks, all from this gate and none relaxing: a case moved from the accept table to the reject table (B1), a binding kept in a variable, and a weak assertion replaced by a stronger one |

## A) PR #68

| Item | Assessment |
|---|---|
| `taxpayers_module_test.go`: 11 operations written by hand from the spec, by operationId and by ServeMux pattern, and the flag owns nothing else | Good. The hand-written list makes the generated map agree with the spec, not just with itself. |
| `taxpayers_preset_smoke_integration_test.go`: each seeded preset plus a fixture preset without the flag, all 11 operations with a real user and taxpayer, and a guard that every owned operation is probed | Good. The guard fails if someone adds an M2 operation without a probe. |
| `taxpayers_flag_test.dart`: the route tree is exactly 4 locations; OFF means not found with zero repository calls; ON means each route opens | Good. It asserts zero calls, not only the 404 screen. |
| QA report M2-12 | Thorough: what was wiped, what was created through the UI, 22 flows, log PII grep, and environment problems kept apart from app bugs. The curl fast path for taxpayers 4–10 is justified in the report, and the 11th went through the UI. |
| TDD evidence | These are pinning tests over behavior that already existed, so they passed on the first run. The report says so. Acceptable for characterization tests. |
| Low (FU11) | The probe comment says destructive operations are "rejected before writing", but archive, unarchive and add-regime do write, so state changes across preset subtests. The assertions only check non-404, so this is not a correctness problem. `ownedOperations` is written twice (in `gen_test` and `cmd/api`). |

## B) Holistic findings

### Fixed on this branch (failing test first)

| # | Sev | Finding | Fix |
|---|---|---|---|
| B1 | **Blocker** | **An RFC whose last character is a letter other than `A` answered `500`.** `domain/rfc.go` allowed any `[A-Z0-9]` as the check digit and treated a mismatch as a warning (D-RFC). The DB CHECK `taxpayers_rfc_masked_format` requires `[0-9A]$`, so the insert failed with 23514. That constraint is not in `infra/errors.go` `constraintErrors`, so it surfaced as `internal_error`. The test table even accepted `A&B020304XYB` on purpose. Spec §2.2 and the mobile shape both say `[0-9A]`. *Clean Architecture: the domain is the authority and the DB is only the backstop; one grammar, many validators.* | `eea7110` (test), `ea4ebf1` (fix). `hasRFCShape` now requires `isCheckDigitRune` for the last character. Such an RFC gets `400 rfc_format`; a mismatch inside `[0-9A]` stays a warning. |
| H1 | **High** | **A resume refresh could show the previous account's taxpayers.** `TaxpayersNotifier._refreshOnResume` checked `ref.mounted` after its await. Riverpod 3.4.3 keeps the notifier across rebuilds (`classListenable.result ??=`), and `ref` is always the current build's, so the check never fires. A refresh started for user A and answered after sign-out and sign-in as B wrote A's list (names, masked RFCs) into B's state. The test reproduced it. *Clean Code ch.13: concurrency; the M1 generation-guard decision.* | `d1a08f5` (test), `2519539` (fix). Every build and refresh takes a generation number, and an older answer is dropped (latest request wins). |
| H2 | **High** | **A write that succeeded was reported as failed when the list reload failed.** `TaxpayerSettings.run` awaited `reload()`, which rethrew the list's `ApiFailure`. Archive, rename and add-regime then showed an error, and a retry answered `409` or `404`. *Clean Code ch.7: an error must mean what it says.* | Same commits. `reload()` keeps the last list on failure and logs it. This also stops `adopt()` leaking an `ApiFailure` from an `unawaited` call. |

### C) QA-deferred items (fixed, test first)

| QA | Decision (spec) | Fix |
|---|---|---|
| D1 cap after four steps | OD-2 / §2.8: the server's `422` is the authority. The client already holds the full list (`include_archived=true`). | `OnboardingFlow(capReached:)` starts at a new `limitReached` step: "Llegaste al máximo de contribuyentes … los archivados también cuentan". Nothing is loaded or sent. Read-only still comes first, because the server checks access before the cap (§2.10). The count is `taxpayerCap = 10` in `domain/taxpayer_cap.dart`, and an unloaded list says nothing. `ReadOnlyNotice` now delegates to a generic `StopNotice`. |
| D2 Undo for read-only | OQ-8 DECIDED: unarchive is `402` for a read-only user, so the spec is followed and unarchive is not allowed. | Once the session knows the user is read-only (a `402` was seen), Archivar opens a confirmation ("no podrás desarchivarlo hasta tener un plan"; house rule: confirm the irreversible), and the toast has no *Deshacer*. Before any `402` the reversible flow is unchanged. Knowing earlier needs the M7 access read (US-T4b). |
| D3 generic CPA warning | §2.5 and contract `Warning.field` + `Regime.review_pending_with` (0.4.0): no contract change. | The review step names the pair: "Registraste juntos «612 · …» y «621 · …». <reason>". The later regime comes from `regimes.N.regime_code`, and its partners are the earlier sent regimes linked in `review_pending_with` on either side (`reviewPendingPartners`). Add regime says "…su combinación con 606 y 612: …". If the field names no sent regime, only the reason is shown. |

### What holds up

- **Dependency rule.** `taxpayers/domain` imports only stdlib, `x/text/unicode/norm` and `platform/money`. GORM and pgconn appear only in `infra`. Interfaces are small and live with their consumers (`TaxpayerManager`, `AccessPolicy`, the stores).
- **Create ordering matches §2.8 and §2.10.** The order is: idempotent retry, then `402`, then the cap under the user-row lock (re-checked after the lock), then `400`, then `422`, then insert. The 5-at-9 race test holds a blocker and waits on `pg_stat_activity`, so a mutation would make it fail.
- **S-1/S-2.** The row id is generated before sealing. The AAD is `key_id ‖ taxpayer.id ‖ user_id`, and tests cover a ciphertext moved to another row or another user. The blind index includes `user_id`.
- **RLS.** ENABLE + FORCE on `taxpayers` and `taxpayer_regimes`, one policy per operation, with the user context set per transaction. Isolation tests probe both a never-set and a reverted session. Column grants keep `person_type`, `id` and `user_id` immutable, and the composite FKs reject forged references. The catalog tables are global config: ENABLE, read-only policy, no write grant.
- **Compatibility as data.** `linkPairs` is O(r + p log p). An unverified pair never blocks (DB CHECK), and the warning shape matches the contract.
- **Mobile layering.** Generated clients are used only in `data/taxpayers_repository.dart`, behind `guardApi`; M1-16 FU1 (`TypeError`) is fixed in core. `ApiFailure` is sealed, and `sortDetails` routes each server verdict to the step that owns the field.
- **Flags.** The server gates at mux registration, the client at router registration through `FeatureModule`. OFF suites assert 404 or not-found with zero calls.
- **Forms.** The primary button is never disabled (busy label, repeat taps ignored). Errors appear on blur or submit, delete is a modal that names the taxpayer, and reduced motion is respected.
- **Tests.** Deterministic: fake clock and client id, `Completer` gates, no sleeps. Table-driven in Go, with an independent oracle (python-stdnum) for the RFC fixtures.

### Follow-ups (not fixed here)

| # | Sev | Finding | Owner / target |
|---|---|---|---|
| FU1 | **High** | **Spec ↔ seed drift in the compatibility matrix.** §2.5's seed table refuses 621 × 612 and 626 × 621, and says "moral: every pair" is refused (US-T1b, OQ-6 "PM single-regime"). The M2-02 seed marks those pairs *allowed with a CPA-review warning* (`docs/reviews/CPA-REVIEW.md`). The server accepts a PM with 601 + 603; only the client's single-select stops it. Either amend §2.5 and OQ-6 with the reason ("an unverified rule never blocks"), or ship a forward migration that sets `allowed = false` for the pairs the spec states as rules. | pm + db-sr. Decide before `module_taxpayers` is ON in a shared environment. **Resolved** (`bugfix/regime-compatibility-reconcile`): 10 pairs refused from the law (forward migration `20261014120000`), 8 PM pairs kept allowed with a warning and §2.5/OQ-6 amended; research Part E. |
| FU2 | Medium | US-T2: "if the existing one is archived, the message suggests unarchiving it". Code and the frozen contract example always send the generic `rfc_already_registered` message (`usecase/present.go`). | pm (amend the spec), or backend-sr (look up the archived state on 23505). M2 hardening. |
| FU3 | Medium | M1-16 FU2 pattern copied: `unawaited(adopt/select/clear)` (`onboarding_screen.dart`, `taxpayer_switcher_sheet.dart`, `taxpayer_list.dart`). A store `PlatformException` still escapes as an unhandled async error. (H2 removed the `ApiFailure` path.) Also, a failed store read or write breaks `activeTaxpayerProvider` instead of falling back (§2.9). | frontend-sr, M2 hardening |
| FU4 | Medium | `OnboardingFlow.back()` and the step-5 *Omitir* are not blocked while the create is in flight. The user can leave, the create resolves on a disposed notifier, and the new taxpayer is never adopted. | frontend-sr, M2 hardening |
| FU5 | Medium | M1-16 FU4 still open and copied: the `SemanticsService.sendAnnouncement` boilerplate (`regimes_step.dart`, `settings_feedback.dart` paywall sheet, `StopNotice`). There are no `takeAnnouncements` assertions for the paywall or for settings errors. | frontend-sr, before M3 forms |
| FU6 | Medium | `inline_text_editor.dart`: focus is not returned to the *Editar* trigger after Save or Cancel (ui-forms rule: focus return). | frontend-sr, M2 hardening |
| FU7 | Medium | SRP/DRY: `onboarding_flow.dart` (~430 lines) mixes the step machine, request building, RFC correction and tax-input paging, and duplicates option drafting and the failure switch with `AddRegimeFlow`. Extract a `RegimeDraft` both flows compose, and a post-create flow. | frontend-sr, refactor before M3 |
| FU8 | Low | §2.9 says the active id is stored "per user". It uses one global key, cleared on sign-out from inside the list notifier, which only runs while that provider is alive. Key it by user id. | frontend-sr, M2 hardening |
| FU9 | Low | Backend: `toOptions` drops the `setOption` error (`adapter/http/present.go:135`). `errors.Join` is used to add context where `%w` belongs (`usecase/regimes.go:179`). The spec says `FOR UPDATE` but the code (correctly) uses `FOR NO KEY UPDATE`, so update §2.8 and S-5. `rfcvault` takes `KeyID` from `ActiveKeyID()` rather than `KeyIDOf(ciphertext)`. Create runs one catalog query per regime (at most 16). Optional: map `taxpayers_rfc_masked_format` as a backstop. | backend-sr, M2 refactor |
| FU10 | Low | `review_pending_with` exists so the client can warn **before** submitting (PR #56 follow-up), but the regimes step does not use it yet. `reviewPendingPartners` (D3) is the building block. | frontend-sr, M2 hardening |
| FU11 | Low | PR #68 leftovers: fix the probe comment and either reset state per preset or say it mutates. Export one `gen` helper for "operations owned by a module" instead of two copies. | qa-sr, next touch |
| FU12 | Low | M1-16 FU3 (one error mapper), FU7 (`platform/httperr` location) and FU10 (archtest allowlist) were not re-verified in this gate. | backend-sr / tech-lead, as tracked in M1-16 |
