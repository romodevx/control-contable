# QA report: iOS walk (OB6), 2026-10-15

| | |
|---|---|
| Scope | The iOS walk that every module QA had left blocked by OB6 (M1-15 … M9-12, ANN-23, M8-07 §5, PLAN "QA §7"): `demo_full` on an iOS simulator, data entered through the screens, figures checked against the goldens, VoiceOver labels, Dynamic Type, dark mode, safe areas, the keyboard and sticky footers, flags OFF; plus the obfuscated iOS release build |
| Branch | `feature/ios-qa-walk` (from `origin/dev` @ `eea54b9`) |
| Device | iPhone 17 simulator, iOS 27.0, Xcode 27.0 (27A266a), shown in Xcode's DeviceHub; the dev flavor (`mx.romodevx.controlContable.dev`) built with `API_BASE_URL=http://localhost:3103` |
| Blocked (owner prerequisites, not faked) | **Push on iOS** (no Firebase project, no APNs key: OB9 + Apple account): permission prompt, device registration from the app, delivery, tap-to-open. **Signed device build / TestFlight** (no Apple team: REL-04). **A real VoiceOver pass by ear** (see §4) |
| Verdict | **Pass after four fixes.** Every golden checked matches to the centavo on iOS. The iOS flavor schemes were missing (F1, fixed); three app defects found on iOS and fixed regression-test-first (F2–F4). None is iOS-only in code; all three were missed by the Android walks |

## 1. Environment

Same method as [the WH-13 report](qa-report-withholdings-2026-10-14.md) §1, with this run's own stack:

- Homebrew PostgreSQL 18.6, a fresh cluster in the session scratchpad on `127.0.0.1:5503` (TCP only), database `cc_iosqa`; roles from `infra/local/postgres/bootstrap/roles.sql`; every password and key generated at run time into a 0600 file in the scratchpad (from `infra/local/scripts/init-env.sh`), never in the repo.
- `cmd/migrate up` as `cc_owner`, `cmd/flags apply-preset demo_full` (10 of 10 ON), `cmd/api` as `cc_app` on `:3103`. The simulator reaches the host as `localhost`; Dart sockets ignore ATS, so plain HTTP to the dev API works (README).
- Plugins resolve through Swift Package Manager (Flutter 3.47): no CocoaPods step, no `Pods/` directory.
- Driving the simulator: [AXe](https://github.com/cameroncooke/AXe) 1.8.0 (`brew install cameroncooke/axe/axe`) for taps, typing, swipes, screenshots and the accessibility tree (`axe describe-ui`, the labels VoiceOver reads). Two tool quirks to know for the next run: Flutter ignores AXe's quick taps after a while (a touch with a 0.12 s hold always works), and the tree keeps stale frames for nodes scrolled into view (re-read after a pause, or use a screenshot).

**Wipe**: the database was created empty for this run; nothing had to be wiped. No other database, process or emulator was touched.

## 2. Created through the app (no SQL, no seed)

One account, `qa-ios@example.com`, signed up in *Crear cuenta* (code from the dev email log). The trial allows 5 incomes; the account got the operator's `comp` entitlement, the only SQL of the run (as in WH-13 and M5-29).

| Taxpayer (onboarding in the app) | Records entered in the forms |
|---|---|
| **Ana RESICO**, PF 626 from enero 2026 (start month picked first on step 3) | **US-W1, the owner's example**: 10/07 100,000.00 + 16 % to a company, professional services · **WH-05b**: 10/08 50,000.00 of goods (*Guardar y crear otro*) · a late 15/09 income of 10,000.00 (History marker) |
| **Ana Despacho 2026**, PF 612 + 605 from enero 2026 | Incomes to a company, fees, regime 612: 15/02 100,000.00 · 15/05 80,000.00 · 15/08 60,000.00 (save-and-new) · expense 10/03 rent 30,000.00, transfer, with CFDI · **annual inputs**: constancia *Empresa A* 300,000.00 / 20,000.00 / 45,000.50 · medical 15,000 (card) and 3,000 (cash) · tuition 30,000 for *Luis*, primaria, descendant with the 1-UMA attestation · donation 5,000 to an authorized donee · *Datos del año* 2025 income 400,000 · *Confirmar todos* on 612 |
| **Empresa QA**, PM 601 from enero 2026 | Income 20/03 500,000.00 (invoice date, accrual) · **US-W6**: expense 15/06 fees 40,000.00 + IVA to a PF 612 supplier · filed payment March 45,000.00 typed |

Also exercised: the RFC check-digit warning, the error summary on an empty save (link + `aria`-style hint on the field), *Cancelar* → "¿Descartar este gasto?", the cash-deduction warning, delete-row buttons, reminder settings (§3.4).

## 3. Results (screenshots `ios-*.png` in `.gstack/qa-reports/screenshots/`, local only)

### 3.1 Figures against the goldens

| Check | Golden / hand arithmetic | iOS | Result |
|---|---|---|---|
| US-W1 card | ISR 1,250.00 · IVA 10,666.67 · Recibes 104,083.33 | identical (`ios-10`) | Pass |
| WH-01-EST, July | ISR 750.00, IVA 5,333.33, total 6,083.33 | identical on Home's chart and its spoken label | Pass |
| WH-05b | ISR 625.00, IVA 0, Recibes 57,375.00; August ISR 0 | identical (`ios-12`); August ISR 0, IVA 8,000 (goods) | Pass |
| 612 fees to a company | 10 % ISR withheld: 10,000 / 8,000 / 6,000 | identical | Pass |
| Annual PF, before inputs (ANN-23 walk) | 210,000 · 22,209.81 · payments 8,215.64 · withheld 24,000 · **Saldo a favor 10,005.83** | identical (`ios-19`, `ios-20`) | Pass |
| + constancia | **Saldo a cargo 10,909.72** | identical (`ios-22`) | Pass |
| + deductions | Registered 48,000, applied 27,900, tuition capped 30,000 → 12,900, cash excluded, donation excluded without 2025 income; **4,347.64**; CPA banner (UMA) | identical (`ios-28`) | Pass |
| + Datos del año 400,000 | **3,171.64**, `Completo` | identical | Pass |
| US-A9 *Confirmar todos* | 9 rows (Feb 8,215.64), figure unchanged, `payments_estimated` gone | identical; DB 9 rows, 821,564 centavos | Pass |
| US-W6 card | Debes retener ISR 4,000.00 · IVA 4,266.67, Pagas 38,133.33, Entregas 8,266.67 **a más tardar el 17 de julio** | amounts identical; **the deadline read "17 de noviembre"** when the date was changed after the amount | **F2, fixed** |
| WH-11-JUN | remit 8,266.67; IVA a favor 2,133.33 | identical (`ios-35`) | Pass |
| WH-11-JUL | credit 4,266.67, labelled "IVA acreditable" (WH-13 F1) | "IVA acreditable (incompleto) −$4,266.67" (`ios-37`) | Pass |
| PM annual 601 | (500,000 − 40,000) × 30 % = **138,000**; March 45,000 typed → **93,000** | identical (`ios-38`) | Pass |
| History marker (M8) | a late record marks the month "Cambió"; opening it clears it | "Cambió desde que lo viste…", then "Visto el 15 oct." (`ios-42`) | Pass |

### 3.2 VoiceOver labels (accessibility tree)

Read with `axe describe-ui`, which reports what VoiceOver reads (labels, values, hints, frames).

| Check | Result |
|---|---|
| Buttons, fields, checkboxes named by their visible text; icon actions named ("Declaración anual", "Historial", "Recordatorios", "Mi cuenta", "Contribuyente activo: … Cambiar") | Pass |
| Radio options: unselected ones carry the hint "Sin seleccionar" | Pass |
| Chart bars: "julio de 2026: IVA 5,333 pesos con 33 centavos, ISR 750 pesos, total 6,083 pesos con 33 centavos" | Pass |
| Field errors exposed as the field's hint ("Escribe este dato.") and in the summary link | Pass |
| **Pickers read their label twice**: "Ejercicio, Ejercicio, 2026" (annual), "Ejercicio fiscal, Ejercicio fiscal, 2026" (History), "Mes, Mes, septiembre de 2026" (estimate), and the start-month / end-of-regime questions as a text and again as the picker | **F3, fixed** |
| Masked RFC read as "RFC ROAA, oculto, AB1" | Pass |

### 3.3 Dynamic Type, dark mode, safe areas, keyboard

| Check | Result |
|---|---|
| Dark mode (system appearance) | Pass: tokens switch, contrast holds on Home, forms, dialogs, Reminders (`ios-45`, `ios-47`, `ios-51`) |
| ~200 % text (`accessibility-large`, body ≈ 194 %) | **Home's app bar overflowed**: "Cerrar sesión" clipped behind the overflow stripe and the taxpayer switcher pushed off screen (`ios-46`) → **F4, fixed** (`ios-54`). Forms, annual, History and Reminders wrap without loss |
| 235 % (`accessibility-extra-large`) | Beyond WCAG's 200 %: "(incompleto)" breaks mid-word in the estimate card's narrow label column (O3) |
| Safe areas (Dynamic Island, home indicator) | Pass in portrait on every screen visited. Landscape not checked: no rotation command in this setup |
| WCAG 2.4.11 sticky footer | Pass: a focused field under the form footer scrolls clear (March payment field 725 → 562 pt) |
| On-screen keyboard covering fields | **Not verified**: DeviceHub keeps the hardware keyboard connected and this session cannot reach its menus (no Accessibility / screen-recording permission), so the software keyboard never showed. Owner step: *I/O → Keyboard → Toggle Software Keyboard*, then tap the last field of a long form |

### 3.4 Reminders, flags, push

| Check | Result |
|---|---|
| Reminders screen without Firebase: "Notificaciones no disponibles en esta versión de la app.", the switch does nothing, no OS prompt | Pass (as designed, OB9) |
| Settings enabled once with `PUT /v1/reminder-settings` (the M9-12 workaround); on screen: due-date toggle off, amounts on ⇒ "Guardado." and the server holds `on_due_date: false`, `show_amounts: true` | Pass (`ios-48`) |
| "Próximos vencimientos": lunes 19 de octubre (weekend move), 17 nov, 17 dic | Pass |
| Send-time picker, text-entry mode | Not verified: AXe's typing does not reach the stock time-picker's two fields (the dial is exposed as sliders). Passed on Android in M9-12 with the same widget |
| `module_annual` and `module_history` OFF | `GET /v1/flags` leaves them out, the annual route answers `404`, and after relaunch the app bar shows neither entry (`ios-53`); `demo_full` re-applied after | Pass |
| Push registration / delivery on iOS | **Blocked** (OB9: no Firebase project; APNs key and Push capability need the Apple account) |

### 3.5 Release build (obfuscated, no signing)

`flutter build ios --release --no-codesign --flavor prod -t lib/main_prod.dart --obfuscate --split-debug-info=build/symbols/prod --dart-define=API_BASE_URL=https://api.example.invalid` → `✓ Built build/ios/iphoneos/Runner.app (21.2MB)`.

| Check | Result |
|---|---|
| `*.map`, `*.symbols`, `*.dSYM` inside `Runner.app` | 0 |
| Files with `sourceMappingURL` | 0 |
| Readable Dart class names in `App.framework/App` (`AnnualYearPicker`, `ExpenseFormModel`, …) | 0; the same grep finds 214 in `build/symbols/prod/app.ios-arm64.symbols`, which stays on the build machine, so the check can fail |
| `__debug` sections in `App` | 0 |
| dSYMs | next to the app in `build/ios/Release-prod-iphoneos/`, not inside it |
| Bundle id / name | `mx.romodevx.controlContable` / "Control Contable" |

The CI artifact gate (`tools/mobile/check-release-artifact.sh`) covers APK/AAB only; an IPA gate belongs with REL-04 (O6).

## 4. Fixes on this branch (each regression test first, red before)

| Commit | What |
|---|---|
| `feat(mobile): iOS flavor schemes dev, staging and prod (OB6)` | **F1.** `flutter run/build ios --flavor dev` failed: "The Xcode project does not define custom schemes", so `make mobile-run` and `make mobile-build-ipa` could not target iOS. Shared schemes `dev`/`staging`/`prod` with `Debug-`/`Profile-`/`Release-<flavor>` configurations, mirroring Android: own bundle id (`.dev`, `.staging`) and home-screen name (`APP_DISPLAY_NAME`). `Package.resolved` pins the SPM Firebase dependencies. Test: `test/platform/ios_flavors_test.dart` (26 red before). Mobile README updated |
| `fix(withholdings): a date changed after the amount refreshes the preview` | **F2.** Typing the amount first and the date after left the withholding card on the old date's facts: US-W6 read "Entregas al SAT … a más tardar el 17 de noviembre" for a 15/06 payment (server: 17 de julio). Text fields notify the form only when their error changes; the models listened to the amounts, not the dates. Same gap in the income form. What was saved was always right (the server derives on save; the toast said "17 de julio"). Tests: `expense_supplier_test.dart`, `income_withholding_test.dart` "changing the … date after the amount". Re-checked on the simulator (`ios-44`) |
| `fix(annual): announce the year picker's label once` + `fix(a11y): every labeled picker announces its label once` | **F3.** Five pickers had a visible caption and the same text as the dropdown's `Semantics(label:)`; VoiceOver (and TalkBack) read it twice. Each caption is excluded from semantics. Tests: `labeled_pickers_semantics_test.dart` (annual, History, estimate) and "announced once" in `onboarding_regimes_test.dart` and `settings_regimes_test.dart`, over a shared `timesSpoken()` helper. Re-checked on the simulator ("Ejercicio fiscal, 2026", "Mes, septiembre de 2026") |
| `fix(a11y): Home's app bar fits at 200 % text` | **F4.** Sign-out was a text button next to four icons; at 200 % it overflowed the bar (WCAG 1.4.4). It is now an icon button whose tooltip and VoiceOver name are "Cerrar sesión", like "Mi cuenta"; behaviour unchanged. Test: `home_app_bar_test.dart` (iPhone and small phone, 100 % and 200 %) |

One suspected defect was **not** one: the PM fees expense once showed E1 preselected as "Una empresa". A clean replay showed no preselection (spec §2.2); the selection came from a stray automated tap.

## 5. Observations (not fixed; for the owners)

- **O1, informational regime as the default regime.** For a 612 + 605 taxpayer the income and expense forms start on **605** (the first regime by code), whose records "no cambia tu estimación"; the user must scroll to the bottom and switch to 612 every time. The spec only defines the single-regime case (MODULE-incomes §221, MODULE-expenses §196). Suggest: default to the first fully calculated regime, or the last one used. Already tracked as PLAN follow-up "QA O3" (annual walk); seen again here.
- **O2, the active taxpayer's name barely fits on Home.** With `demo_full` the bar has the switcher plus five icons; on a 402-pt iPhone the name shows as "Empr…" at 200 % and was squeezed to its chevron at 100 % before F4. Five actions exceed the usual three plus overflow; worth a design pass (e.g. History/annual/reminders on Home's body or an overflow menu).
- **O3, mid-word wrap beyond 200 %.** At 235 % the estimate card's label column is narrow enough that "(incompleto)" breaks inside the word ("(inc / ompleto)"). WCAG asks for 200 %, where it wraps between words.
- **O4, `LoadFailed` at 200 % on a 360-pt phone.** In a widget test, the shared error state's retry button left its message 0 pt wide (`RenderFlex` overflow) when the dashboard failed to load. Not seen on the device (the load never failed); worth a test of `LoadFailed` at 200 %.
- **O5, the placeholder Home overflows at 200 % on a small phone** when `module_dashboard` is OFF (no preset does that today).
- **O6, no IPA artifact gate in CI.** The iOS release was checked by hand (§3.5); add the checks to `check-release-artifact.sh` or a macOS job with REL-04.
- **O7, money-field prefix.** The "$ " prefix is a separate accessibility node next to each money field in this tree; confirm by ear whether VoiceOver stops on it.
- Owner steps that need a person or an account: a real VoiceOver pass by ear (gestures, rotor, focus return after dialogs), the software keyboard over the last field of long forms (§3.3), landscape, and push once OB9/APNs exist.

## 6. Suites

| Suite | Where | Result |
|---|---|---|
| `flutter analyze`, `flutter test` | local | Pass: 0 issues; 1540 passing, 39 skipped (mock- and golden-gated) |
| `dart format --set-exit-if-changed` | local | Pass |
| Backend | not touched by this branch | CI |

## 7. Teardown

This run's API (`:3103`) and its PostgreSQL cluster (`:5503`, data directory in the session scratchpad) were stopped by this run, and the iPhone 17 simulator it booted was shut down. The dev app stays installed on that simulator with the QA account (local data only). The Android emulator was not touched. AXe stays installed via Homebrew for the next iOS run.
