# Runbook: reminders in a shared environment (`module_reminders`)

How to take push reminders from "built against fakes" to ON in staging or production, and the
one recurring task they bring: seeding next year's holidays every December. What the module
does for the user: [module page](../modules/reminders.md).

| | |
|---|---|
| Owner | devops-sr (Firebase, secrets, deploys), db-sr (holidays migration), mobile release (config files, APNs), owner (Firebase and Apple accounts: OB9, OB7) |
| Blocked by | **OB9** Firebase project, **OB7** Apple Developer account (APNs key), **OB6** Xcode (iOS build and walk) |
| Gate | M9-13 tech lead **PASS**, M9-14 security **PASS** (2026-10-15, [PR #170](https://github.com/romodevx/control-contable/pull/170), follow-ups [PR #172](https://github.com/romodevx/control-contable/pull/172)) |
| Calendar | Holidays for year N+1: open the task in **early December** of year N, merge before **20 December** (the December period is due in January of N+1) |

Never commit a Firebase config file, a service account key, an APNs `.p8` key or a push token
key. The repository is public and scanned (gitleaks pre-commit hook, GitGuardian). Every
placeholder below is in angle brackets; replace it in the console or the secret manager, never
in a file.

## 1. What runs where

| Piece | Needs | Without it |
|---|---|---|
| API (`cmd/api`, `cc_app`) | `PUSH_TOKEN_ENC_KEY`, `PUSH_TOKEN_INDEX_PEPPER` (seal the token on `registerDevice`); the holidays tables (read once at boot) | Outside dev it refuses to start |
| Worker (`cmd/worker`, `cc_worker`) | The **same** two push token keys (it opens what the API sealed); `FCM_PROJECT_ID`; FCM credentials (`FCM_CREDENTIALS_JSON` or the runtime's service account) | Keys missing outside dev: refuses to start. No `FCM_PROJECT_ID`: starts, logs `push disabled`, and the reminder job **skips every run without claiming** (nothing is consumed). Unusable credentials: refuses to start |
| App (release build per flavor) | `google-services.json` (Android), `GoogleService-Info.plist` (iOS), the Push Notifications capability and the APNs key in Firebase (iOS) | Boots normally; the reminders screen says "Notificaciones no disponibles en esta versión de la app." and registers nothing |
| Database | `cc_auth_definer` (owns `claim_push_token`) and `cc_worker_definer` (owns `due_reminder_users` and the purge functions) **before** `migrate up` | The M9 migrations refuse to run, with a hint |

## 2. Firebase project (OB9, once per environment)

One Firebase project per shared environment keeps staging pushes away from production devices.

1. Create the project in the Firebase console (`<firebase-project-id>`; it is a Google Cloud
   project id: 6–30 characters, lowercase letters, digits and hyphens, the worker checks the
   shape at boot).
2. Check that the **Firebase Cloud Messaging API (V1)** is enabled (Project settings → Cloud
   Messaging). The legacy server-key API is not used.
3. Register the apps, one per flavor this environment serves:

   | Platform | Identifier |
   |---|---|
   | Android prod | `mx.romodevx.control_contable` |
   | Android staging | `mx.romodevx.control_contable.staging` |
   | Android dev | `mx.romodevx.control_contable.dev` |
   | iOS | `mx.romodevx.controlContable` (one scheme today; per-flavor bundle ids arrive with the iOS schemes, OB6) |

4. **Sender credentials for the worker.** Preferred: run the worker as a dedicated service
   account and grant it the **Firebase Cloud Messaging API Admin** role (`roles/firebasecloudmessaging.admin`) on the project; leave
   `FCM_CREDENTIALS_JSON` unset and the worker uses Application Default Credentials (no key
   exists to leak). Only when the worker runs outside Google Cloud: create a key for that
   service account and store the JSON **directly** in the secret manager as
   `FCM_CREDENTIALS_JSON`; never download it to a laptop folder that syncs, never paste it in a
   ticket. Grant no other role.

## 3. APNs key (iOS, OB7)

1. In Apple Developer → Certificates, Identifiers & Profiles → Keys, create a key with
   **Apple Push Notifications service (APNs)** enabled. The `.p8` file downloads **once**.
2. Upload it in Firebase → Project settings → Cloud Messaging → Apple app configuration, with its
   Key ID and the Team ID. One key serves sandbox and production.
3. Keep the `.p8` in the team's secret manager (it cannot be downloaded again), then delete the
   local copy.
4. On the App ID, enable **Push Notifications**; the Runner target needs the capability (the
   `aps-environment` entitlement) in Xcode (OB6).

## 4. App config files per flavor

Download each app's file from Firebase → Project settings → Your apps. The files identify the
project, they are not secrets in the cryptographic sense, but the repository never holds them
(`test/platform/firebase_config_test.dart` fails if one appears in the tree).

| Platform | Where the file goes (on the build machine only) | Notes |
|---|---|---|
| Android, every flavor in one file | `mobile/android/app/google-services.json` | Must contain the clients of every flavor the build produces |
| Android, per flavor | `mobile/android/app/src/{dev,staging,prod}/google-services.json` | Each flavor's build needs its own file |
| iOS | `mobile/ios/Runner/GoogleService-Info.plist`, added to the Runner target | One plist until the per-flavor schemes (OB6) |

The Gradle plugin `com.google.gms.google-services` is applied only when an Android file exists,
so a build without it still succeeds and ships the "no disponible" state. Store each file in the
secret manager and have the release pipeline write it before `flutter build`; that step belongs
to REL-04 and is **not wired yet**. Until then, a release build is made on a machine where the
file was placed by hand and removed afterwards.

## 5. Database roles and secrets

**Definer roles.** Both are bootstrap roles created once per environment before `migrate up`:
follow [records-purge §2](records-purge.md#2-once-per-environment-the-definer-roles) (it creates
`cc_auth_definer` and `cc_worker_definer` and checks their posture). Then `migrate up` as
`cc_owner` applies `20261021100000_holidays`, `20261021101000_reminders`,
`20261022100000_reminders_job_reads` and `20261022101000_notification_log_purge`.

**Push token keys.** Generate each one straight into a new secret, never printed or written to a
file:

```sh
openssl rand -hex 32 | <add as the first version of the environment's PUSH_TOKEN_ENC_KEY secret>
openssl rand -hex 32 | <add as the first version of the environment's PUSH_TOKEN_INDEX_PEPPER secret>
```

| Name | Kind | Value | Read by |
|---|---|---|---|
| `PUSH_TOKEN_ENC_KEY` | secret | exactly 32 bytes, hex | API and worker (same version) |
| `PUSH_TOKEN_INDEX_PEPPER` | secret | at least 32 bytes, hex | API and worker (same version) |
| `PUSH_TOKEN_ENC_KEY_ID` | plain variable | integer, default `1` | API and worker |
| `FCM_PROJECT_ID` | plain variable | `<firebase-project-id>` | worker |
| `FCM_CREDENTIALS_JSON` | secret, optional | service account key JSON (§2 step 4) | worker |

- Both keys are distinct per environment and distinct from every auth and RFC key: the API and
  the worker refuse to start when two of them are equal.
- **There is no rotation path for `PUSH_TOKEN_ENC_KEY` yet** (one keyring entry, M9-05 N2).
  Replacing it makes every stored token unreadable: the job then skips those devices until each
  app re-registers on its next sign-in or token refresh. Do not rotate it except after a leak.
  Rotating `PUSH_TOKEN_INDEX_PEPPER` also changes collapse ids (at most one duplicate that day).

## 6. Every December: seed next year's holidays

The due-date rule reads the `holidays` / `holiday_years` tables. A year that is not in
`holiday_years` is **unseeded**: `getTaxCalendar` answers `422 calendar_unavailable` for it, and
any due date that falls in it (December's period, due in January) is computed **weekends-only**
and logs `event=reminders_calendar_unseeded` (API and worker). Seeded today: **2026, 2027**. Do
this every year next to the [yearly tax parameter update](tax-params-update.md#11-holidays-for-the-next-year-module_reminders),
in the same window.

1. **Read the sources for year N+1:**
   - CFF art. 12 ¶1 (current text on diputados.gob.mx): 1 January; first Monday of February; third
     Monday of March; 1 and 5 May; 16 September; third Monday of November; 25 December; and
     1 December every six years when the federal executive changes (**next: 2030**). Check the
     article's last reform date: a reform can add or move a day.
   - RMF of year N+1, rule 2.1.6 (days SAT declares non-business: Holy Thursday and Friday). The
     RMF is published in late December; if it is not out yet, project Holy Thursday/Friday from
     Easter, say so in the row's name ("(proyectado; RMF N+1 no publicada)") and correct it in
     January with another migration.
   - Do **not** add SAT vacation periods (CFF 12 ¶2: business days for payments), RFC-digit grace
     days or the Friday extension (OQ-R1).
2. **Write the migration** `backend/migrations/<YYYYMMDDHHMMSS>_holidays_<N+1>.sql`, following
   `20261021100000_holidays.sql`: one `holiday_years` row (`seeded_on` = today, a `source_note`
   naming the CFF reform date and the RMF), then one `holidays` row per day with `name` (Spanish),
   `basis` (`cff_12` or `rmf_2_1_6`), `moves_due_date` stated on every row, an `https` `source_url`
   and `verified_on`. The `Down` deletes that year's `holidays` rows, then its `holiday_years`
   row. A wrong day is fixed **forward** with a new migration (`UPDATE`/`DELETE` of that row),
   never by editing a merged one.
3. **Tests**, test first: extend `migrations/holidays_integration_test.go`
   (`TestHolidays_SeedCoversTheResearchYears`) with the new year's days, and the fiscal kernel's
   golden `TestMonthlyDueDateGolden` (`internal/kernel/fiscal/duedate_test.go`) with the 12 periods
   whose due date falls in N+1. A day 17 that now meets a holiday changes an estimate's `due_date`:
   say so in the PR (no `EngineVersion` bump: the figures are unchanged).
4. **Record the research**: add the year's table to
   [tax-regimes research §A.4.1](../research/tax-regimes-mx-2026.md) with the moved dates and the
   Fridays kept.
5. **Ship**: tech-lead review, `make -C backend test-integration` and the CI required checks
   green. Deploy order: `migrate up`, then **restart the API** (it reads the holidays once at
   boot); the worker reads them on every run.
6. **Check**: `GET /v1/tax-calendar?year=<N+1>` answers `200` (it was `422`), and no
   `reminders_calendar_unseeded` event appears for a year ≤ N+1.

## 7. Enable `module_reminders`

`module_reminders` stays **OFF** in staging and production until every item holds. Whoever flips
the flag verifies each one and records it in the deploy PR.

- [ ] **Gates**: M9-13 PASS and M9-14 PASS (done, 2026-10-15); the deploy security gate (REL-09).
- [ ] **Prerequisite**: `module_taxpayers` is enabled in this environment (and through it its own
      checklist). `module_tax_engine` is optional: without it reminders carry names only.
- [ ] **Firebase** (§2) and **APNs** (§3) exist for this environment.
- [ ] **Definer roles** exist and the M9 migrations ran (§5).
- [ ] **Secrets** in the secret manager (§5), the same versions wired to the API and the worker;
      `APP_ENV` is not `dev`.
- [ ] **Holidays** seeded for the current year and the next one if it is December (§6).
- [ ] **Release builds** of the app carry the config files of this environment's flavor (§4) and
      pass `mobile-artifact-check`.
- [ ] **Live push check** passed on one Android and one iOS device (§8).
- [ ] **Monitoring**: alerts on `event=reminders_calendar_unseeded` and
      `event=reminders_push_disabled`; the worker's `"job":"reminders"` failures visible.

**Deploy order**, then the flag:

1. Bootstrap roles (once), then `migrate up` as `cc_owner`.
2. Deploy the **API** with the push token keys. Check `GET /healthz`.
3. Deploy the **worker** with the same keys, `FCM_PROJECT_ID` and its credentials. Its startup log
   must say `push enabled` (with `service_account_key` false when it uses its own service account)
   and must **not** say `push disabled` or `reminders disabled`.
4. Publish the app builds (§4).
5. Enable the flag with the environment's flag tooling (`cmd/flags` as `cc_owner`, credentials
   from the secret manager):
   `make -C backend flags ARGS="enable module_reminders"`. `GET /v1/flags` lists it within 5 s;
   the worker reads it on its next run.

Getting steps 3 and 5 out of order is no longer destructive (M9-13): a worker without push
claims nothing, so the reminders of that day are still sent once push is configured, as long as
it is the same day. A day with the flag ON and no push worker at all is skipped (never sent
late).

## 8. First live push check

The walk QA M9-12 could not do (OB9). On a staging build, one Android and one iOS device, a test
account with two taxpayers:

1. *Recordatorios* → "Recibir recordatorios de pago" on → the OS prompt appears; allow it. A
   `devices` row exists for the account (count only, as the operator: never select the token
   columns).
2. Deny on a second install: the screen shows the blocked notice and "Abrir ajustes del sistema";
   no device row.
3. Set "Hora del aviso" a few minutes ahead on a day with a reminder (or use the early day): one
   notification arrives, on the **lock screen** names only; with "Mostrar montos…" on, the
   amounts. Exactly one, even with two worker instances running.
4. Tap it: the estimate of that period opens.
5. Sign out, sign in as another user on the same phone with reminders off: the next reminder of
   the first user does not reach this phone.
6. Uninstall the app: the next send logs `devices_unregistered` and the device row is gone.

Record the result in the PLAN (OB9 live checks of M9-07, M9-09, M9-12) and in the module page.

## 9. Check it is working

On the worker's logs, filter on `"job":"reminders"`:

- `"msg":"job finished"` every 5 minutes on exactly one instance (the leader).
- `event=reminders_push_disabled` (once per process, then debug lines): this worker has no
  `FCM_PROJECT_ID`. Nothing is consumed; fix the configuration the same day.
- `event=reminders_calendar_unseeded` with a `year`: seed that year (§6).
- `"msg":"job failed"`: read the error. A deadlock or a timeout during a leader failover is
  harmless (the next run repeats; the UNIQUE key prevents a second send). Anything else is a bug.

Logs carry user ids and periods only, never a token, a name or an amount. Delivery outcomes are
in `notification_log` (`status`, `devices_ok`, `devices_failed`, `error_class`); query counts,
never per-user rows, unless investigating one user's report with their consent.

## 10. Turn it off

`make -C backend flags ARGS="disable module_reminders"`: the routes answer `404`, the app hides
the entry and the job stops sending on its next run. Devices, settings and the log are kept, so
turning it back on resumes with each user's settings; past days are not caught up. To stop pushes
without touching the flag, unset `FCM_PROJECT_ID` on the worker and redeploy it: nothing is
claimed while it is unset. The records purge keeps running either way.
