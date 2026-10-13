# P0-31 · Tech-lead gate: mobile foundation

| Field | Value |
|---|---|
| Scope | `mobile/` tree and `.github/workflows/mobile.yml` on `dev` @ `6a669a0`, reviewed as one system (not per PR) |
| Covers | P0-22 scaffold (#15), P0-23 generated client + P0-24 FlagsProvider (#20), P0-25 mobile CI (#18) |
| Reviewer | tech-lead |
| Date | 2026-10-13 |
| Verdict | **APPROVED WITH FOLLOW-UPS.** No blocker remains on `dev`. M1 mobile work (M1-11..M1-13) may start once the items marked **M1 prerequisite** are planned into it |

## Evidence (re-run on this branch)

| Check | Command | Result |
|---|---|---|
| Format | `dart format --set-exit-if-changed .` | 0 changed (324 files) |
| Analyze | `flutter analyze` (very_good_analysis + strict-casts/inference/raw-types + riverpod_lint) | No issues |
| Tests + coverage | `make mobile-cover` | All pass (1 skip: `--tags mock`, needs Prism); 97.47 % of non-generated `lib/` (gate 80 %) |
| Mock integration | `make mobile-test-mock` | Not run here (needs `make mock`); CI does not run it either, see F-12 |
| Release build + artifact gate | CI job `release-android` | Not re-run locally (CI ran it on #18); config read and judged below |

Probe (temporary test, not committed): a `go()` into an enabled module while `GET /v1/flags`
is still pending lands on not-found and stays there after the flags arrive. Confirms F-01.

## What is right (keep it)

- **Architecture.** `core/` (config, network, flags, router, theme) plus `features/<slice>/`
  with a `FeatureModule` per slice that the router composes. A slice is added in one place
  (`featureModulesProvider`) and stays invisible while its flag is off. Ready for M1+.
- **Riverpod 3 idioms.** `AsyncNotifier` for flags, `select` in `moduleEnabledProvider` so
  widgets rebuild only for their module, `ref.mounted` after await, `onDispose` for every
  owned resource (`Dio`, `GoRouter`, `ValueNotifier`, `AppLifecycleListener`). Config is a
  scoped override that throws when missing: no hidden globals.
- **go_router.** Built once per feature set; flag changes go through `refreshListenable`, so a
  refresh that touches another module does not reset navigation (tested). Ownership is by
  route identity at any depth, so nested and shell routes are covered.
- **Flags.** One enum (`AppModule`), one check (`ref.useFlag` / `FlagGate`), fail-closed on
  first load, last-known answer kept on a failed resume refresh, unknown keys dropped.
  Comments state plainly that the server 404 is the boundary. Matches `feature-flags.md`.
- **Design tokens = direction A.** Every hex in `directions/confianza_sobria.dart` matches the
  DIRECTIONS.md palette table (light and dark). The 4-pt spacing grid, the 6/8/12 radii and the
  44/28/20/18/17/16/14/13 type scale all match. Contrast pairs are asserted in
  `theme_tokens_test.dart`. Faces are bundled with their OFL notices (no runtime fetch).
- **Network hardening.** `followRedirects: false` plus `HttpsOnlyInterceptor` covers the gap
  where Dart sockets ignore Android NSC and iOS ATS. Prod and staging refuse to boot without an
  HTTPS `API_BASE_URL`, and the Makefile refuses to build such an artifact.
- **Generated client.** Normalizer and kind guard are tested, drift is checked in CI, generated
  code is excluded from analysis and from coverage. Matches ADR-0003.
- **CI.** Actions pinned by SHA, `persist-credentials: false`, `contents: read`, lockfile
  enforced, committed l10n checked, an artifact gate that self-tests before it judges, symbols
  kept as a short-lived private artifact. R8 minify and resource shrinking are on. Matches
  `frontend-build.md`.

## Findings

Severity: **Blocker** (fix before merge) · **Major** (M1 must fix) · **Minor** (fix when the
area is touched) · **Info**.

| ID | Sev | Where | Finding | Action |
|---|---|---|---|---|
| F-01 | Major | `core/router/app_router.dart` redirect | While the first flags answer is pending, `enabled` is `{}`, so a deep link (notification, cold-start restore, M1 post-login redirect) into a module is sent to `/no-encontrado` and never returns once flags arrive. Verified with a probe. | **M1 prerequisite.** Add a boot/splash state: redirect waits (or goes to a splash) while `flagsProvider.isLoading`, keeps the requested location, then resolves. Same mechanism carries the M1 auth redirect. Test first. |
| F-02 | Major | `lib/` (missing) | **No form primitives exist.** M1-11 (register) is the first form; without shared widgets each screen re-implements the house rules and drifts. | **M1 prerequisite**: see "Form primitives M1 needs" below. |
| F-03 | Major | Android manifest / M1-12 | `android:allowBackup` defaults to true and there are no `dataExtractionRules`. Once M1-12 stores refresh tokens, they can land in cloud or device-transfer backups. | M1-12: exclude the token store from backup (`allowBackup="false"` or explicit rules), with a manifest test in the style of `android_manifest_test.dart`. Security gate item. |
| F-04 | Minor | `core/flags/flags_provider.dart` | A failed **first** fetch (cold start offline) returns `{}` as `AsyncData`. Every module stays hidden until the next app resume. No retry, and the UI cannot tell "nothing enabled" from "could not ask". | With F-01: retry with backoff on first-load failure, and let the splash show a retry action. |
| F-05 | Minor | `core/router/app_router.dart` ↔ `features/home/home_routes.dart` | Import cycle: core imports the home feature, and the feature imports `AppRoutes` from core. Not a bug, but it bends the dependency rule (core should not know a feature). | When the bottom-nav shell lands (F-07), move `AppRoutes` to `core/router/app_routes.dart` and register home through the same composition as other slices. |
| F-06 | Minor | iOS | Only the `Runner` scheme exists; there are no per-flavor iOS schemes or configurations (OB6). `make mobile-build-ipa FLAVOR=...` cannot build a flavor. There is no iOS job in CI. | Before the first TestFlight build (REL track): flavor schemes, a macOS CI job with the same artifact gate. |
| F-07 | Minor | `app_theme.dart` `navigationBarTheme` | DIRECTIONS A wants a 3 px top rule on the active item and five labeled tabs. The theme only makes the indicator transparent; there is no shell route yet. | M2/M3, when the second tab exists: `StatefulShellRoute` plus a token-driven nav bar, with tabs gated through `FlagGate`. |
| F-08 | Minor | CI | `security.md` asks for a dependency audit in CI. Pub dependencies are pinned and lockfile-enforced, but nothing scans them (e.g. `osv-scanner` on `pubspec.lock`). | devops-sr follow-up, alongside the Go `govulncheck` job. |
| F-09 | Info | `pubspec.yaml` | `intl: any`. It is resolved by `flutter_localizations` and the lockfile is enforced, so there is no drift risk today. It is the only unpinned direct dependency. | Pin at the next dependency bump. |
| F-10 | Info | DIRECTIONS.md "Next steps" | Step 1 (freeze tokens into `docs/design/DESIGN.md`) and step 2's per-theme golden tests are not done. The contrast assertions cover the color side. | Golden tests arrive with the first real screen (M1-11), light and dark. |
| F-11 | Info | Icons | DIRECTIONS A specifies Lucide outline at 1.75 stroke, 22 px. No icon package yet. | Add when the first icon is needed; check current docs via Context7 and pin it. |
| F-12 | Info | CI | `flags_mock_test.dart` (`--tags mock`) never runs in CI, so the client's wire contract against the Prism mock is only checked locally. | Optional CI job that starts Prism and runs `make mobile-test-mock`. |

### Fixed in this PR (small, test first)

| ID | Was | Fix | Test |
|---|---|---|---|
| X-01 | `Info.plist` declared no localizations, so iOS's own UI (text selection menu, permission prompts, share sheet) stays in English under an es-MX app | `CFBundleLocalizations = [es]` | `test/platform/ios_info_plist_test.dart` (failed first) |
| X-02 | No accessibility guardrail at all | Tap-target (Android 48 / iOS 44), labeled-target and text-contrast guidelines on the existing screens; 200 % text scale renders without overflow | `test/app/accessibility_test.dart`. It passes on today's screens; it exists so every new screen is added to it |
| X-03 | Stale docs: `designDirectionProvider` said "until the owner picks… PR #3"; home placeholder named the dashboard "M4" (it is M6); DIRECTIONS.md said fonts load through `google_fonts` (they are plain bundled assets) | Comments and doc corrected | n/a (docs) |

## Readiness against the house rules

### ui-forms-and-modals.md: form primitives M1 needs (F-02)

Theme groundwork already present: 52 px input and button height, 44 px minimum targets,
`MaterialTapTargetSize.padded`, 17 px input text, a 3:1 `outline` border, and focus and error
borders. Missing, to build under `lib/core/forms/` (widget tests first) before or as the first
task of M1-11:

1. **`AppTextField`**: visible label **above** the field (not Material's floating
   `labelText`), required/optional marked in text, `keyboardType` / `textInputAction` /
   `autofillHints` as required parameters, error wired to semantics. Hint text is never a label.
2. **Validation policy**: required-but-empty errors on submit only; format errors on blur once
   the field has a value; after an error, re-check on every change. This is one controller, not
   `AutovalidateMode.always`.
3. **`FormErrorSummary`**: on a failed submit, a summary at the top that receives focus
   (semantics announce), each item jumps to its field.
4. **`FormFooter`**: one primary (*Guardar*), an optional secondary (*Guardar y registrar
   otro*: keeps the context fields, clears the rest, refocuses the first field), *Cancelar*
   as a text button. The primary is **never disabled**: busy state plus ignoring repeat taps,
   with the server idempotent.
5. **Focus visibility**: a sticky footer must never cover the focused field
   (`Scrollable.ensureVisible` with footer-height padding; keyboard insets).
6. **Unsaved changes**: `PopScope` asks only when the form is dirty.
7. **Undo**: reversible deletes act immediately and offer Undo in a snackbar that does not
   auto-dismiss (or Undo also lives on the row). Confirmation dialogs only for irreversible
   actions (account deletion in M1-13).
8. **Container choice**: 1–4 fields inline, 5–6 a bottom sheet sized to content, more than 6 its
   own page (Registrar ingreso is a page, per DIRECTIONS).

### l10n

ARB-based, `nullable-getter: false`, a single `es_MX` locale with a neutral `es` template,
committed output checked in CI, and the mandatory estimate legend already a key. Now also
declared to iOS (X-01). M1: date and money formatting through `intl` with `es_MX` (tabular
figures are already on in the theme), and no hard-coded strings (add a lint or test when the
first form lands).

### Accessibility

Text scaling is not clamped (good). Targets and contrast are covered by X-02. M1 must:
add each new screen to `accessibility_test.dart`, test at 200 %, and give hero amounts a
`Semantics` label that reads the amount and currency in words ("12 345 pesos con 50
centavos"), not the formatted string.

### Build hardening

Android: obfuscated Dart, split debug info kept out of the artifact, R8 on, artifact gate with
a self-test. Release signing still uses the debug key until REL-04. That is acceptable for CI
artifacts and must not reach a store upload. iOS: see F-06.

## What M1 must address (summary)

1. F-01 boot/splash state in the router (also carries the auth redirect). Test first.
2. F-02 form primitives (list above) before M1-11 screens.
3. F-03 token storage excluded from backups (M1-12), with a manifest test.
4. F-04 retry on first flags fetch failure (with F-01).
5. Every new screen: added to `accessibility_test.dart`, light and dark goldens, and 200 % text.
