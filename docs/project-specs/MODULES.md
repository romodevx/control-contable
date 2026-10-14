# MODULES — Control Contable

| Field | Value |
|---|---|
| Status | DRAFT (Phase 1 Discovery), pending owner approval |
| Date | 2026-10-13 |
| Source | [PRD.md](PRD.md) |

The modules below are vertical slices, listed in delivery order. Each one runs the full
cycle: spec → OpenAPI → data+RLS → backend TDD → Flutter → flag → QA → gates → docs.
Flags default OFF, and a disabled module's routes return **404** server-side. Endpoints
are under `/v1` and are provisional: the Architect freezes them in `docs/api/openapi.yaml`.
"RLS" marks user-scoped tables (ENABLE + FORCE, keyed on `user_id`). "Global" marks
read-only reference data with no RLS.

## Platform vs. product

| Layer | Item | Flag | Notes |
|---|---|---|---|
| Platform | P0 Foundation | none (always on) | Not a product module; it is what every module runs on |
| Product | M1–M9 | `module_<name>` | Can be shipped and demoed on their own |

## P0 — Foundation (platform, no flag)

- **Scope**: monorepo skeleton (`backend/`, `mobile/`, `docs/`); config; the error
  envelope; health checks; versioned migrations; the per-transaction RLS session context
  (`SET LOCAL app.current_user_id`); the flag service and middleware (404 when OFF) with
  presets; the Flutter `FlagsProvider` / `useFlag` equivalent; the **access policy port**
  (trial clock + per-user usage counters + `pro` entitlement lookup, including comp); CI
  (lint, tests, contract tests, govulncheck, a release-build artifact check).
- **Endpoints**: `GET /healthz`, `GET /v1/flags` (enabled modules for the client).
- **Tables**: `feature_flags`, `flag_presets` (global); `users` (RLS; identity root, moved here from M1 per PLAN decision 1 so P0 FKs resolve); `user_usage` (RLS; created-records counters); `entitlements` (RLS; trial/pro/comp, end date).
- **Depends on**: nothing.

## Product modules (delivery order)

| # | Module | Flag key | MoSCoW | Depends on |
|---|---|---|---|---|
| M1 | Auth & Account | `module_auth` | Must | P0 |
| M2 | Taxpayers & Onboarding | `module_taxpayers` | Must | M1 |
| M3 | Incomes | `module_incomes` | Must | M2 |
| M4 | Expenses | `module_expenses` | Must | M2 |
| M5 | Tax Engine (estimate + breakdown) | `module_tax_engine` | Must | M2, M3, M4 (reads via ports) |
| M6 | Dashboard | `module_dashboard` | Must | M5 |
| M7 | Subscriptions (RevenueCat) | `module_subscriptions` | Must | M1, P0 access policy |
| M8 | History | `module_history` | Should | M5 |
| M9 | Reminders (push) | `module_reminders` | Should | M2 (M5 optional for amounts) |

M3 and M4 can run in parallel once the M2 contract is frozen. The M5 strategies are pure
domain code and can start right after the research doc lands, in parallel with M3/M4. M7
can run in parallel with M5/M6.

### M1 — Auth & Account (`module_auth`)
- **Scope**: register (with privacy and terms consent), login, refresh with rotation and
  reuse detection, logout, user profile, in-app account deletion, current legal documents.
  The trial starts at registration (writes to `entitlements`).
- **Endpoints**: `POST /v1/auth/register` (202 + emailed code), `POST /v1/auth/verify-email` (creates the account, issues tokens; ADR-0001 OWNER 1, option A), `POST /v1/auth/login`, `POST /v1/auth/refresh`,
  `POST /v1/auth/logout`, `GET|PATCH /v1/me`, `DELETE /v1/me`, `GET /v1/legal/documents`.
- **Tables**: `refresh_tokens` (RLS), `consents` (RLS), `legal_documents` (global), `trial_registry` (keyed email hash; no RLS, service-only access), `email_verifications` (pending sign-ups; no RLS, service-only access). `users` is created in P0; M1 owns its behaviour.
- **Out of scope**: social login, 2FA, password reset by SMS (email reset is Should).

### M2 — Taxpayers & Onboarding (`module_taxpayers`)
- **Scope**: the regime catalog and compatibility rules (data from research); taxpayer
  CRUD; onboarding order **física/moral → RFC → regimes → per-regime options**; active
  taxpayer switcher; archiving a regime or taxpayer.
- **Endpoints**: `GET /v1/regimes?person_type=fisica|moral`,
  `GET|POST /v1/taxpayers`, `GET|PATCH|DELETE /v1/taxpayers/{taxpayerId}`,
  `POST /v1/taxpayers/{taxpayerId}/archive`,
  `POST /v1/taxpayers/{taxpayerId}/regimes`,
  `PATCH|DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` (DELETE archives the regime, with an end period).
- **Tables**: `taxpayers` (RLS; `rfc_encrypted`, `rfc_hash` unique per user, `person_type`, `display_name`, `archived_at`), `taxpayer_regimes` (RLS; options, start/end period), `regimes`, `regime_compatibility` (global).

### M3 — Incomes (`module_incomes`)
- **Scope**: create/list/edit/delete with undo for incomes per taxpayer + regime; IVA rate;
  withholdings; `invoiced_on` + `collected_on`; trial-cap check through the access policy port.
- **Endpoints**: `GET|POST /v1/taxpayers/{taxpayerId}/incomes?period=YYYY-MM&regime=`,
  `GET|PATCH|DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}`,
  `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore`.
- **Tables**: `incomes` (RLS; `taxpayer_id`, `regime_code`, `amount_cents`, `iva_rate`, `iva_cents`, `isr_withheld_cents`, `iva_withheld_cents`, `invoiced_on`, `collected_on`, `deleted_at`).

### M4 — Expenses (`module_expenses`)
- **Scope**: the same CRUD as M3, plus the category catalog with per-regime deductibility
  hints, the payment method and the warning on cash payments above the threshold.
- **Endpoints**: `GET /v1/expense-categories?regime=`,
  `GET|POST /v1/taxpayers/{taxpayerId}/expenses?period=YYYY-MM&regime=`,
  `GET|PATCH|DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}`,
  `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore`.
- **Tables**: `expenses` (RLS), `expense_categories`, `category_regime_rules` (global).

### M5 — Tax Engine (`module_tax_engine`)
- **Scope**: a `RegimeStrategy` interface (one strategy per regime: full, partial or
  informational); versioned parameter sets per fiscal year; tax inputs (YTD totals,
  coeficiente de utilidad, losses, prior payments); the estimate per regime and per
  taxpayer; the breakdown with legal basis; the **legend as a response field**; the golden
  test suite as a CI gate. It reads incomes and expenses through read ports and does not
  import M3/M4 code.
- **Endpoints**: `GET /v1/taxpayers/{taxpayerId}/estimates?period=YYYY-MM`,
  `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown?regime=`,
  `GET|PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs?fiscal_year=`.
- **Tables**: `taxpayer_tax_inputs` (RLS). Tax parameters (tariffs, rates) are versioned YAML embedded in the binary, not tables (ADR-0002; supersedes the earlier `tax_parameter_sets` / `tax_tables`).

### M6 — Dashboard (`module_dashboard`)
- **Scope**: Home for the active taxpayer: this month's estimate, income vs. expenses,
  6–12 month trend, trial banner, empty state. Read-only aggregation.
- **Endpoints**: `GET /v1/taxpayers/{taxpayerId}/dashboard?period=YYYY-MM&months=12`.
- **Tables**: none. It reads through M5 and the record read ports (indexes on `(user_id, taxpayer_id, collected_on)`).

### M7 — Subscriptions (`module_subscriptions`)
- **Scope**: RevenueCat integration (app user ID = user ID); webhook ingestion
  (authenticated with a shared secret, idempotent by event ID, safe for out-of-order
  events); post-purchase sync through the RevenueCat REST API; plan status; it writes
  `pro` into `entitlements`. Products: weekly, monthly, yearly → entitlement `pro`.
- **Endpoints**: `GET /v1/subscription`, `POST /v1/subscription/sync`,
  `POST /v1/webhooks/revenuecat` (no user JWT; bearer secret; dedicated DB role).
- **Tables**: `subscriptions` (RLS), `revenuecat_events` (service-only; unique `event_id`, raw payload, processed_at).
- **When OFF**: the trial rules in P0 still apply and nobody can buy. Demos use comp entitlements.
- **Contract follow-up from M2 (PR #58)**: the app cannot know a user is read-only before a
  write: `GET /v1/me` (frozen) carries no access and `GET /v1/subscription` is gated by this
  flag, while the trial rules apply with it OFF. M7 must expose `AccessSummary` to the app
  regardless of `module_subscriptions` (an additive field on `GET /v1/me`, or an always-on
  read; ADR for the frozen-schema change). Until then the onboarding shows the paywall up
  front only after a `402` in the same session (`readOnlyAccessProvider`, mobile).

### M8 — History (`module_history`)
- **Scope**: monthly summaries for a fiscal year; open any month's breakdown (reuses M5); a "changed since last viewed" marker.
- **Endpoints**: `GET /v1/taxpayers/{taxpayerId}/history?fiscal_year=`, `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed`.
- **Tables**: `estimate_views` (RLS; period, records fingerprint, viewed_at).

### M9 — Reminders (`module_reminders`)
- **Scope**: device token registration; reminder settings; the tax calendar (due date
  moved for weekends and holidays); a scheduled worker that sends one push per user,
  listing their taxpayers; amounts only with an explicit opt-in.
- **Endpoints**: `POST /v1/devices`, `DELETE /v1/devices/{deviceId}`,
  `GET|PUT /v1/reminder-settings`, `GET /v1/tax-calendar?year=`.
- **Tables**: `devices` (RLS), `reminder_settings` (RLS), `notification_log` (RLS), `holidays` (global).
- **Also (from M1-10, ADR-0005)**: the worker purges full `rate_limit_buckets` rows
  (`tat` in the past) on a schedule; cc_worker needs a narrow grant for it.

## Mapping to the original brief

| Brief endpoint | Now |
|---|---|
| auth register/login/refresh | M1 `/v1/auth/*` |
| profile | M1 `/v1/me` + M2 `/v1/taxpayers` |
| income list/create/delete | M3, scoped by taxpayer |
| expense list/create/delete | M4, scoped by taxpayer |
| GET calculations/estimate | M5 `/v1/taxpayers/{id}/estimates` (+ breakdown) |
| subscription status / checkout | M7 status + RevenueCat IAP (checkout is replaced by the store sheet) |

## Presentation presets (flag sets)

Each preset has an E2E smoke test.

| Preset | Flags ON |
|---|---|
| `demo_core` | auth, taxpayers, incomes, expenses, tax_engine, dashboard |
| `demo_monetization` | `demo_core` + subscriptions |
| `demo_full` | all M1–M9 |
