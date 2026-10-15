# Follow-ups register

Cross-module follow-ups that outlive the PR that raised them. The per-module tables in
[PLAN-control-contable.md](PLAN-control-contable.md) stay the source of record for module tasks.
This file collects what is still open across modules, what blocks it and who picks it up.

Verified against `origin/dev` at `c3baf21` (2026-10-15).

**Status values**

- `open-unblocked`: an agent can start now.
- `blocked-on-owner`: waits on an owner decision, account, person or manual pass (PLAN §6, `OB*`).
- `blocked-on-env`: waits on something that has to happen in each environment first.
- `accepted-risk`: security accepted it; act only when the stated trigger happens.

**Upkeep.** When a PR closes an item, delete its row in that PR and cite the PR in the commit
body. An item added here names its origin PR. Stale rows are bugs.

## Auth and email (M1)

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| Real email provider adapter (M1-06, OB2). The outbox is done ([#177](https://github.com/romodevx/control-contable/pull/177)), so this is only the provider and its secret. `module_auth` cannot be ON in a shared environment without it | PLAN M1-06 | blocked-on-owner | backend-sr |
| Outbox delivery: the run timeout equals the interval (`DeliveryRunTimeout = DeliveryInterval`, `backend/internal/auth/usecase/delivery.go`), so a slow provider cuts a batch of 20 and each unsent row loses an attempt to the 2-minute lease. Add a per-send timeout and size the batch to `RunTimeout / sendTimeout`, or stop counting attempts that never reached the provider. Do it with M1-06 | [#177](https://github.com/romodevx/control-contable/pull/177) security note 1 | blocked-on-owner (lands with OB2) | backend-sr |
| Real legal documents (terms, aviso de privacidad) replace the `.invalid` placeholders; consents re-affirmed at verify (OB10) | PLAN M1 follow-ups (#34, #40) | blocked-on-owner | backend-sr + owner (lawyer) |

Dropped: #177 security note 2 (outbox key rotation drops in-flight rows). Done in
[#180](https://github.com/romodevx/control-contable/pull/180): `AUTH_OUTBOX_PREVIOUS_KEY` slot and
the drain step in [runbooks/key-rotation.md](../runbooks/key-rotation.md).

## Subscriptions (M7)

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| Pricing (OB1), RevenueCat account (OB8), Apple Developer and Google Play accounts (OB7). Per-scheme auth ([#176](https://github.com/romodevx/control-contable/pull/176)) and access on `GET /v1/me` are done; M7-04 live check, M7-08, M7-10 and M7-12 wait on these | PLAN M7, §6 | blocked-on-owner | devops-sr (store config), backend-sr, frontend-sr |

## Reminders and push (M9)

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| Firebase project with the APNs key (OB9): live push walk on Android and iOS ([runbook §8](../runbooks/reminders.md#8-first-live-push-check)) before `module_reminders` is ON in a shared environment | PLAN M9 (OB9 live) | blocked-on-owner | qa-sr + owner |

## Tax engine, withholdings and annual (CPA)

All items below are in [CPA-REVIEW.md](../reviews/CPA-REVIEW.md) and wait on the CPA (OB5).

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| M5-33: CPA sign-off of the monthly golden suite; un-skip ⚠️ cases or record decisions | PLAN M5-33 | blocked-on-owner | qa-sr |
| WH-17: CPA review of the withholding rules (A1–A4, A6–A8, A10) and withheld-IVA timing (OQ-W4); each answer bumps `WithholdingTableVersion`, then a backfill run | PLAN WH-17 | blocked-on-owner | qa-sr |
| ANN-27: CPA sign-off of the annual goldens; un-skip AN-04 / AN-07 PTU | PLAN ANN-27 | blocked-on-owner | qa-sr |
| OQ-R1: Friday extension (CFF art. 12 ¶5). Today the earlier Friday is kept and dates are "sujetas a confirmación" | PLAN M9, MODULE-reminders OQ-R1 | blocked-on-owner | qa-sr + pm |
| Monthly half-centavo items: ISR as shown (rounded before credits) and the 601 "base × 30 %" display centavo | [#156](https://github.com/romodevx/control-contable/pull/156) | blocked-on-owner | qa-sr |

## Data and migrations

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| Drop legacy `incomes.counterparty_type` in its own migration. Ledger reads already moved to `counterparty_kind` ([#178](https://github.com/romodevx/control-contable/pull/178)); writes still mirror the column (`20261018100000_withholdings_backfill_worker.sql` grants it). Needs the WH-08 backfill applied in every environment first ([runbook](../runbooks/withholdings-backfill.md)) | PLAN WH-08 contract step | blocked-on-env | db-sr + backend-sr |
| Create the NOLOGIN `cc_worker_definer` role once per environment before the backfill and key re-seal migrations ([runbook §0](../runbooks/withholdings-backfill.md#0-once-per-environment-the-cc_worker_definer-role)) | WH-08, [#180](https://github.com/romodevx/control-contable/pull/180) | blocked-on-env | devops-sr + db-sr |

## Key rotation (WP3, ADR-0012)

From the [#180](https://github.com/romodevx/control-contable/pull/180) security and tech-lead review
(non-blocking). All are still open on `dev`.

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| ADR-0012 is still **Proposed**. The architect accepts or amends it before any shared-environment rotation | #180 note 6 | open-unblocked | architect |
| Dedicated definer role for `taxpayers_key_reseal`. The policy is gated on `app.key_reseal_scan`, which any caller can set for the session (`backend/migrations/20261024100000_key_reseal.sql`); today safe only because every other `cc_worker_definer` function on taxpayers filters explicitly | #180 note 1 | open-unblocked | db-sr |
| Carry the re-seal cursor across runs. Each run restarts from `""`, so more than 10,000 unreadable rows at the front (50 pages × 200) stop it from reaching later rows. Or document the limit in the runbook | #180 note 2 | open-unblocked | backend-sr |
| `LoadResealKeys` checks RFC and email keys differ from each other, not from the push and outbox keys or the peppers on the worker | #180 note 3 | open-unblocked | backend-sr |
| A re-seal bumps `users.updated_at` and `taxpayers.updated_at` through their triggers (cosmetic today, no ETags) | #180 note 4 | open-unblocked | db-sr |
| Pepper-rotation ADR: lookups try both versions; `email_hash` and `rfc_hash` stay unique across versions | #180 note 5 | open-unblocked | architect |

## Mobile and accessibility

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| Manual TalkBack pass (real Android device) and VoiceOver (iOS) for the dashboard, history and annual screens | PLAN M6 FU6, M8 QA §5, ANN QA §7 (OB6) | blocked-on-owner | qa-sr + owner |
| M8 TalkBack focus return: widget test that focus returns to the history row after pop (no such test under `mobile/test/features/history/` yet) | PLAN M8 TL obs. 3, [#149](https://github.com/romodevx/control-contable/pull/149) | open-unblocked | frontend-sr |
| Android local toolchain on the dev machine: `cmdline-tools` and SDK licenses missing (CI is fine) | Coordinator note, M8/M9 QA | open-unblocked | devops-sr |

## Repository hygiene

| Item | Origin | Status | Owner agent |
|---|---|---|---|
| Close the open GitGuardian alerts in the GitGuardian dashboard (test secrets are now generated at runtime) | Coordinator note | blocked-on-owner | security + owner |

## Accepted risks

Recorded so nobody re-opens them by accident. Act only on the trigger.

| Risk | Origin | Trigger to act | Owner agent |
|---|---|---|---|
| Reminders definers reuse shared owners instead of a dedicated `cc_device_definer`: `claim_push_token` is owned by `cc_auth_definer`, `due_reminder_users` (feeds the `notification_log` exactly-once run) by `cc_worker_definer` (`20261021101000_reminders.sql`) | [#170](https://github.com/romodevx/control-contable/pull/170) security gate, pinned by `TestReminderFunctions_DefinerReach` | A second definer that reads or writes `devices`, or dynamic SQL in any `cc_auth_definer` function | db-sr + security |
| Flag cache-miss timing: a disabled route reads the DB once per TTL, a nonexistent one never does (tiny oracle) | [#19](https://github.com/romodevx/control-contable/pull/19) review FU1 | Flags become user-sensitive, or a pre-warm/background refresh is added for another reason | backend-sr |
