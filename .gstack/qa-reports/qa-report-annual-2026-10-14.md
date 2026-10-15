# QA report: AN Annual estimate (ANN-23), Android emulator, 2026-10-14/15

| | |
|---|---|
| Scope | Module AN (`module_annual`, MODULE-annual.md) as merged on `dev` through #150 (ANN-22). Checks: UI figures equal the AN-01..AN-08 goldens; flag ON/OFF, including a prerequisite OFF; read-only users and the paywall on writes; persona física (PF) and persona moral (PM); "Falta" CTAs on incomplete sections; deduction caps and the 500-a-year limit; the filed-payments confirm flow; the disclaimers on every estimate |
| Branch | `feature/annual-qa-fixes` (from `origin/dev` @ `36cac35`), PR #153 |
| Device | Android emulator `emulator-5554` (`Medium_Phone_API_36.1`), already running and reused. Before installing, I asked the other agent working in parallel (`wh07-expenses`) whether it was using the emulator; it said it never had |
| App | dev flavor, **profile** build for arm64, `--dart-define=API_BASE_URL=http://10.0.2.2:3093`. The previously installed dev app was uninstalled first, so the walk started with no app data |
| Stack | This run's **own** PostgreSQL 18 (Homebrew `postgresql@18` binaries): its own cluster in the session scratchpad, TCP on loopback only, `127.0.0.1:5493`, database `cc_annqa`, roles from `infra/local/postgres/bootstrap/roles.sql` with fresh random passwords (`init-env.sh`), `cmd/migrate up` (every migration), `cmd/flags apply-preset demo_full`. API `cmd/api` (APP_ENV=dev) as `cc_app` on `:3093`. Docker Desktop stayed down |
| iOS | **Not run** (OB6, as in every earlier module QA) |
| Evidence | 43 screenshots `NN-*.png` in `.gstack/qa-reports/screenshots/` on the QA machine (git-ignored) |
| Verdict | **Pass, one accessibility defect fixed** (§4). Every figure on screen matches the arithmetic by hand and the server. All suites green (§5) |

## 1. What was wiped

Nothing shared was touched. The database was **created empty** for this run (new cluster, new
`cc_annqa`, 0 users). After migrating, `demo_full` turned all 10 modules ON, and `GET /v1/flags` listed
all 10. The installed app was uninstalled before this run's APK was installed.

## 2. What was created through the UI

| Account / taxpayer | Created in the app |
|---|---|
| *Ana QA Anual* | Sign-up, then the emailed code (read from the dev API log). Accepted the privacy notice and the terms |
| *Ana Honorarios* (PF) | 612 (IVA fully taxable) + 605. The start month was left at the default, **octubre de 2026** (I missed the month picker), so nothing in 2026 falls inside the regime window. Opening its annual estimate showed a $0 section, `incomplete`, and "Falta: Constancias de sueldos"; that CTA was followed (§3) |
| *Ana Despacho 2026* (PF) | 612 + 605 **from enero de 2026**, chosen in the month picker. **Incomes** (612, persona moral, honorarios, 16 %, the app computed 10 % ISR withheld): 15/02 100,000.00 · 15/05 80,000.00 · 15/08 60,000.00 (*Guardar y crear otro*). **Expense** (612, rent, transfer, with a CFDI): 10/03 30,000.00. **Constancia** *Empresa A*: gravados `300,000.00`, exentos `20000`, ISR `45000.5` (flexible formats accepted). **Personal deductions**: medical 15,000 (electronic, self); medical 3,000 **cash**; tuition 30,000 for *Luis*, primaria, a descendant with the 1-UMA attestation checked; donation 5,000 to an authorized donee. **Datos del año**: 2025 accumulable income 400,000. **Filed payments**: *Confirmar todos* on 612 |
| *Ana Consultores SC* (PM) | 601 from enero de 2026. The SAT check-digit warning on the RFC was shown and accepted; optional inputs skipped. **Income** 20/03 500,000.00 (persona moral, honorarios). **Filed payment** typed for March only: `45,000.00` |

Two items used a workaround, because the UI has no way to create them:

- **End of the trial (read-only):** the `trial` entitlement row was moved into the past in this run's
  own database. This is the same statement as the `makeReadOnly` test fixture.
- **Flag changes:** made with `cmd/flags`, against this database only.

Records stopped where the trial's limit of 5 incomes and 5 expenses (counted across taxpayers) left
them. The read-only GETs used to compare the screen with the server went through the API as the QA
account.

## 3. The walk: screen vs hand arithmetic vs server

| Check | Expected | On screen | Result | Evidence |
|---|---|---|---|---|
| Entry "Declaración anual" on Home with `demo_full` | Present | Present in the header | Pass | `05` |
| Year picker, year in progress | 2026, `year_in_progress` "…hasta septiembre de 2026" | Same; ISR computed with Jan–Sep only | Pass | `10` |
| PF, 612 only, before any input | Accumulable 240,000 − 30,000 = **210,000**. Tariff row 5: 16,069.64 + 34,264.33 × 17.92 % = **22,209.81**. Withheld 24,000; app payments 8,215.64 (February only, chained) | 210,000 · 22,209.81 · 8,215.64 · 24,000 · **Saldo a favor $10,005.83**, `incomplete`, "Falta: Constancias de sueldos", "Pagos: 9 estimados por la app", the "Solo se devuelve…" note and, at ≥ 10,000, "necesitas e.firma" | Pass | `11`, `12` |
| Falta CTA → constancias → back | The constancia joins: 510,000. Row 7: 67,981.92 + 85,646.02 × 23.52 % = **88,125.86**. Credits 8,215.64 + 24,000 + 45,000.50 | Section `Completo`, **Saldo a cargo $10,909.72** (refreshed on return) | Pass | `15`, `16` |
| Deduction caps | Tuition primaria capped at 12,900 ("Tope aplicado"); cash medical excluded with its line; registered 48,000, applied 27,900 | "Registrado $48,000.00 · Aplicado $27,900.00", "Tope aplicado: colegiaturas de Luis, de $30,000.00 a $12,900", "No deducible por pago en efectivo: $3,000.00"; the cash warning appeared in the form | Pass | `18`, `21` |
| Required errors on submit | Summary at the top, the link focuses the field | "Revisa 1 dato antes de guardar · Monto: Escribe este dato." The link moved focus to Monto | Pass | `17` |
| 7-field deduction form on its own screen, 1-UMA attestation for a dependant | Own screen; checkbox shown for non-self | Same | Pass | `19`, `20` |
| Delete + Undo | The toast stays (no auto-dismiss) and restores the row | Still there after 8 s; *Deshacer* → "Deducción restaurada", totals back | Pass | `22` |
| Donation without 2025 income | Excluded; section `incomplete`; Falta names the input | Applied unchanged; "Falta: Ingresos acumulables del año anterior"; **Saldo a cargo $4,347.64** = row 7 on 482,100 (81,563.78) − 77,216.14 | Pass | `24`, `25` |
| CPA strong banner | A dependant's UMA rule raises `uma_for_minimum_wage` | Amber banner "Requiere validación de un contador … UMA" above the legend and the figure | Pass | `24` |
| Falta → Datos del año → 400,000 | Donation applies (7 % cap 28,000); base 477,100; ISR 80,387.78; balance 3,171.64 | 32,900 · 477,100 · 80,387.78 · **Saldo a cargo $3,171.64**, `Completo` | Pass | `27` |
| `informative_income` (PF total income > 500,000) | 240,000 + 320,000 = 560,000 → note shown | "Además debes informar…(LISR 150)" | Pass | `25` |
| **Confirmar todos** (US-A9) | Months 1–9 stored as filed (10–12 not closed); `payments_estimated` gone; figure unchanged | DB: 9 rows, Feb 8,215.64, others 0; "Confirmado" on each; annual still **$3,171.64**, warning gone | Pass | `28`, `29`, `30` |
| Breakdown | Lines with legal basis | "Ingresos cobrados … 240,000 · Fundamento: LISR 109", "Deducciones autorizadas … menos 30,000" | Pass | `31` |
| **PM 601** | 500,000 × 30 % = **150,000**; PTU base 500,000, PTU 50,000; suggested CU 1.0000; deadline 31 mar 2027; CPA topics `cu_rounding`, `ptu_rate`; `inflation_adjustment_excluded` | Same; the CPA banner and "Sin validar por un contador" on the PTU box; "Falta: Pagos provisionales presentados" naming enero..septiembre (the monthly estimates are incomplete without a CU) | Pass | `32`–`34` |
| PM: data block | Only the provisional payments and Datos del año (no deductions or constancias) | Same | Pass | `36` |
| PM typed payment | March 45,000 → balance 105,000; March leaves the Falta list | DB: one row (3, 4,500,000); **Saldo a cargo $105,000**; Falta names every month but March | Pass | `35`, `43` |
| Disclaimers always | `annual_estimate`, `no_annual_iva`, `rmf_2027_pending` on every estimate; the legend next to the figure | Present on PF (before and after the inputs), PM, and as read-only | Pass | `07`, `25`, `34` |
| **Read-only** (trial ended) | Reads work. Add → 402 → paywall, what was typed is kept; the list then says read-only. Delete + Undo allowed (ADR-0009). Confirmar todos → paywall | Exactly that: "Tu cuenta es de solo consulta"; the form kept `1,000.00`; delete + *Deshacer* worked; Confirmar todos opened the paywall | Pass | `37`–`40` |
| **Flag OFF** `module_annual` | `GET /v1/flags` leaves it out; annual route `404`; no entry | Same (after an app restart) | Pass | `41` |
| **Prerequisite OFF** (`module_tax_engine` OFF, annual ON) | Annual implied OFF: absent, `404`, no entry (History hidden too) | Same | Pass | `42` |

The app learns "read-only" only from the first refused write: `readOnlyAccessProvider` is documented
as UX only. So the first *Agregar* opens the form and the paywall shows on save. This is by design;
the server's 402 is the boundary.

## 4. Defect found and fixed

**Accessibility: a section with a single missing input was announced as one button.**

When a section had exactly one "Falta" row, the `ListTile` of that row merged into the section's
semantics container. TalkBack then read the whole card (regimes, status chip, obligation and its
reason, deadline, notes, and the row's label) as **one button**, and activating it opened that input's
screen. With two or more missing rows, the conflicting tap actions happened to split them into
separate nodes, so the bug showed only in the common single-input case: every taxpayer in this walk.

- Found on the device: `uiautomator` showed the section as `Button "Regímenes 605, 612 / Incompleto / …"`
  with no node for "Constancias de sueldos". A semantics-tree probe in a widget test confirmed it.
- **Regression test first:** `annual_screen_test.dart` › *"a single missing input is its own button;
  the section is not one"*. It was red, with the link's label equal to the whole section.
- **Fix** (`annual_link_row.dart`): `Semantics(container: true)` around the row. It is then always its
  own node, whatever its siblings. This also covers the "Datos para tu anual" rows (the PM block has
  only two). Green; full mobile suite 1390 passing. Re-checked on the device: "Falta:" is text, and
  "Pagos provisionales presentados / enero…" is its own button (`43`).

## 5. Automated evidence

| Suite | Where | Result |
|---|---|---|
| `go test ./...` (backend unit) | local | Pass |
| `make test-integration` (unit + integration + contract on real PostgreSQL, incl. the annual golden gate and the new QA suite) | CI, PR #153 | Pass (the first push failed only on my test's expectations, see §6; a later run was abandoned when the runner could not download `setup-go`, and its re-run passed) |
| `flutter analyze`, `flutter test` (1390 passing, 39 skipped: mock-gated and golden-gated) | local + CI | Pass |
| Release build + artifact gate, codegen drift, govulncheck, gitleaks, GitGuardian | CI | Pass |

### Tests added on this branch

| Test | What it pins |
|---|---|
| `backend/cmd/api/annual_qa_integration_test.go` `TestAnnualQA_confirmAllKeepsTheFigure` | US-A9 end to end: estimate (credits estimated by the app, `payments_estimated`) → filed-payments GET → the app's confirm-all PUT → every estimated month becomes `filed`, the warning disappears, no figure moves |
| `… TestAnnualQA_personaMoral` | PM through the stack: `pm_601`, `inflation_adjustment_excluded`, PF-only inputs `422 person_type_not_allowed`, filed payments readable |
| `… TestAnnualQA_readOnlyAndArchivedOnEveryWrite` | Every one of the 15 operations. Read-only: the 6 reads 200, the deletes and the restore open (ADR-0009), the other 6 writes 402. Archived: every write 422 `taxpayer_archived` |
| `… TestAnnualQA_deductionBoundOnTheRealStore` | 500 live deductions a year on PostgreSQL: the 500th accepted, the 501st `limit_reached`, a deleted row frees its place, restoring into a full year refused |
| `mobile/test/features/annual/annual_goldens_ui_test.dart` (9 cases) | AN-01, 02, 04, 05, 06, 07: balance in words + amount to the centavo + deadline by person type; section lines each in the row of its own label. AN-08: cap 45,000 and ISR 31,807.91. AN-03 (a)/(b): an informative difference, never a balance. The test was shown able to fail (a pairing mutation turned it red) |
| `mobile/test/features/annual/annual_screen_test.dart` (1 case) | §4 regression |

## 6. Observations (not defects, for the owners)

- **The spec text vs ADR-0009 on read-only:** MODULE-annual US-A13 and §5 say read-only "writes return
  402", but ADR-0009 (and CONTRACT.md 0.6.1) keep delete and restore open, as the code does. My first
  version of the test followed the spec and failed in CI; the test now follows the ADR. Owner: doc
  update in ANN-26.
- **The app's prefilled estimate differs from the monthly estimate screen.** The annual chains the
  provisional payments it credits (`monthlyISR` ignores a missing `prior_provisional_payments`). The
  filed-payments screen therefore prefills May with **$0.00 "Estimado por la app"**, while the May
  estimate screen, which assumes no prior payments, shows ISR **$4,598.27 (incompleto)**. The annual is
  the more accurate figure, but the two screens disagree with no explanation. Suggest a hint on the
  filed-payments rows, or noting it in the user guide (ANN-26).
- **Incomes and expenses (M3/M4, not AN):** in the record forms, the regime radio defaults to the first
  regime held. For a 605 + 612 taxpayer that is the informative 605, and its "Este régimen es
  informativo" note sits above the fold, so a user can save a record that does not count. Worth a
  look by the M3/M4 owners.
- The onboarding month picker sits below a long regime list and defaults to the current month. I
  missed it on the first taxpayer (start = October, so the annual window was empty). A start month
  later than the first record is refused anyway, but the default is easy to overlook.
- `.gstack/` is git-ignored. This report is added with `git add -f` because the task asks for it in
  the PR; the screenshots stay local.

## 7. Not covered

- iOS (OB6).
- A full TalkBack pass beyond the semantics tree. Checked: the balance card's single spoken summary
  (on screen), the order banner → legend → headline, and §4.
- Archived taxpayer in the UI: covered through the stack by the new integration test only.
- RESICO PF / PM, 625 and 603 on the device. Their figures are pinned by the goldens and the new
  widget test.

## 8. Teardown

Stopped only what this run started:

- the API on `:3093` (its own background task);
- the PostgreSQL cluster on `:5493` (`pg_ctl stop`).

Both ports are free. The cluster's data directory is in the session scratchpad. The dev app built
for `:3093` is still installed on the emulator; reinstall it for another API port. Flags were left at
`demo_full` before the stop. Nothing else was stopped or changed.
