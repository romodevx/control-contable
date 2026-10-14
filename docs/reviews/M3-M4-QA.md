# QA report: M3 Incomes and M4 Expenses (M3-10, M3-11, M4-11, M4-12), Android emulator, 2026-10-14

| | |
|---|---|
| Scope | Flag registration tests (M3-10, M4-11) and the full M3 + M4 walk on a device (M3-11, M4-12): incomes per regime and IVA rate, PM withholdings, duplicate CFDI, save-and-create-another, edit, delete + Undo, filters, trial caps, expense categories and hints, cash rules, non-deductible expenses, cross-user 404, read-only, flags OFF |
| Branch | `feature/incomes-expenses-qa-fixes` (from `origin/dev` @ `82584eb`) |
| Device | Android emulator `emulator-5554` (`Medium_Phone_API_36.1`), Android 16, arm64-v8a |
| App | dev flavor, **profile** build for arm64, `--dart-define=API_BASE_URL=http://10.0.2.2:3011` (QA logging proxy to this run's own API on `:3010`) |
| API | `make -C backend api API_PORT=3010` (APP_ENV=dev, dev email adapter), local PostgreSQL 18 (docker compose, `127.0.0.1:5434`). Ports 3010/3011 were free; no other agent's process was touched |
| iOS | **Not run** (OB6, as in M1-15 and M2-12) |
| Evidence | Screenshots `NN-*.png` in `.gstack/qa-reports/screenshots/` on the QA machine (git-ignored) |
| Verdict | **Pass after fixes.** One server defect (flag prerequisite) and four app defects found, each fixed with a regression test first; 9 items deferred |

## 0. M3-10 / M4-11: flags `module_incomes` and `module_expenses`

The baseline migration (P0-10) already seeds both flags **OFF** and puts them in `demo_core`,
`demo_monetization` and `demo_full` (asserted by `migrations/baseline_integration_test.go`).
Tests added, mirroring PR #48 (M1-14) and PR #68 (M2-11):

| Test | What it pins |
|---|---|
| `backend/adapter/http/gen/records_module_test.go` | The generated flag map gives each module exactly its 7 operations (by operationId and by ServeMux pattern, written by hand from the specs) and nothing else |
| `backend/cmd/api/records_preset_smoke_integration_test.go` | Every seeded preset plus four fixtures (no record modules, incomes only, expenses only, records without taxpayers) applied as `cmd/flags` applies them; all 14 operations probed with a real user, taxpayer, income and expense. A module answers only when the preset holds it **and** `module_taxpayers`; otherwise the hidden-module 404 and absent from `GET /v1/flags`. A guard fails if a new operation is not probed |
| `backend/internal/platform/flags/prerequisites_test.go` | `module_incomes` and `module_expenses` declare `module_taxpayers` |
| `mobile/test/features/{incomes,expenses}/*_flag_test.dart` | The route tree of each slice is exactly its 3 locations; with the flag OFF each is not found and nothing loads; ON, each opens |

**Defect F1 (fixed, server):** MODULE-incomes §6 and MODULE-expenses §6 say both modules are implied
OFF when `module_taxpayers` is OFF, but `flags.Prerequisites` only declared the tax engine. With
incomes and expenses ON and taxpayers OFF, all 14 record operations answered (the smoke's DELETE
removed an income) and `GET /v1/flags` listed both modules. Fixed by declaring the dependency in
the one graph the gate, `GET /v1/flags` and `cmd/flags` read. Verified live too: with taxpayers OFF,
`cmd/flags list` shows both as `hidden` and all 14 operations answer `404 not_found`.

## 1. What was wiped

- `make -C backend init-env` created a fresh `infra/local/.env` in this worktree (random passwords and keys).
- `assert-local.sh` confirmed `docker=unix:///…/docker.sock db_host=127.0.0.1`. `make -C backend db-reset`
  destroyed the `control-contable` data volume (left by an earlier run; no API was connected to it)
  and bootstrapped the roles again. `make -C backend migrate-up`: every migration applied.
- After the wipe: 0 users, 9 flags all OFF. `module_auth`, `module_taxpayers`, `module_incomes`,
  `module_expenses` turned ON with the flag CLI (local only).
- The previously installed app was uninstalled (fresh app data).

## 2. What was created through the UI

| Record | How |
|---|---|
| User *Ana QA* `ana.m3@example.com` | Sign-up form + both consents → code from the API log → verify |
| *Ana Honorarios* (PF, `PEXJ900215HK8`), 606 *Gastos reales* + 612, border IVA off on both | Onboarding opened by itself |
| 5 incomes of Ana (see flows 1-6) | Income form, three of them with *Guardar y crear otro* |
| 5 expenses of Ana (see flows 11-16) | Expense form, two of them with *Guardar y crear otro* |
| User *Bea QA* `bea.m3@example.com` + *Bea Consultora* (PF, `AAAA010101AA0`, 612) + 1 income | Sign-up, verify, onboarding, income form |
| User *Carla QA* `carla.m3@example.com` + *Carla Diseno* (PF, `GOMA800101AB6`, 612) + 2 incomes + 1 expense | Created after the fixes to re-verify them on the device |

Direct API calls were for evidence only: the cross-user probes, the flag-OFF probes, one PATCH of
an expense description to read the warnings the server returns, and a DB update to simulate an
expired trial (LOCAL only, flow 19). No seed scripts and no SQL inserts.

## 3. Flow results

### Incomes (M3)

| # | Flow | Result | Evidence |
|---|---|---|---|
| 1 | 612, persona moral, 10,000.00 at 16 %: live summary "IVA 1,600.00 · Retenciones −2,066.67 · Neto 9,533.33"; *Retenciones* opens prefilled 1,000.00 ISR (10 %) + 1,066.67 IVA (2/3); folio entered → `201` | Pass | `09`, `10` |
| 2 | *Guardar y crear otro*: the form stays, keeps date, regime and payer, clears the rest, focuses *Descripción*, toast "Ingreso guardado"; withholdings recomputed for 5,000.00 (500.00 + 533.33) | Pass | `11` |
| 3 | Same folio again → `201` with the non-blocking warning "Ya hay otro ingreso con este folio fiscal. Si es un pago parcial del mismo CFDI, está bien." (toast stays until closed) | Pass | `13` |
| 4 | 606 exempt (casa habitación, persona física), 8,000.00 → `201`; 612 at 0 %, `3,000.50` typed with separators → `201`; 606 at 16 % from a persona moral: 1,200.00 ISR + 1,280.00 IVA prefilled → `201` | Pass | `15`, `16` |
| 5 | Month totals: 5 incomes, 38,000.50 before IVA, IVA 4,320.00, withholdings −5,580.00, all matching the inputs | Pass | `16` |
| 6 | Invalid amount: a mistyped amount shows "Escribe un monto, por ejemplo 1,234.50." and nothing is sent | Pass | — |
| 7 | Edit: amount 3,000.50 → 3,500.00 → `PATCH 200`, totals updated, toast "Ingreso actualizado" | Pass | `17` |
| 8 | Filter by regime *Arrendamiento* → `?regime=606`, 2 incomes and their own totals; previous month → empty September | Pass | `18`, `19` |
| 9 | Delete → `204`, toast "Ingreso eliminado" with *Deshacer* still there after 12 s; *Deshacer* → `restore 200`, totals back | Pass | `20`, `21` |
| 10 | Trial cap: 6th income → `402 entitlement_required`, sheet "Llegaste al límite de ingresos de prueba … La prueba incluye 5 ingresos…", the form keeps what was typed; discard asks first | Pass | `22`, `08` |

### Expenses (M4)

| # | Flow | Result | Evidence |
|---|---|---|---|
| 11 | Expenses still allowed with incomes at 5/5: the expense form opens and saves | Pass | `23` |
| 12 | Category picker (searchable) with hints per regime: *Software* in 606 "Normalmente no es deducible… LISR art. 115", switching the regime to 612 reloads the catalog: "Normalmente deducible… LISR art. 103"; legend "Esto es un aproximado…" always shown | Pass | `24`-`26` |
| 13 | Cash $2,500.00 + IVA (*Papelería*, 612): on choosing *Efectivo* the form shows both notes ("Los pagos en efectivo de más de $2,000.00 no son deducibles." and "El IVA de este pago en efectivo no es acreditable."), *¿Es deducible?* defaults off; saved → `201`, `deductible=false`, toast with `cash_over_threshold`, list badge "No deducible" | Pass | `27`, `28` |
| 14 | Edit the same expense and override *¿Es deducible?* to on ("Lo elegiste tú", *Usar el valor sugerido*) → `PATCH 200` with both warnings in the toast (`cash_over_threshold` + `iva_not_creditable_cash_payment`, spec §2.5) | Pass | `29`, `30` |
| 15 | *Gastos personales* (search "personal"), exempt, transfer → hint "Normalmente no es deducible… Se aplican en tu declaración anual" → saved non-deductible | Pass | `31` |
| 16 | Without CFDI (*Luz, agua…*) → note "Sin factura, el gasto no es deducible." and `no_cfdi` warning; fuel paid in cash shows "El combustible pagado en efectivo no es deducible." | Pass (see D-2) | `33` |
| 17 | Filter by 606 → 3 expenses, totals of the filtered set | Pass | `38` |
| 18 | Delete at 5/5 → `204`; *Deshacer* after 10 s → `restore 200` (no slot used); 6th expense → `402`, "Llegaste al límite de gastos de prueba" | Pass | `35`-`37` |

### Cross-user, read-only, flags

| # | Flow | Result | Evidence |
|---|---|---|---|
| 19 | Read-only (LOCAL simulation: Bea's trial entitlement moved to the past by SQL as the bootstrap superuser): create income → `402` "Tu cuenta es de solo consulta"; edit → `402`; create expense → `402`; delete → `204`, *Deshacer* → `restore 200` | Pass | `40`-`43` |
| 20 | Cross-user: Bea's list is empty. With Bea's token, the 12 id-bearing operations on Ana's taxpayer, income and expense → `404 not_found`; a **valid** create body on Ana's taxpayer → `404` (both modules); Bea's own taxpayer with Ana's income/expense ids → `404`. Empty-body creates answer `400` before the ownership check (no existence leak: an unknown taxpayer answers the same) | Pass | probe |
| 21 | Flags OFF (incomes and expenses): all 14 operations with a real bearer and real ids → `404 not_found`; nothing deleted (6 incomes and 5 expenses live before and after); app relaunch: home hides *Ingresos* and *Gastos* | Pass | `44`, probe |
| 22 | `module_taxpayers` OFF with incomes and expenses ON (after the F1 fix): `GET /v1/flags` = `["module_auth"]`, all 14 operations `404`, `cmd/flags list` shows both `hidden` | Pass | probe |

### Logs: errors, PII, secrets

- API log: 39 lines, 0 WARN/ERROR. Audit events `income.created/updated/deleted/restored` and
  `expense.created/updated/deleted/restored`, each with only `user_id`, `taxpayer_id`, the record id
  and `request_id` (S-6). The dev email adapter redacts the address.
- `grep -iE 'PEXJ|AAAA01|GOMA|example.com|tortuga|ajolote|quetzal|eyJ|6F3B2C1D|3A9C7E21|logotipo|Asesoria|Gasolina'`
  over the API log: 0 hits (no RFC, email, password, token, CFDI UUID or description).
- Proxy log: 201 requests, 0 `5xx`. 8 `429 rate_limited` came from this run's own burst of probe
  calls (see E3), none from the app.
- logcat: no Flutter errors or crashes.

## 4. Issues

### Fixed (regression test first, then the fix; one commit each)

| ID | Severity | Defect | Test | Fix |
|---|---|---|---|---|
| F1 | High (server boundary) | Incomes and expenses reachable with `module_taxpayers` OFF (spec §6 of both modules) | `08d8db5` | `7e57970` |
| D1 | Medium (a11y) | The expense category picker and the taxpayer switcher wrap their button in `Semantics(button: true, excludeSemantics: true)`, which drops the tap action: TalkBack/VoiceOver announce a button they cannot activate (uiautomator showed both as not clickable). Once tappable, the switcher also needed the 48 dp Android target its own guideline test checks | `e19af74` | `41d2d20` |
| D2 | Medium (UX) | After *Guardar y crear otro*, a warning toast stays on the form until closed and covers *Cancelar* and *Guardar y crear otro* (`33`). It now floats above the footer | `bf8cf30` | `a5d44e1` |
| D3 | Low (a11y, WCAG 2.2 2.4.11) | Expense *Guardar y crear otro* focuses the category picker but leaves the form scrolled at the bottom, so the focused control is off screen (`32`). AppForm now reveals the focused field | `548a4aa` | `981a653` |
| D4 | Medium (privacy/UX) | What the forms remember "in this session" (last payer, last payment method, category usage, both Undo stacks) survived sign-out: Bea's first income started on Ana's last payer (persona moral, withholdings prefilled). Now forgotten when the session ends | `f7f366d` | `e4380a7` |

Re-verified on the emulator with a rebuilt app: the switcher and category are clickable for
accessibility services (`45`, `49`), the duplicate-CFDI toast floats above the footer (`47`), the
category is shown after *Guardar y crear otro* (`49`), and a new account starts on *Persona
física* and *Tarjeta de débito* (`46`).

### Deferred

| ID | Severity | Item | Owner / next step |
|---|---|---|---|
| D-1 | Low (UX) | The form shows the IVA note "El IVA de este pago en efectivo no es acreditable." before the user overrides deductibility, while the server only returns `iva_not_creditable_cash_payment` on an override (§2.5). The note is true, but the two disagree | pm + frontend-sr: show it only with the override, or have the spec return it always |
| D-2 | Low (UX) | The fuel-in-cash note shows for 606, whose rule is `not_deductible` / `cash_payment_rule = not_applicable`; the server returns no cash warning there. The client keys the note on the category code, not on the rule | frontend-sr: read the rule's cash behavior from the categories response |
| D-3 | Low (copy) | "Gasto actualizado, con un aviso:" is singular even with two warnings | Copy |
| D-4 | Info | US-E story "300.00 of combustible in cash → warning" was not re-run on the device for 612 (Ana was at 5/5); the contract tests cover it | — |
| D-5 | Low (UX) | `8 % (región fronteriza)` is offered while the taxpayer's regimes have the border option off (spec §2.5 lets the user declare it per record) | pm: hide or warn when the regime option is off |
| D-6 | Low (UX) | The regime filter resets when changing month; the empty month says "Registra tu primer ingreso…" even when other months have records | pm + frontend-sr |
| D-7 | Low (UX) | The 612 filter chip label is clipped ("…Empresar") at phone width | frontend-sr: short regime names in chips |
| D-8 | Low (UX) | After *Guardar y crear otro*, *Cancelar* on an untouched form asks "¿Descartar este ingreso?": the kept context fields count as unsaved changes (`48`) | frontend-sr: compare against the prepared-next baseline |
| D-9 | Info | Onboarding starts every regime in the current month, so records dated before it are refused; a user starting mid-year cannot record earlier months without editing the regime | pm (M2 onboarding) |

## 5. Environment problems (not app bugs)

| ID | Problem | Workaround |
|---|---|---|
| E1 | The API has no access log (same as M1-15 E3, M2-12 E2) | QA-only logging proxy on `:3011` (method, path, status, error code; no headers, tokens or bodies) |
| E2 | `adb input text` sometimes drops the first character right after a field is focused (one amount arrived as `,000.50`) | Pause 1 s after focusing; the app rejected the malformed amount correctly |
| E3 | The proxy makes every request come from 127.0.0.1, so the probe bursts hit the per-IP limiter (`429`) | Waited 20 s and re-ran the probes |
| E4 | The local database container was left by another worktree's run | Wiped with `db-reset` from this worktree's freshly generated `.env` |

## 6. Checks run on the branch

- `make -C backend check`: fmt, golangci-lint (0 issues), unit tests pass.
- `go test -tags=integration ./...` (backend): all packages pass, presets smokes included.
- `make mobile-check`: format, analyze (0 issues), 973 tests pass. `make mobile-cover`: 95.79 % (min 80 %).

## 7. Local state left behind

- **All 9 flags OFF**. This run's API (`:3010`) and proxy (`:3011`) stopped.
- Users *Ana* (1 taxpayer, 5 incomes and 5 expenses, both trial caps used), *Bea* (read-only by the
  simulation, 1 income) and *Carla* (2 incomes, 1 expense) remain in the local database.
