# ADR-0010 — The regime's income basis is served by the catalog, embedded in held regimes

| Field | Value |
|---|---|
| Status | ACCEPTED (architect, 2026-10-15; contract 0.16.0) |
| Date | 2026-10-15 |
| Related | [MODULE-incomes](../project-specs/MODULE-incomes.md) §2.2, migration `20261013200000_incomes.sql` (`regimes.income_basis`), contract `IncomeBasis`, `Regime`, `TaxpayerRegime`, M5 QA FU1/D6 |
| Owner | architect / backend-sr / mobile |

## Context

Which date books an income for ISR depends on the regime (MODULE-incomes §2.2): accrual
regimes (personas morales of Título II: 601, 620, 622, 623, 624) book on `invoiced_on`,
every other regime on `collected_on`. The server keeps this as catalog data,
`regimes.income_basis` (NOT NULL, CHECK `cash | accrual`), and derives `incomes.booked_on`
from it in a trigger. The contract did not expose it, so the Flutter income form mirrored
the table in Dart (`_accrualRegimes = {'601','620','622','623','624'}`) to choose which
date to show first and how to label the list. That is a second copy of catalog data, keyed
by code alone while the catalog key is `(code, person_type)`, and it would silently drift
the day a regime's basis changes (the migration comment already requires re-deriving
`booked_on` then; nothing would update the app).

## Options

1. **Keep the Dart table.** No contract change. Rejected: duplicated catalog data, the
   failure mode above, and it contradicts the rule the expenses form already follows
   (`ExpenseCategory.default_operation_kind`, PR #137: the server exposes catalog data the
   form needs instead of the app copying it).
2. **Only on the catalog (`Regime.income_basis`, `GET /v1/regimes`).** One source, but the
   income form works from the taxpayer's held regimes (`Taxpayer.regimes`); it would need a
   second request (and a cache keyed by person type) just to pick a date field.
3. **On the catalog, embedded in `TaxpayerRegime`** (chosen). `Regime.income_basis` is the
   definition; `TaxpayerRegime.income_basis` is the same column read through the join that
   already embeds `name` and `treatment` (catalog on `(code, person_type)`). One source in
   the database, no extra round trip, no Go or Dart copy of the table.

## Decision

1. New frozen schema `IncomeBasis` (`enum [cash, accrual]`). The values are set by law
   (LISR cash vs accrual), so the enum is closed, not append-only; adding a value would be
   a breaking change under CONTRACT.md.
2. `Regime.income_basis` and `TaxpayerRegime.income_basis` are **required** on responses.
   Both schemas are response-only (requests use `TaxpayerRegimeCreate` /
   `TaxpayerRegimeUpdate`), so a new required response member is additive and
   non-breaking: a client that ignores it behaves as before. Required from the first
   release, because the producer ships in the same change and the column is NOT NULL
   since its migration: there is no row without a value, so no backfill precondition.
3. The backend reads the value from `regimes.income_basis` in the existing queries
   (`listRegimes`, the held-regimes join). No Go constant lists the accrual regimes.
4. Contract version 0.15.0 → **0.16.0** (minor: new schema and response fields). Re-frozen:
   `Regime`, `RegimeList`, `TaxpayerRegime`, `Taxpayer`, `TaxpayerList`, `createTaxpayer`,
   `addTaxpayerRegime` (examples and descriptions); new: `IncomeBasis`.
5. The app reads `TaxpayerRegime.income_basis`; an unknown value from a newer server is
   treated as `cash` (the collection date shows first; the server still answers
   `400 required` on `invoiced_on` if it is an accrual regime, so nothing is mis-booked).

## Consequences

- Changing a regime's basis is one catalog migration (which also re-derives
  `incomes.booked_on`); the app follows without a release.
- `TaxpayerRegime` grows by one short string per held regime (at most 16 per taxpayer).
- The income form layout is unchanged by this ADR; only where the basis comes from.
- Rollout order: the backend (0.16.0) deploys before an app built on this client. The
  generated `TaxpayerRegime` and `Regime` require the field, so a 0.16.0 app against a
  0.15.0 server fails to parse every taxpayer; an older app against the new server is
  unaffected (it ignores the member).
- Three spellings of the closed enum (the `regimes_income_basis_valid` CHECK, the incomes
  domain `IncomeBasis`, the contract `IncomeBasis`) are pinned together:
  the domain parses the column, and `TestIncomeBasisMatchesTheIncomesDomain` (cmd/api)
  fails when the contract and the domain disagree.
