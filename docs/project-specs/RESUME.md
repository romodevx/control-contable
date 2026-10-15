# Resume here

Paused on 2026-10-15 by the owner (other projects first). Planned restart: the week of
2026-10-19. Read this page first, then [FOLLOW-UPS.md](FOLLOW-UPS.md).

## Where the project stands

`dev` at `38ceac3`, contract `0.17.0`, newest ADR `0012`. No open PRs. Only `dev` and `main`
exist, locally and on GitHub. `main` has not been updated; moving to `main` (release) needs the
owner's explicit approval.

| Module | State on `dev` | What is left |
|---|---|---|
| P0 platform, M1 auth, M2 taxpayers, M3 incomes, M4 expenses, M6 dashboard | Done, gates passed | M1-06 real email provider (see "Owner decisions") |
| M5 tax engine (engine 1.1.1) | Done, gates passed | M5-33 CPA sign-off |
| M8 history | Done (QA, device and iOS walks, gates, docs) | Manual TalkBack/VoiceOver pass |
| Withholdings (WH) | Done (QA, gates, docs) | WH-17 CPA review |
| Annual return (ANN) | Done (QA, gates, docs) | ANN-27 CPA sign-off |
| M9 reminders | Done (QA, gates, docs, security follow-ups) | Live push needs Firebase + APNs |
| M7 subscriptions | Prep done: per-scheme auth, `access` on `GET /v1/me` | Everything else waits on pricing and store accounts |
| Hardening | Email outbox, key rotation (ADR-0012), income basis in the contract, collection date on accrual incomes, iOS walk | Items in FOLLOW-UPS.md |

## Owner decisions and inputs pending

1. **Email provider (M1-06).** Proposal agreed in conversation, not yet confirmed:
   - Now (tests, closed beta): a **new Gmail account dedicated to the app**, sending through
     Google SMTP with an app password (2-step verification on). About 500 emails a day.
     Never the owner's personal Gmail.
   - Before real users: an **own domain** (`.mx` or `.com`) with Resend or Amazon SES, SPF and
     DKIM set up. A `@gmail.com` sender cannot be used through SES/Resend (fails DMARC).
   - The `EmailSender` port and the outbox are ready; only the adapter and its secret change.
   - The app sends: the 6-digit verification code (register and resend) and the
     "someone tried to sign up with your email" notice. There is no password reset today.
2. **Pricing + RevenueCat, Apple Developer and Google Play accounts** (M7).
3. **Firebase project with the APNs key** (live push on Android and iOS).
4. **CPA:** sign `docs/reviews/CPA-REVIEW.md` (M5-33, WH-17, ANN-27, OQ-R1 Friday extension,
   monthly half-centavo display items).
5. **Lawyer:** terms of service and aviso de privacidad (replace the `.invalid` placeholders).
6. **Manual TalkBack/VoiceOver pass** on a real phone.
7. **Close the open GitGuardian alerts** in the GitHub dashboard.
8. **ADR-0012 acceptance** (architect) before the first key rotation in a shared environment.
   Never rotate `RFC_ENC_KEY` without following `docs/runbooks/key-rotation.md`.

## First steps when resuming

1. `git checkout dev && git pull` and skim `git log --oneline -15`.
2. Read [FOLLOW-UPS.md](FOLLOW-UPS.md); pick `open-unblocked` rows, or whatever the owner has
   unblocked from the list above.
3. If the owner confirmed the Gmail option: implement the SMTP adapter behind `EmailSender`
   (TDD, secret from env/secret manager, never in the repo), then QA the register flow.
4. Run the delivery loop as before: branch from `dev` → implement (Codex first) → PR → tech-lead
   and/or security review → QA → squash-merge to `dev` → `~/.claude/bin/git-prune-merged --apply`.

## Machine notes (owner's Mac)

- Git remote uses the SSH alias `git@github.com-personal:romodevx/control-contable.git`
  (plain `github.com` has no key configured; the old `github.com-inteligente` key was removed).
- Docker Desktop was not starting from the terminal. QA used a native Homebrew
  `postgresql@18` cluster per agent (own data dir and port); steps are in
  `.gstack/qa-reports/qa-report-withholdings-2026-10-14.md`.
- Xcode 27 is active with the iOS 27 simulator and CocoaPods. Android `cmdline-tools` and the
  SDK licenses are missing locally (CI has them).
- The repo is public: never commit secret-shaped values; tests generate them at runtime. The
  pre-commit gitleaks hook is installed with `make hooks`.
