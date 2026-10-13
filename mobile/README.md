# Control Contable — mobile (Flutter)

Flutter client for Control Contable. Stack and pins: [ADR-0001](../docs/adr/0001-stack.md)
(mobile section). Toolchain: **Flutter 3.47.7 / Dart 3.13.5**.

## Prerequisites

- Flutter 3.47.7 (`flutter --version`).
- Android: Android SDK (Android Studio). iOS: full Xcode + CocoaPods (owner item OB6).
- The API on port 3000, or the Prism mock on port 3001 (`make mock`), which works before
  the backend does.

## Everyday commands (from the repo root)

| Command | What it does |
|---|---|
| `make mobile-check` | `dart format --set-exit-if-changed .`, `flutter analyze` (0 issues), `flutter test` |
| `make mobile-run` | Runs the dev flavor against `http://localhost:3000` |
| `make mobile-run-mock` | Runs the dev flavor against the Prism mock, `http://localhost:3001` |
| `make mobile-run FLAVOR=staging API_BASE_URL=https://…` | Runs another flavor |
| `make mobile-build-apk FLAVOR=prod API_BASE_URL=https://…` | Obfuscated release APK |
| `make mobile-build-aab FLAVOR=prod API_BASE_URL=https://…` | Obfuscated release app bundle (Play upload) |
| `make mobile-cover` | Tests with coverage; fails below 80% on non-generated `lib/` code (`tool/coverage_gate.dart`) |
| `make mobile-artifact-check FLAVOR=prod` | Fails if the release APK/AAB ships symbols, debug info, `.map` files or `sourceMappingURL` (`tools/mobile/check-release-artifact.sh`; macOS: `READELF=<ndk>/llvm-readelf`) |

CI (`.github/workflows/mobile.yml`) runs format, analyze, `mobile-cover`, then the prod APK +
AAB build and `mobile-artifact-check`. The split debug info (`build/symbols/prod`) is kept as a
private CI artifact for crash de-obfuscation and never ships in the app. An AAB's
`BUNDLE-METADATA/` (native symbol tables, R8 `proguard.map`) is Play-only and allowlisted:
`flutter build appbundle` refuses to finish without it.

An Android emulator reaches the host as `10.0.2.2`: add `API_BASE_URL=http://10.0.2.2:3000`.

## Flavors

| Flavor | Entrypoint | API base URL | Android id |
|---|---|---|---|
| dev | `lib/main_dev.dart` | `http://localhost:3000`; `USE_MOCK=true` → `http://localhost:3001` | `mx.romodevx.control_contable.dev` |
| staging | `lib/main_staging.dart` | `--dart-define=API_BASE_URL=https://…` (required) | `….staging` |
| prod | `lib/main_prod.dart` | `--dart-define=API_BASE_URL=https://…` (required) | `mx.romodevx.control_contable` |

`API_BASE_URL` always wins when set. staging and prod refuse to boot without it or with plain
HTTP (`AppConfigError`), so no environment host is baked into the source.

Android product flavors live in `android/app/build.gradle.kts`. **iOS schemes are pending**:
they need a full Xcode (OB6). Until then, run iOS with `-t lib/main_<flavor>.dart` and no
`--flavor`.

## Layout

```
lib/
  main_<flavor>.dart     entrypoints → app/bootstrap.dart
  app/                   root widget (MaterialApp.router), bootstrap
  core/
    config/              AppFlavor, AppConfig, appConfigProvider (overridden at boot)
    network/             dioProvider (the one HTTP client)
    router/              appRouterProvider (go_router), AppRoutes, NotFoundScreen
    theme/               design tokens (ThemeExtension) and the Material mapping
  features/<module>/     one folder per module slice: <module>_routes.dart, presentation/,
                         application/, data/ — a disabled module contributes no routes
  l10n/                  ARB files (es template, es_MX) + generated AppLocalizations
```

## Design tokens

Widgets never hard-code a color, radius or face; they read `context.tokens` (`AppTokens`, a
`ThemeExtension`) or the Material theme built from it in `core/theme/app_theme.dart`.

- `core/theme/directions/confianza_sobria.dart`: direction **A · Confianza sobria**
  ([DIRECTIONS.md](../docs/design/DIRECTIONS.md)), light and dark.
- `designDirectionProvider` picks the active direction. To change the look, add
  `directions/<name>.dart` and return it there. No widget changes.
- `test/core/theme/theme_tokens_test.dart` asserts every documented contrast pair passes AA,
  so a new direction cannot ship a failing pair.
- Fonts (Public Sans, Source Serif 4; SIL OFL 1.1, latin subset) are bundled in
  `assets/fonts/`: no runtime font fetching.

## Localization

es-MX is the only locale. Copy lives in `lib/l10n/app_es.arb` (template); `app_es_MX.arb`
holds only Mexican overrides. `flutter gen-l10n` regenerates `app_localizations*.dart`
(also run by `flutter pub get` / `flutter test`); commit the result.

## Release hardening

- Dart code: `--obfuscate --split-debug-info=build/symbols/<flavor>`. Symbols stay on the
  build machine and are never bundled (`*.symbols` is git-ignored).
- Android: R8 minify + resource shrinking on release.
- `avoid_print` is an error: production code does not print.
- The CI artifact gate (no symbols or maps inside the APK/IPA) is P0-25.

## Analysis

`analysis_options.yaml`: `very_good_analysis` 11.0.0, `strict-casts`, `strict-inference`,
`strict-raw-types`, and `riverpod_lint` 3.1.9 as an analysis server plugin.
