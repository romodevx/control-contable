# MODULE-taxpayers — M2 Taxpayers & Onboarding (`module_taxpayers`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-13 · MoSCoW: **Must** · Depends on: M1, P0
Sources: PRD §2, §5, §6 (US-T1..T3), §7, §8, Q4/Q7/Q9/Q14 · MODULES.md M2 · PLAN M2-01..M2-15 ·
research `docs/research/tax-regimes-mx-2026.md` (A.1, A.2.3–A.2.7, B.1, B.2, Part C) ·
`docs/architecture/data-model.md` (taxpayers, taxpayer_regimes, regimes, regime_compatibility) ·
ADR-0002 §1–§2 · `docs/architecture/platform.md` (access policy) · reviews P0-30 (D-3, **D-4**),
M1-17 (#43), M1-16 (FU4, FU6, FU10).

## 1. Scope

A user (login) manages **one or more taxpayers** (contribuyentes, each with its own RFC) under
one subscription. M2 delivers the regime catalog, taxpayer onboarding, taxpayer and regime
maintenance, archive/delete, and the active-taxpayer switcher every later module scopes to.

Owner decisions applied (2026-10-13):

| # | Decision |
|---|---|
| OD-1 | One user manages many taxpayers (RFCs) under one subscription; no per-RFC pricing |
| OD-2 | Cap **10 taxpayers per user for every plan, trial included**; the count is taken under a lock so concurrent creates cannot overshoot (security carry-over) |
| OD-3 | Onboarding order: **persona física / moral FIRST → RFC (+ display name) → regimes → per-regime options** |
| OD-4 | RFC **required**: format + date validated (reject); homoclave check digit verified but **non-blocking** (D-RFC, §9); **no SAT lookup** |
| OD-5 | RFC stored **AES-256-GCM encrypted** + **HMAC blind index**, unique per user; only the masked form is ever returned |
| OD-6 | Year-to-date tax inputs are **optional** at onboarding (US-T4 lives in M5) |
| OD-7 | Dashboard is **per active taxpayer** with a switcher; a combined view is later (Q14, Could) |

| Op | Endpoint | Contract (M2-01) |
|---|---|---|
| listRegimes | `GET /v1/regimes?person_type=fisica\|moral` | exists |
| listTaxpayers | `GET /v1/taxpayers?include_archived=` | exists |
| createTaxpayer | `POST /v1/taxpayers` (taxpayer + regimes + options, one transaction) | exists; add `additionalProperties: false`; the `GOMA800101AB3` example fails módulo 11 (expected digit `6`): replace it with `GOMA800101AB6` |
| getTaxpayer / updateTaxpayer | `GET / PATCH /v1/taxpayers/{taxpayerId}` | exists; add `rfc` correction (OQ-1) |
| deleteTaxpayer | `DELETE /v1/taxpayers/{taxpayerId}?confirm=true` | exists |
| archiveTaxpayer | `POST /v1/taxpayers/{taxpayerId}/archive` | exists |
| unarchiveTaxpayer | `POST /v1/taxpayers/{taxpayerId}/unarchive` | **new — add in M2-01 (OQ-2)** |
| addTaxpayerRegime | `POST /v1/taxpayers/{taxpayerId}/regimes` | exists |
| updateTaxpayerRegime | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | exists |
| archiveTaxpayerRegime | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}?end_period=&confirm=` | exists |

Ports M2 provides to other modules (M2-06):
- `TaxpayerRegimeChecker` (M3, M4, M5-21): "does taxpayer T, owned by the caller, hold regime R
  for period P, and is T not archived?" Records outside `[start_period, end_period]` are refused
  by M3/M4 through this port.
- `TaxpayerProfileReader` (M5): person type, regimes with options and period windows.

## 2. Domain rules

### 2.1 Person type
- `fisica | moral`, chosen first, **immutable** after create (no UPDATE grant on the column;
  composite FK `(id, user_id, person_type)` carries it to `taxpayer_regimes`).
- It filters the catalog and fixes the RFC length.

### 2.2 RFC value object (M2-04, table-driven TDD)

| Rule | Física | Moral |
|---|---|---|
| Length | 13 | 12 |
| Name part | 4 letters `[A-ZÑ&]` | 3 letters `[A-ZÑ&]` |
| Date | `YYMMDD`, a real calendar date (Feb 29 only when `YY % 4 == 0`) | same |
| Homoclave | 2 chars `[A-Z0-9]` | same |
| Check digit | last char `[0-9A]`, SAT módulo 11 — **warning, not rejection** (D-RFC) | same |

- **Normalization** before any check, hash or encryption: trim, upper-case, NFC, strip inner
  spaces and `-`. The client normalizes too (UX: accept the RFC typed with spaces or dashes).
- **Check digit (módulo 11)**: dictionary `0-9 → 0..9, A..N → 10..23, & → 24, O..Z → 25..36,
  space → 37, Ñ → 38`. A moral RFC is left-padded with one space to 13 characters. Sum
  `value(c_i) × (13 − i)` for `i = 0..11`; `r = sum mod 11`; digit = `0` if `r = 0`, `A` if
  `11 − r = 10`, else `11 − r`. The algorithm is cited in the domain code and its tests
  include real published RFCs (SAT examples) plus mutated digits.
- **Check-digit mismatch is non-blocking (D-RFC).** About 1.5% of real SAT-issued RFCs fail
  the published algorithm, so a mismatch never rejects: the taxpayer is created/updated and
  the response carries `warnings: [{code: rfc_check_digit_mismatch, field: rfc, message}]`;
  the app asks the user to double-check. Format, date, charset, length and generic RFCs still
  reject (`400 validation_failed`). The warning is not stored and never echoes the RFC.
- **Generic RFCs are refused** (`XAXX010101000` público en general, `XEXX010101000` extranjero), checked before the check digit because they do not satisfy módulo 11;
  they do not identify a taxpayer.
- Never verified against SAT (PRD NG6).

### 2.3 Regime catalog (`regimes`, `regime_compatibility`, global, seeded in M2-02)
- Key `(code, person_type)`: 626 = RESICO PF **and** RESICO PM; 610 exists for both.
- Every SAT `c_RegimenFiscal` key is **selectable**; `treatment` says what the app can do:

| Treatment | PF | PM |
|---|---|---|
| full | 606, 612, 625 (definitive option), 626 | 601, 626 |
| partial | 621 (RIF, legacy), 625 (non-definitive) | 603 (IVA only), 622 (IVA only, AGAPES deferred) |
| informational | 605, 607, 608, 610, 611, 614, 615, 616 | 610, 620, 623, 624 |

Informational/partial regimes show their plain-language `explanation` at selection time
(text lands in M5-19; M2-02 seeds a placeholder that M5-19 replaces by forward migration).

### 2.4 Per-regime options (`regimes.option_defs`, validated server-side)

| Regime | Option key | Type | Allowed | Default | Required |
|---|---|---|---|---|---|
| 606 PF | `deduction_mode` | enum | `blind_35`, `actual` | `blind_35` | yes |
| 612 PF | `iva_fully_taxable` | boolean | | `true` | yes |
| 625 PF | `activities` | multi_enum (≥ 1) | `transport_delivery`, `lodging`, `goods_services` | — | **yes, no default** |
| 625 PF | `definitive_option` | boolean | | `false` | yes |
| 625 PF | `rfc_given_to_platform` | boolean | | `true` | yes |
| 621 PF | `rif_year` | enum | `1`..`10` | — | **yes, no default** |
| 601 PM | `iva_fully_taxable` | boolean | | `true` | yes |
| 626 PM | `iva_fully_taxable` | boolean | | `true` | yes |
| others | — | | | | |

- Option types: `enum`, `multi_enum`, `boolean`, and `rate` (a `DecimalRate` string bounded by
  the def's inclusive `min`/`max`; no v1 regime uses it yet, the shape is frozen for M5).
- Omitted options with a default are filled server-side; required options with no default
  must be sent. Unknown keys, wrong types and values outside `allowed_values` →
  `400 validation_failed` naming `regimes.N.options.<key>`.
- **Not options**: 601 *coeficiente de utilidad*, YTD totals, PTU, losses, prior payments and 625
  `prior_year_income` are **tax inputs per fiscal year** (ADR-0002 §2; `taxpayer_tax_inputs`,
  M5-20/M5-21). See OQ-3 for how onboarding offers them.
- 606 quarterly option: not supported (research A.2.4 step 5, table not transcribed).

### 2.5 Compatibility (data, not code)
- Two regimes conflict only if their **period windows overlap** (`[start_period, end_period]`,
  open end = ∞). A RESICO PF that exits to 612 from the next month (113-E) is therefore valid:
  626 ends `2026-06`, 612 starts `2026-07`.
- `regime_compatibility` is the **complete matrix** (M2-02): one row per unordered pair of
  the same person type (`regime_a < regime_b`) with `allowed` and a `reason` (shown to the
  user). `allowed = false` → `422 invalid_regime_combination`. A pair whose verdict is not
  verified (`needs_cpa_review`) is **always allowed** (DB CHECK) and the write answers with
  the non-blocking warning `regime_combination_needs_cpa_review` (field
  `regimes.N.regime_code` on create, `regime_code` on add regime; `message` = `reason`).
  A catalog test asserts normalization, that every code exists for that person type, and
  that every reason is non-empty. Seed (⚠️ CPA review before launch, OQ-6), reconciled
  with the law in M2-13 FU1 (migration `20261014120000_regime_compatibility_reconcile.sql`,
  [research Part E](../research/tax-regimes-mx-2026.md)). Authority order: the law, then
  this spec, then the seed. A pair the law does not settle is **allowed with the CPA-review
  warning**, never refused ([CPA-REVIEW](../reviews/CPA-REVIEW.md)). Refused pairs:

| Person type | Pair(s) | Reason (basis) |
|---|---|---|
| fisica | 626 × 612, 626 × 606 | RESICO PF covers business, professional and rental income; it replaces 612/606 (113-E) |
| fisica | 616 × any | "Sin obligaciones" cannot coexist with an obligation |
| fisica | 610 × any resident regime | A non-resident without PE is not a resident taxpayer |
| fisica | 621 × 612 | RIF was only for business income without a professional degree; leaving RIF means 612 (LISR art. 111 ¶1, ¶15 as in force until 2021) |
| fisica | 621 × 626 | A RIF taxpayer kept RIF or moved to RESICO; RESICO only admits salary and interest besides (LISR 113-E ¶1, ¶6; Transitorio 2º fr. IX DOF 12-11-2021) |
| fisica | 626 × 611 | Shareholders may not use RESICO (113-E) |
| moral | 626 × 601, 603, 622, 623, 624 | RESICO PM exclusions and exit to Título II (arts. 206, 214) |
| moral | 620 × 601, 622, 623, 624 | The deferring cooperative applies its chapter *instead of* Título II (art. 194 ¶1) |
| moral | 620 × 626 | Título VII Cap. VII is excluded from RESICO PM (art. 206 fr. IV) |
| moral | 622 × 601 | An AGAPES PM must use Cap. VIII (art. 74 ¶1) |
| moral | 622 × 623 | Cap. VIII does not apply to Cap. VI groups (art. 74 ¶2) |
| moral | 622 × 624 | Each requires ≥ 90 % of income from its own activity (arts. 72, 74) |
| moral | 610 × any | Non-resident (Título V) |

Pairs the law does not settle stay **allowed with a warning** (not refused): PM 601 × 603,
601 × 623, 601 × 624, 603 × 620, 603 × 622, 603 × 623, 603 × 624, 623 × 624; PF 621 and 626
with the other Título IV regimes listed in CPA-REVIEW. "A legal entity holds exactly one
ISR regime" is **not** a rule of the law (e.g. a Título III entity owes Título II tax on
some income, art. 80; a Cap. VI group company stays under Título II, art. 59 ¶4).

### 2.6 Periods
- `start_period` defaults to the **current month** (America/Mexico_City, server clock);
  allowed range `2022-01` (catalog floor) … current month.
- `end_period` (archiving a regime): `start_period ≤ end_period ≤ current month`, and
  **≥ the latest period that has a record** in that regime (no orphaned records outside
  the window).
- Re-adding an archived regime **reopens the same row** (PK `(taxpayer_id, regime_code)`):
  `end_period` is cleared and `start_period` kept; the gap holds no records (OQ-5).

### 2.7 Archive vs delete

| Action | Taxpayer | Regime |
|---|---|---|
| Archive | `archived_at` set. Hidden from the list and the switcher by default; `GET` by id still works; estimates/history for past periods stay readable. **Writes refused**: new or edited incomes/expenses, regime changes, rename (`422 business_rule_violation`, detail `taxpayer_archived`). Still counts toward the cap; RFC stays reserved. Reversible (unarchive). | `end_period` set; records kept and still counted in estimates for periods ≤ `end_period`; new records after `end_period` refused by M3/M4. Needs `confirm=true` when records exist (`409 regime_has_records`). The last active regime cannot be archived (`422`, detail `last_active_regime`): archive the taxpayer instead. |
| Delete | **Irreversible**, `confirm=true` required. Hard delete cascades to `taxpayer_regimes`, `incomes`, `expenses`, `taxpayer_tax_inputs`, `estimate_views` (M3/M4/M5/M8 FKs `ON DELETE CASCADE` from the taxpayer). `user_usage` counters are **not** decremented (PRD §7: deleting never frees a trial slot). Frees one cap slot and the RFC. | Not offered: DELETE on a regime archives (MODULES M2). |

### 2.8 Cap of 10 (OD-2)
- Counted per user over **all** taxpayers, archived included (archiving must not become a way
  around the cap); deleting frees a slot.
- Enforcement in `CreateTaxpayer`, inside the transaction, before the insert:
  `SELECT … FROM users WHERE id = app_user_id() FOR UPDATE` (serializes this user's creates),
  then `count(*)` of the user's taxpayers, then `access.Policy.CheckTaxpayerLimit`. Not an
  entitlement: paying does not lift it (`422 taxpayer_limit_reached`, not 402).
- Idempotent retry (client-supplied `id` already owned by the caller) returns `200` with the
  existing taxpayer **before** the cap check, so a retry at 10/10 never fails.
- "Raised on request" (PRD Q4) is out of scope for v1 (OQ-4).

### 2.9 Active taxpayer (client state)
- No server endpoint: the selected `taxpayerId` is stored on the device per user
  (non-secret preference) and sent as the path parameter of every M3–M9 call.
- On launch: stored id valid and not archived → use it; else the first non-archived taxpayer
  by `created_at`; none → onboarding.
- The switcher (from any main screen) lists non-archived taxpayers: display name, person type,
  masked RFC. Switching reloads the current screen for the new taxpayer.

### 2.10 Read-only access (OQ-8, DECIDED)
- Every M2 mutation except delete and archive goes through `access.Policy` in its use case:
  with no running trial and no `pro`/`comp` entitlement the user is **read-only**.

| Operation | Read-only user |
|---|---|
| list/get taxpayers, list regimes | allowed |
| archive taxpayer, archive regime (`DELETE …/regimes/{code}`), delete taxpayer | allowed |
| create taxpayer, unarchive, add regime, rename, `PATCH rfc`, `PATCH` options | `402 entitlement_required` |

- The policy check runs **before** the cap check and before validation of business rules
  (schema validation at the edge still answers 400 first). An idempotent create retry whose
  `id` already exists returns `200` regardless (no write happens).
- Client: these actions open the paywall (UX only); onboarding a first taxpayer after expiry
  is therefore paywalled too.

## 3. User stories (Given / When / Then)

### US-T1 Onboard a persona física
- **Given** a signed-in user with no taxpayers and `module_taxpayers` ON, **when** the app
  opens, **then** onboarding starts at step 1 "¿Persona física o persona moral?" and nothing
  else is asked before that choice.
- **Given** I chose física, **when** I enter `display_name` and a 13-char RFC with a valid
  check digit, **then** step 3 shows only `GET /v1/regimes?person_type=fisica`, with the
  informational ones labeled and explained.
- **Given** I picked 606 and 625, **when** step 4 shows their options, **then** 606 is
  prefilled `blind_35`, 625 asks for `activities` (none preselected) and shows
  `definitive_option=false`, `rfc_given_to_platform=true`.
- **When** I finish, **then** one `POST /v1/taxpayers` creates the taxpayer, both regimes and
  options atomically → `201`, the response carries `rfc_masked` (`GOMA******AB6`) and never
  the RFC, and the new taxpayer becomes the active one.
- **Given** the request fails mid-transaction, **then** nothing is persisted (no taxpayer
  without regimes).

### US-T1b Onboard a persona moral
- **Given** I chose moral, **when** I enter a 12-char RFC, **then** the regime list is the
  moral catalog; picking 601 asks only `iva_fully_taxable` (default true); `201` on finish.
- **Given** moral, **when** I select two regimes the law excludes (e.g. 601 + 626), **then**
  `422 invalid_regime_combination` with the catalog reason; a pair under CPA review (e.g.
  601 + 603) is `201` with `regime_combination_needs_cpa_review` (§2.5). The client keeps the
  moral list single-select (UX only, the common case); the server is the authority.

### US-T1c RFC validation
| Given | Then |
|---|---|
| física + 12-char RFC, or moral + 13-char RFC | `400 validation_failed`, `rfc` / `rfc_length_mismatch` |
| wrong check digit (e.g. last char changed) | `201` (or `200` on PATCH), `warnings[0]` = `rfc` / `rfc_check_digit_mismatch` (D-RFC) |
| impossible date (`ABCD991332XX1`) | `400`, `rfc` / `rfc_date` |
| characters outside the alphabet | `400`, `rfc` / `rfc_format` (edge pattern) |
| `XAXX010101000` or `XEXX010101000` | `400`, `rfc` / `rfc_generic` |
| `" goma-800101 ab6 "` (spaces, dashes, lower case) | normalized, accepted if otherwise valid |
| any of the above | the RFC value never appears in the error message, logs or traces |

### US-T2 Add more taxpayers, duplicate RFC, cap
- **Given** I already manage RFC X (active or archived), **when** I create another taxpayer
  with X in any spelling that normalizes to X, **then** `409 rfc_already_registered`; if the
  existing one is archived, the message suggests unarchiving it.
- **Given** another user manages RFC X, **when** I create X, **then** `201` (uniqueness is per
  user; nothing reveals the other user).
- **Given** I hold 10 taxpayers (archived included), **when** I create an 11th, **then**
  `422 taxpayer_limit_reached`.
- **Given** I hold 9, **when** 5 creates arrive concurrently, **then** exactly 1 succeeds and 4
  get `422` (integration test with real PostgreSQL, parallel goroutines).
- **Given** I retry a create with the same client `id`, **then** `200` with the existing
  taxpayer, even at 10/10. A client `id` that exists for another user → `409 conflict`
  (generic; RLS hides the row, the PK collision must not leak ownership).

### US-T3 Add a regime / end a regime
- **Given** a física with 612 active, **when** I add 626 starting the same month, **then**
  `422 invalid_regime_combination` naming 626 × 612 and the reason.
- **Given** 626 ended `2026-06`, **when** I add 612 with `start_period 2026-07`, **then** `201`.
- **Given** 612 already active, **when** I add 612 again, **then** `409 conflict`.
- **Given** a física, **when** I add a moral-only code (601), **then**
  `400 validation_failed` (`regime_code` / `not_for_person_type`); the composite FK `(regime_code, person_type)` is the DB backstop.
- **Given** 606 has records up to `2026-08`, **when** I `DELETE …/regimes/606?end_period=2026-09`
  without `confirm`, **then** `409 regime_has_records`; with `confirm=true` → `200`,
  `end_period: 2026-09`, records kept.
- **Given** the same, **when** `end_period=2026-07`, **then** `422 business_rule_violation`
  (`end_period` / `records_after_end_period`).
- **Given** 612 is the only active regime, **when** I archive it, **then** `422
  business_rule_violation` (`last_active_regime`).
- **Given** 606 archived, **when** I add 606 again, **then** the row reopens (`end_period: null`).

### US-T3b Per-regime options
- **Given** 606 with `blind_35`, **when** I PATCH `{options: {deduction_mode: actual}}`, **then**
  `200`; the client warns that the change applies to every period of the regime (OQ-7).
- **Given** 625, **when** I PATCH `activities: []` or an unknown value, **then** `400` naming
  `options.activities`.
- **Given** 621 without `rif_year`, **when** I create, **then** `400` naming
  `regimes.N.options.rif_year`.

### US-T3c Rename, archive, unarchive, delete
- **When** I PATCH `display_name` (1–80 chars, trimmed, NFC), **then** `200`; `person_type`
  in the body is rejected (`400`, `additionalProperties`).
- **When** I PATCH `rfc` with a valid RFC of the same person type (OQ-1), **then** `200`, it is
  re-encrypted with the row-id AAD and re-hashed; a duplicate under my account → `409
  rfc_already_registered`; a wrong length for the type → `400 rfc_length_mismatch`.
- **When** I archive the active taxpayer, **then** `200`, it leaves the switcher and the app
  switches to the next non-archived one (or onboarding if none).
- **Given** an archived taxpayer, **when** I add an income, a regime, or rename it, **then**
  `422 business_rule_violation` (`taxpayer_archived`); **when** I read its past estimates or
  history, **then** `200`.
- **When** I unarchive, **then** `200` and it returns to the switcher (no cap check: it already
  counts).
- **When** I DELETE without `confirm=true`, **then** `400 validation_failed`. **With** it, after
  an in-app confirmation that names the taxpayer and states "se borrarán sus ingresos y
  gastos", **then** `204`; its incomes, expenses, tax inputs and history are gone; my trial
  counters are unchanged; the RFC can be registered again.

### US-T4 Switch the active taxpayer
- **Given** two taxpayers, **when** I pick B in the switcher on Home, **then** Home, incomes,
  expenses and estimates show B; after an app restart B is still selected.
- **Given** the stored active taxpayer was deleted on another device, **when** the app opens,
  **then** it falls back silently to the first non-archived taxpayer.

### US-T4b Read-only user (OQ-8)
- **Given** my trial ended and I have no plan, **when** I create a taxpayer, unarchive one,
  add a regime, rename, correct an RFC or change options, **then** `402 entitlement_required`
  and nothing changes.
- **Given** the same, **when** I list, open, archive a taxpayer or regime, or delete a
  taxpayer (with `confirm=true`), **then** it succeeds as for a paying user.
- **App (M2-08)**: once a write has answered `402` in this session, opening the onboarding
  shows the paywall before step 1 instead of four steps that cannot be saved. Knowing it
  before the first write needs an access read the frozen contract lacks: M7 follow-up in
  MODULES.md.

### US-T5 Cross-user isolation
- **Given** user B's taxpayer id, **when** user A calls any M2 endpoint with it, **then** `404
  not_found` (never 403), for every verb.

## 4. Errors → ErrorCode

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Schema/pattern violation, unknown field | 400 | `validation_failed` | per field |
| RFC length vs person type / format / date / generic | 400 | `validation_failed` | `rfc_length_mismatch`, `rfc_format`, `rfc_date`, `rfc_generic` |
| Regime not in catalog for the person type | 400 | `validation_failed` | `not_for_person_type` |
| Bad option key/type/value, missing required option | 400 | `validation_failed` | `unknown_option`, `invalid_option`, `required` |
| `start_period` / `end_period` out of range | 400 | `validation_failed` | `out_of_range` |
| Delete without `confirm=true` | 400 | `validation_failed` | `confirm_required` |
| Missing/invalid session | 401 | `unauthorized` | |
| Not owner, unknown id, flag OFF | 404 | `not_found` | |
| Duplicate RFC for this user | 409 | `rfc_already_registered` | |
| Regime already active; client id owned by someone else | 409 | `conflict` | |
| Archive regime with records, no confirm | 409 | `regime_has_records` | |
| Overlapping incompatible regimes | 422 | `invalid_regime_combination` | `incompatible` (+ pair) |
| 11th taxpayer | 422 | `taxpayer_limit_reached` | |
| Write on archived taxpayer; last active regime; records after end | 422 | `business_rule_violation` | `taxpayer_archived`, `last_active_regime`, `records_after_end_period` |
| Read-only user: create, unarchive, add regime, rename, PATCH rfc/options | 402 | `entitlement_required` | |
| Per-user write budget exceeded | 429 | `rate_limited` | |
| Anything else | 500 | `internal_error` | |

No new `ErrorCode` values are needed; detail codes are added to the contract descriptions in M2-01.
A check-digit mismatch is not an error: `Taxpayer.warnings` carries `rfc_check_digit_mismatch` (D-RFC).
An overlapping regime pair still under CPA review is not an error either: `Taxpayer.warnings`
(create) or `TaxpayerRegime.warnings` (add regime) carries `regime_combination_needs_cpa_review` (§2.5).

## 5. Security requirements (gate M2-14)

| # | Requirement | Test |
|---|---|---|
| S-1 | **D-4 (P0-30, blocking):** RFC ciphertext AAD = `key_id ‖ taxpayer.id ‖ user_id`. The use case generates the row id (uuidv7) **before** encrypting; it never relies on the DB default | A ciphertext copied to another row (same user, and another user) fails to decrypt |
| S-2 | Blind index `rfc_hash = HMAC-SHA256(k_rfc_index, user_id ‖ normalized RFC)`; key separate from the encryption key and from the email-index key; including `user_id` prevents cross-user linkage of the same RFC | Same RFC, two users → different hashes; same user, two spellings → same hash |
| S-3 | `rfc_masked` derived server-side (PF `AAAA******XX9`, PM `AAA******XX9`); the full RFC never leaves the server (no endpoint returns it) | Contract + response-body scan |
| S-4 | RFC never logged, traced or put in error messages; redaction test on the create path at debug level | Log capture test |
| S-5 | Cap count under `SELECT … FOR UPDATE` on the user row in the create transaction (OD-2) | Concurrency test (§3 US-T2) |
| S-6 | RLS ENABLE + FORCE on `taxpayers`, `taxpayer_regimes`; owner policies; composite FKs `(taxpayer_id, user_id, person_type)` and `(regime_code, person_type)`; no UPDATE grant on `person_type`; `rfc_*` updated only by the RFC-correction use case (OQ-1) | Isolation suite (data-model §RLS) incl. cross-user composite-FK rejection |
| S-7 | Account deletion cascades every M2 table; extend the deletion-cascade guard to indirect chains (M1-17 #43) | Guard test lists M2 tables |
| S-8 | Repositories never commit/rollback or `SET` (P0-30 D-3 lint lands with the first M2 repository) | archtest |
| S-9 | Audit log events `taxpayer.created/rfc_corrected/archived/unarchived/deleted`, `taxpayer_regime.added/archived` with user id + taxpayer id only | Log capture test |
| S-10 | Per-user write budget on taxpayer/regime mutations (default 30/h) via the ADR-0005 limiter | 429 test |
| S-11 | Contract validation at the edge, `additionalProperties: false` on every M2 request body | Contract tests |

## 6. Feature flag `module_taxpayers`

- Default OFF; registered by migration (M2-11).
- OFF → every M2 route returns `404 not_found` (server-side boundary), including
  `GET /v1/regimes`.
- Dependency: M3, M4, M5, M6, M8, M9 need a taxpayer, so with `module_taxpayers` OFF their
  routes also answer `404` even if their own flag is ON (implied OFF, same mechanism as M5→M6/M8).
- Client (single `useFlag`/FlagsProvider equivalent in Flutter): OFF → no onboarding, no
  switcher, Home shows only the account area. UX only.
- Account deletion (M1) still cascades taxpayers when the flag is OFF.
- Suites run with the flag ON and OFF (M2-07); presets `demo_core`/`demo_monetization`/`demo_full`
  include it.

## 7. UX notes (rules/ui-forms-and-modals.md)

- Onboarding is a 4-step page flow (more than 6 fields across steps), single column, one
  decision per step; step 1 is two large choices (≥ 44 px).
- RFC field: `autocapitalize=characters`, no autocorrect, accepts spaces/dashes, format errors
  on blur once typed; the check-digit message is plain Spanish ("El RFC no es válido; revisa
  la homoclave").
- Regime picker: multi-select for física, single-select for moral; informational regimes show
  a short explanation before selection.
- Delete taxpayer is irreversible → confirmation dialog naming the taxpayer. Archive is
  reversible → do it and offer Undo (persistent, not auto-dismissed).

## 8. Out of scope

- SAT verification of the RFC or of the taxpayer's regimes (PRD NG6); Constancia de Situación
  Fiscal import; CURP.
- Combined "all my taxpayers" dashboard (Q14, Could).
- Sharing a taxpayer between users, accountant roles (NG7).
- Raising the cap per user (OQ-4); tax inputs UI (M5-21); 606 quarterly option; AGAPES
  treatment; RESICO expulsion/eligibility automation (prompted annually, later).
- Versioned (time-ranged) regime options (OQ-7).

## 9. Decisions (coordinator, 2026-10-13) — all DECIDED

| # | Question | Decision |
|---|---|---|
| OQ-1 | Correct a mistyped RFC (same person type) instead of delete + recreate? | **DECIDED: yes.** `PATCH rfc` re-validates, re-encrypts (S-1), re-hashes, rechecks uniqueness; contract change in M2-01 (today's contract says immutable). `rfc_*` columns get an UPDATE grant only through this use case |
| OQ-2 | Add `POST …/unarchive`? | **DECIDED: yes** (otherwise an archived RFC blocks re-creation) |
| OQ-3 | Where do the 601 coeficiente / YTD inputs go in onboarding? | **DECIDED:** optional, skippable step 5 "Datos para el cálculo", shown only when `module_tax_engine` is ON, writing through M5's tax-inputs endpoint; otherwise the estimate shows the `incomplete` CTA |
| OQ-4 | How is the cap "raised on request"? | **DECIDED: not in v1.** Fixed `TaxpayerCap = 10`; a per-user override column when the first request arrives |
| OQ-5 | Re-adding an archived regime | **DECIDED: reopen the row** (PK unchanged); revisit if users switch back and forth |
| OQ-6 | Compatibility matrix (§2.5) | **DECIDED:** ship the seed as written, ⚠️ marked; CPA review before launch; 625 × 626 allowed. **Amended (M2-13 FU1):** "PM single-regime" is a client UX default, not a server rule; the server refuses only the pairs the law excludes (§2.5) and warns on the rest |
| OQ-7 | Option changes mid-year | **DECIDED:** not versioned in v1, UI warns; CPA to confirm whether to lock per fiscal year |
| OQ-8 | What may a read-only user (trial expired, no `pro`/`comp`) do in M2? | **DECIDED (overrides the PM default): read, delete and archive only.** Creating a taxpayer, unarchiving, adding a regime, renaming, correcting the RFC and changing options are refused with `402 entitlement_required` (`ErrReadOnly`), consistent with the access policy. Archiving a regime is an archive and stays allowed. See §2.10 |
| OQ-9 | Does an archived taxpayer count toward the cap? | **DECIDED: yes**; delete frees the slot |
| D-RFC | Reject an RFC whose módulo 11 check digit does not match? | **DECIDED (coordinator, 2026-10-13): no.** ~1.5% of real SAT-issued RFCs fail the published algorithm. A mismatch is accepted with a non-blocking `warnings` entry (`rfc_check_digit_mismatch`, field `rfc`) on `createTaxpayer`/`updateTaxpayer`; format, date, charset, length and generic RFCs still reject. Frozen in the M2-01 contract (0.4.0) |
