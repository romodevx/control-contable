# QA report: Withholdings (WH-13), 2026-10-14

| | |
|---|---|
| Scope | Automatic ISR/IVA withholdings (MODULE-withholdings, ADR-0008) as merged on `dev` @ `be14173`: kernel, engine 1.1.0, the income and expense forms, the use cases, backfill #142, contract cleanup 0.12.0 #146. Checks: WM goldens WH-01..WH-12 and the owner's example, US-W1 / US-W4 / US-W6 entered through the screens, flags ON/OFF, the backfill dry run on a copy, the legacy `withholding: null` path, the persona moral supplier side |
| Branch | `feature/withholdings-qa-fixes` (from `origin/dev` @ `be14173`) |
| Device walk | **Done** on the Android emulator (`emulator-5554`, visible window), against this run's own API and database (§1). Not blocked |
| Verdict | **Pass after one fix.** Every golden matches to the centavo on screen, through the API and in the estimates. One app defect fixed (F1, the estimate card's IVA label). Two observations deferred (§5) |

## 1. Environment

Docker Desktop was down (`docker info` fails) and cannot be started from the terminal, so `make -C backend db-up` was not usable. Instead, the run used its **own** stack, separate from every other agent's:

- Native PostgreSQL 18.6 (`brew install postgresql@18`; keg-only, its `share`/`lib` were symlinked under `/opt/homebrew` so `initdb` finds them). A fresh cluster in the session scratchpad, `127.0.0.1:5491` (TCP only), database `cc_whqa`. Roles bootstrapped with the repo's own `infra/local/postgres/bootstrap/roles.sql` and random passwords generated at run time (never written to the repo).
- `cmd/migrate up` as `cc_owner`, `cmd/flags apply-preset demo_full`, `cmd/api` as `cc_app` on `:3091` (dev email adapter for the verification code).
- The dev APK (profile, arm64) built with `API_BASE_URL=http://10.0.2.2:3091`, installed on the running emulator (it was idle at the launcher; the app was not installed before).
- Copies for the destructive checks: `cc_whqa_copy` (a `pg_dump` of `cc_whqa`) and a second API on `:3092` against it.

**Wipe**: the database was created empty for this run, so nothing had to be wiped; the only rows before the walk were the migrations' catalogs and flags. No other database, container or process was touched. The integration suites (testcontainers) need Docker, so they ran in CI only.

## 2. Created through the app (no SQL, no seed)

One account (`qa-wh@…`, signed up through *Crear cuenta*, code from the dev email log). The trial allows 5 incomes; the account got the operator's `comp` entitlement (data-model.md; the same as M5-29), the only SQL besides §4's legacy rows.

| Taxpayer (onboarding in the app) | Records entered in the forms |
|---|---|
| Ana RESICO, PF **626** from January 2026 | US-W1: 100,000.00 + IVA, 2026-07-10, "Una empresa", "Servicios profesionales" · WH-05b: 50,000.00 of goods, 2026-08-10 |
| Empresa QA, PM **601** from January 2026 | US-W6: expense 2026-06-15, "Honorarios y servicios profesionales" 40,000.00 + IVA, supplier "Una persona física con negocio o profesión", 612 · WH-06: 10,000.00 of freight · WH-07: 100,000.00 of other business services |
| Beto, PF **612 + 625** (goods and services, RFC given to the platform) | WH-05: 50,000.00 of goods to a company (612) · WH-08: 20,000.00 of scrap (612) · WH-09: 10,000.00 via the platform (625) |

Also exercised: delete a taxpayer (the first RESICO one was created with October as its start month, §5 O1), *Guardar y crear otro* (keeps date and Q1, clears the amount), the error summary on a failed save, the deduction warning toast, History → month → estimate → breakdown.

## 3. Results

### On screen (screenshots `wh-*.png` in `.gstack/qa-reports/screenshots/`, local only)

| Check | Golden | App | Result |
|---|---|---|---|
| US-W1 card | ISR 1,250.00 · IVA 10,666.67 · Recibes 104,083.33, ISR-4 + IVA-6 texts | identical (`wh-01`) | Pass |
| WH-01-EST, July | ISR 750.00, IVA 5,333.33 (16,000.00 − 10,666.67) | identical; breakdown lines "IVA trasladado cobrado 16,000" and "IVA retenido por clientes −10,666.67" (`wh-03`) | Pass |
| WH-05b (US-W4) | ISR 625.00, IVA 0, net 57,375.00; August ISR 0 to pay (1.10 % × 50,000 = 550 < 625) | identical (`wh-02`); August ISR 0 | Pass |
| WH-06 (US-W4) | IVA 400.00 (4 % of the price), net 11,200.00 | identical (`wh-08`) | Pass |
| WH-07 (US-W4) | nothing withheld, net 116,000.00, no "subcontratación" option | identical (`wh-09`); the option list has none | Pass |
| WH-05 / WH-08 / WH-09 (US-W4) | 0 / IVA 3,200.00 net 20,000.00 / ISR 250.00 IVA 800.00 net 10,550.00 | identical (`wh-10`..`wh-12`) | Pass |
| US-W6 card | Debes retener ISR 4,000.00 · IVA 4,266.67, Pagas 38,133.33, Entregas 8,266.67 a más tardar el 17 de julio + deduction note | identical (`wh-04`); E1 not preselected for fees (spec §2.2) | Pass |
| WH-11-JUN | remit 8,266.67; creditable 2,133.33 | "Retenciones a enterar 8,266.67", IVA a favor 2,133.33 (`wh-07`) | Pass |
| WH-11-JUL | credit 4,266.67 in "IVA retenido el mes pasado…" | breakdown identical (`wh-06`); **the estimate card labelled it "IVA acreditable pagado"** (`wh-05`) | **F1, fixed** |
| US-W7 | no "Tu proveedor" for a persona física | none (`wh-17`) | Pass |
| Flags OFF (`module_incomes`, `module_expenses`) | routes 404, entry points hidden | all 18 probes (list, create, preview × both modules × 3 taxpayers, real bearer and ids) `404 not_found`; after relaunch Home hides "Registrar ingreso/gasto" and the month's records (`wh-13`) | Pass |

### Through the use cases (live API, second account)

`createIncome` / `createExpense` with the bodies the app sends, for **WH-01, 02, 03, 03b, 04, 05, 05b, 06, 06b, 07, 08, 09, 10, 11 (expense), 12**: 15 / 15 equal the kernel goldens (ISR, IVA, net, rule ids); WH-12 carries `cpa_unverified: true` and `withholding_cpa_unverified`. July estimate of the 626 taxpayer: ISR 750.00, IVA 5,333.33, total 6,083.33. These are now a CI test (§4).

### Backfill dry run on a copy and the legacy `null` path

In `cc_whqa_copy` only, the US-W1 income and the US-W6 expense were turned back into pre-WH-06/07 records (facts and derivation cleared; the expense's whole IVA creditable), as `incomes_legacy_contract_integration_test.go` does.

| Step | Result |
|---|---|
| Dry run (`cc_worker`) | incomes: 1 to re-derive, 0 different, 1 assumed facts; expenses: the same; ids only, no amounts or user ids. Content hash of every income/expense row identical before and after: **nothing written** |
| Legacy income `GET` (API on the copy) | `withholding: null`, facts served from `counterparty_type` (`pm`, `professional_services`), valid 0.12.0 |
| Legacy income edit (description only) | `200`, derived: ISR-4/IVA-6, 1,250.00 / 10,666.67, `table_version 2026.1` |
| Legacy PM expense edit without supplier facts (API) | `400 supplier_kind required` (expected: the persona moral must answer E1; runbook §1) |
| `-apply`, then dry run again | expense derived (612 assumed, same amounts), creditable split 2,133.33 paid / 4,266.67 withheld, July credit 4,266.67; second dry run **0 to re-derive** in both tables |
| Legacy PM expense edit **in the app** (the US-W6 row made legacy the same way on the main DB) | Pass: the form opens "Tu proveedor" with nothing chosen; saving empty shows "Revisa 1 dato antes de guardar" linked to E1 (`wh-15`); choosing PF 612 and saving derives 4,000.00 / 4,266.67, remit 2026-07-17; History had marked June and July "Cambió" while the row was legacy |

## 4. Fixes and tests on this branch

| Commit | What |
|---|---|
| `fix(mobile): label the regime's creditable IVA without "pagado"` | **F1.** Since engine 1.1.0, `IvaComponents.creditable_paid_cents` is both creditable parts (the IVA paid this month + the IVA withheld last month, `domain/iva.go`). The estimate card printed it as "IVA acreditable pagado", so in July (US-W6) it read "IVA acreditable pagado −$4,266.67" while the breakdown said "IVA acreditable pagado: $0.00". Label is now "IVA acreditable". Regression test first: `estimate_screen_test.dart` "the creditable IVA is not labelled as paid…" (red before, green after). Re-checked on the emulator with a rebuilt APK: July reads "IVA acreditable (incompleto) −$4,266.67" (`wh-18`) |
| `test(withholdings): WM goldens WH-01..WH-12 through the API with estimates` | `backend/cmd/api/withholdings_qa_integration_test.go` `TestWithholdingsQA_goldensThroughTheAPI`: every golden through `createIncome`/`createExpense`, the preview equals what was stored, WH-01-EST, RESICO August at 0, WH-11-JUN/JUL. Coverage gap, not a defect (the requests passed live first); runs in CI (needs Docker) |

Suites: see §6.

## 5. Observations (not fixed; for the owners)

- **O1, onboarding start month is easy to miss.** Step 3 puts "¿Desde qué mes tributas…?" below every regime (about five screens down for a persona física); the walk first created the RESICO taxpayer starting October 2026 by accident and could not enter July. The default follows MODULE-taxpayers §160 (current month); placement is a UX choice for design review, not a withholdings defect.
- **O2, legacy expense E3 shows no answer.** For an expense the backfill has not derived, E3 starts empty while the card already derives with the category default (the server applies it when E3 is omitted, `expense_supplier_facts.dart`). Amounts are right; preselecting the category default for a stored expense with `operation_kind: null` would match the new-expense form. Disappears once the backfill runs.
- `/v1/flags` answered `429` twice right after the API probes' burst from the same address; `200` a minute later. Rate limit as designed, unrelated to withholdings.
- `.gstack/` is git-ignored; this report was added with `git add -f`. The screenshots stay local (4 MB).

## 6. Suites

| Suite | Where | Result |
|---|---|---|
| `go test ./...` (backend unit) | local | Pass (every package) |
| `flutter analyze`, `flutter test` | local | Pass: 0 issues; 1375 tests, 39 skipped (mock- and golden-gated) |
| `make test-integration` incl. the new test | CI | see the PR checks |

## 7. Teardown

The run's API processes (`:3091`, `:3092`) and its PostgreSQL cluster (`:5491`) were stopped by this run; the cluster's data directory was in the session scratchpad. The dev app stays installed on the emulator with the QA account (local data only). `postgresql@18` stays installed via Homebrew for the next QA run.
