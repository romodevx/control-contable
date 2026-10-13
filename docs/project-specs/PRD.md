# PRD — Control Contable

| Field | Value |
|---|---|
| Status | DRAFT v2 (Phase 1 Discovery), pending owner approval |
| Owner | Andres (Product Owner) |
| Date | 2026-10-13 |
| Related | [MODULES.md](MODULES.md), `docs/research/tax-regimes-mx-2026.md` (in progress) |
| Changes v2 | Personas morales in scope; `taxpayer` entity separate from user; multi-taxpayer per subscription; RevenueCat IAP; monorepo layout |

## 1. Problem

Mexican taxpayers, *personas físicas* and small *personas morales*, must make monthly
provisional ISR payments and monthly IVA payments, due on day 17 of the following month.
Most do not know during the month how much they will owe. They hear it late from their
accountant or the SAT portal, and cash flow takes the hit. No simple mobile tool turns
"what I earned and spent this month" into "this is roughly what I will owe, and why". The
problem gets worse for anyone who looks after more than one RFC: their own, a spouse's,
or their company's.

## 2. Product summary

A Flutter mobile app backed by a Go monolith and PostgreSQL.

- A **user** (login identity) registers and manages one or more **taxpayers**
  (*contribuyentes*). Each taxpayer has an RFC, a type (*física* or *moral*) and one or more
  tax regimes.
- Onboarding asks **persona física or persona moral FIRST**, then RFC, then regime(s)
  (filtered by type).
- The user manually records incomes and expenses. Each record belongs to exactly one
  taxpayer + regime, and each expense has a category.
- The app shows an **estimate** of monthly IVA and provisional ISR per regime and per
  taxpayer, with a breakdown that explains each step, charts and history.
- Push reminders warn before the day-17 deadline.
- Monetization: a 7-day free trial with caps, then weekly, monthly or yearly plans through
  in-app purchases (RevenueCat). **One subscription covers every taxpayer the user
  manages**. There is no per-RFC or accountant pricing.

Repository: monorepo with `backend/` (Go, Clean Architecture: `domain/ usecase/ adapter/ infra/`),
`mobile/` (Flutter) and `docs/`.

## 3. Personas

| Persona | Taxpayer(s) | Regime (expected) | Need |
|---|---|---|---|
| **Freelance professional** (designer, developer, doctor) | 1 física | 612 or 626 RESICO PF | Know monthly ISR/IVA; understand which expenses lower it |
| **Small landlord** | 1 física | 606 Arrendamiento | Compare blind deduction (35%) with actual expenses |
| **Platform worker** (Uber, Didi, Rappi, Airbnb, ML seller) | 1 física | 625 Plataformas | See what the platform already withheld and what is still owed |
| **Mixed-income employee** | 1 física | 605 + one of the above | Salary is informational; estimate the side income |
| **Small-company owner** | 1 moral (SA de CV / SAS) + own física | 601 General PM or 626 RESICO PM; plus a física regime | Estimate the company's provisional ISR (coeficiente de utilidad) and IVA, and their own |
| **Household / informal bookkeeper** (owner, relative, or accountant helping others) | Several físicas and/or morales | Mixed | Switch between RFCs quickly; one subscription; see per-RFC estimates and deadlines |

All personas are non-accountants or time-poor, phone-first, Spanish-speaking and
price-sensitive. The accountant persona uses the app as an **informational** helper, not as
practice-management software (see NG7).

## 4. Goals and non-goals

### Goals
- G1. A user gets a first monthly IVA + ISR estimate in under 5 minutes from install.
- G2. Every estimate can be explained: a breakdown shows each step, the rule applied and its legal basis.
- G3. **Every SAT regime for personas físicas and personas morales** is in the catalog,
  grounded in official sources (LISR, LIVA, CFF, RMF 2026, c_RegimenFiscal catalog). Each
  regime is a pluggable strategy in the tax engine. Each regime is classified as *full
  estimate*, *partial estimate* or *informational only*, with an explanation of why (§5).
- G4. A user manages several taxpayers under one account and one subscription. Estimates
  are shown per regime, per taxpayer, plus a monthly total per taxpayer.
- G5. Sustainable revenue through IAP subscriptions.

### Non-goals (v1)
- NG1. **Not tax advice** and not a substitute for an accountant. Nothing is filed or submitted to SAT.
- NG2. No CFDI issuance, no CFDI/XML download or import, no SAT portal connection (e.firma / CIEC).
- NG3. No annual declaration, no DIOT, no payroll, no electronic accounting (*contabilidad electrónica*), no IEPS.
- NG4. No web app. No Stripe or Mercado Pago (IAP only, via RevenueCat).
- NG5. No bank sync or Open Finance.
- NG6. No verification that the user is entitled to an RFC (informational app; see Q9).
- NG7. No accountant features: no client invitations, no shared access between users, no per-RFC billing.

## 5. Regimes (TO BE CONFIRMED by research)

The final list, rates, tariffs, compatibility rules, cash vs. accrual basis and citations
come from `docs/research/tax-regimes-mx-2026.md`. The table below is the working hypothesis.
"Treatment" is a PM proposal for research to validate.

| Code | Regime | Type | Proposed treatment |
|---|---|---|---|
| 605 | Sueldos y Salarios | F | Informational (employer withholds) |
| 606 | Arrendamiento | F | Full (blind 35% vs. actual deductions; IVA where it applies) |
| 607 | Enajenación o Adquisición de Bienes | F | Informational (per-event, not monthly) |
| 608 | Demás ingresos | F | Partial / informational |
| 610 | Residentes en el Extranjero sin EP en México | F, M | Informational |
| 611 | Dividendos | F | Informational (annual) |
| 612 | Actividad Empresarial y Profesional | F | Full (cumulative YTD provisional ISR; monthly IVA) |
| 614 | Intereses | F | Informational (withheld) |
| 615 | Obtención de premios | F | Informational (withheld) |
| 616 | Sin obligaciones fiscales | F | Informational (no estimate) |
| 621 | Incorporación Fiscal (legacy) | F | Research: active only if transitional cases remain |
| 625 | Plataformas Tecnológicas | F | Full (definitive vs. non-definitive withholding option) |
| 626 | RESICO | F | Full (rate on collected income, no deductions; IVA monthly) |
| 601 | General de Ley Personas Morales | M | Full (provisional ISR = coeficiente de utilidad × YTD income; IVA monthly) |
| 603 | Personas Morales con Fines no Lucrativos | M | Partial (IVA if taxed activities; ISR generally not provisional) |
| 620 | Sociedades Cooperativas de Producción (deferral option) | M | Research |
| 622 | Actividades Agrícolas, Ganaderas, Silvícolas y Pesqueras | M | Research (exemptions, reduced rate) |
| 623 | Opcional para Grupos de Sociedades | M | Informational (group consolidation is out of reach of manual entry) |
| 624 | Coordinados | M | Research |
| 626 | RESICO Personas Morales | M | Full (cash-basis provisional ISR on profit; IVA monthly) |

Honest constraint (Clean Coder): "all regimes" means **all regimes are in the catalog and
selectable**. It does not mean all of them get a monthly number. Regimes whose tax is
withheld at source, computed per event, annual-only or group-level are labeled
*informational* in the app with a plain-language explanation.

Rules the product must respect (subject to research):
- **IVA** is on a cash basis (*flujo de efectivo*) for everyone.
- **ISR** is on a cash basis for most físicas and for RESICO PM. For PM 601 it is accrual-based.
- Incomes therefore carry both `invoiced_on` (accrual) and `collected_on` (cash). The engine picks the date it needs.
- A taxpayer may hold several regimes at once. Some combinations are invalid (e.g., RESICO
  PF exclusions). The compatibility rules are data, not code.
- PM 601 needs prior-year inputs: the *coeficiente de utilidad*, plus pending losses and
  prior provisional payments made this year.
- 612 needs year-to-date figures. The engine takes these as **tax inputs** per
  taxpayer + regime + fiscal year.

## 6. User stories and acceptance criteria

IDs map to modules in MODULES.md. "Mandatory legend" means the exact text
**"Esto es un aproximado. No todos los gastos aplican para tu giro comercial."**

### Auth (`module_auth`)
- **US-A1** As a new user, I register with email + password.
  - AC: the server enforces the password policy. A duplicate email returns a generic
    response, so accounts cannot be enumerated. The user must accept the *aviso de privacidad*
    and the terms, and the server stores the document version + timestamp. The trial starts here (server clock).
- **US-A2** As a user, I log in and stay logged in.
  - AC: short-lived access token plus a rotating refresh token. Reusing a refresh token
    revokes the whole token family. Logout revokes. Login is rate-limited.
- **US-A3** I edit my user profile (name, notification preferences link).
- **US-A4** I delete my account from inside the app.
  - AC: required by the App Store and Google Play. It deletes all my taxpayers and records
    within the retention policy. The app does not cancel the store subscription itself; it
    explains how to cancel it.

### Taxpayers and onboarding (`module_taxpayers`)
- **US-T1** During onboarding I choose **persona física or persona moral first**, then
  enter the RFC and a display name, then pick one or more regimes.
  - AC: the regime list is filtered by person type. The RFC format is checked against the
    type (13 characters for física, 12 for moral) and against the check digit. At least one
    regime is required. An invalid combination is rejected with an explanation. Each regime
    asks for its options: 625 platform type(s) and withholding option, 606 deduction mode,
    601 coeficiente de utilidad.
- **US-T2** I add more taxpayers later from Settings, and switch the active taxpayer from any main screen.
  - AC: the same RFC cannot appear twice under my account. The switcher shows each
    taxpayer's display name, type and masked RFC. The selected taxpayer persists between sessions.
- **US-T3** I edit a taxpayer, add or remove regimes, or archive the taxpayer.
  - AC: removing a regime that has records needs confirmation and archives the regime
    with an end period; its records are kept. Archiving a taxpayer hides it but keeps its
    data; deleting it is a separate, irreversible action with confirmation.
- **US-T4** I enter tax inputs the estimate needs (year-to-date totals for prior months,
  coeficiente de utilidad, pending losses, provisional payments already made).
  - AC: optional at onboarding. When a strategy needs a missing input, the estimate is
    still shown, flagged "incomplete", with a call to action.

### Incomes (`module_incomes`)
- **US-I1** I record an income for the active taxpayer: regime, amount, IVA rate, dates and withholdings if any.
  - AC: amounts in MXN, stored as integer centavos. IVA rate from an allowed list (16%,
    8% border, 0%, exempt). Withholding fields appear only when the regime or customer type
    allows them. `collected_on` is required for cash-basis regimes. PM 601 requires
    `invoiced_on` and `collected_on` may stay empty (pending collection).
- **US-I2** I list a month's incomes, filter by regime, edit one, and delete one.
  - AC: paginated. Delete can be undone (Undo, per the house UI rules). Totals update immediately.

### Expenses (`module_expenses`)
- **US-E1** I record an expense with a category, regime, amount, IVA paid, payment method and date paid.
  - AC: categories come from a server catalog, and each category carries deductibility
    hints per regime. The app warns that cash payments above the legal threshold are not
    deductible. The mandatory legend sits next to the deductibility hints. RESICO PF shows
    that expenses do not reduce ISR (they may still matter for IVA).
- **US-E2** I list, filter, edit and delete expenses (same AC as US-I2).

### Tax engine (`module_tax_engine`)
- **US-X1** I see the estimated IVA and ISR for a month for the active taxpayer, per regime and in total.
  - AC: the mandatory legend is returned **as an API field** and shown on every estimate
    surface. Negative IVA shows as *saldo a favor*, not as a payment. Informational regimes
    show their explanation instead of a number.
- **US-X2** I open a breakdown that explains each step.
  - AC: each line shows the concept, the amount and the legal basis (article or rule). The
    parameter set (fiscal year, version) and the engine version are shown.
- **US-X3** Estimates are right for known cases.
  - AC: a golden suite of worked examples per full/partial regime, validated against
    official tables or examples and reviewed by a CPA, passes at 100% as a CI gate. Tariffs
    and rates are versioned data per fiscal year. Adding a regime means adding one strategy
    plus its parameters; no other module changes.

### Dashboard (`module_dashboard`)
- **US-D1** On Home I see, for the active taxpayer, this month's estimated taxes, income
  vs. expenses, and a 6–12 month trend chart.
  - AC: p95 under 1 s server-side. The mandatory legend is visible. The empty state points
    to the first record. A trial banner shows the days and records left.

### History (`module_history`)
- **US-H1** I browse past months for the active taxpayer and open any month's breakdown.
  - AC: each month is recalculated from the current records. It is marked when records
    changed after I last viewed it.

### Subscriptions (`module_subscriptions`)
- **US-S1** I subscribe weekly, monthly or yearly from the paywall.
  - AC: the purchase goes through the store sheet (RevenueCat SDK, app user ID = our
    user ID). The backend updates the entitlement from RevenueCat webhooks: authenticated,
    idempotent by event ID, safe when events arrive out of order. A post-purchase sync
    endpoint covers webhook delay. "Restore purchases" is available.
- **US-S2** In Settings I see my plan status.
  - AC: shows the trial time and records left, or the plan and its renewal date. A link
    opens the store's subscription management.

### Reminders (`module_reminders`)
- **US-R1** I get a push reminder before each monthly deadline.
  - AC: opt-in. Defaults: day 14 and the due date at 10:00 America/Mexico_City
    (configurable). The due date moves to the next business day when day 17 is a weekend or
    holiday. With several taxpayers, a single reminder lists them. Amounts appear only if
    "show amounts in notifications" is on (lock-screen privacy).

## 7. Trial and plan rules

| Rule | Recommendation |
|---|---|
| Unit of entitlement | The **user** (login), never the taxpayer. One plan covers all of the user's taxpayers |
| Trial start | At registration (server clock); no payment method |
| Trial end | **Whichever comes first**: 7 days, OR trying to create a 6th income or a 6th expense |
| Caps scope | **Per user, across all taxpayers** (5 incomes + 5 expenses in total). Per-taxpayer caps would let trial users add taxpayers to multiply free records |
| Cap counting | Records **created** (deleting does not free a slot) |
| Taxpayers in trial | Same fair-use cap as paid plans: 10 (Q4). Decided on P0-16 (2026-10-13): an uncapped trial let one account register unbounded RFCs |
| After trial, no plan | Read-only: existing data, dashboard, history and breakdown stay visible. Creating records shows the paywall |
| Paid plans | Weekly, monthly, yearly; one entitlement (`pro`); unlimited records and taxpayers (fair-use: see Q4) |
| Lapse / refund | Back to read-only; data is never deleted for non-payment |
| Re-registration | Deleting the account and signing up again with the same email does not grant a new trial (only a keyed hash of the email is kept) |
| Store intro offers | Not used in v1, to avoid two overlapping trials |
| Enforcement | Server-side on the create endpoints (error envelope, `402 entitlement_required`). The client paywall is UX only |
| Demo / comp | A server-side "comp" entitlement grants `pro` to demo and QA accounts (presentation mode) |

## 8. Legal and compliance

- **Disclaimer**: the mandatory legend appears on every estimate (an API field plus the UI).
  Onboarding and the terms state that the app is informational, not tax advice, and that the
  user stays responsible for their filings.
- **Privacy (LFPDPPP 2025)**: the Ley Federal de Protección de Datos Personales en Posesión
  de los Particulares was published in the DOF on 2025-03-20. Requirements: an integral and
  a simplified *aviso de privacidad* at signup; consent stored with its version; ARCO rights
  (access, rectification, cancellation, opposition) served in-app plus a documented email
  channel; a retention and deletion policy; a published list of processors (hosting,
  RevenueCat, Apple/Google, push provider). Legal review is required before launch.
- **Third-party data**: users may enter other people's RFC and financial data (household or
  accountant persona). The terms must state that the user has the right to process that
  data. Legal must confirm our role (responsable vs. encargado) (Q10).
- **Data minimization**:
  - What we collect: the RFC (personal data for físicas), encrypted at application level
    and with a keyed hash for uniqueness, always masked in the UI and in logs.
  - What we never collect: CURP, e.firma, CIEC or CFDI files.
  - Protection: financial records are encrypted in transit and at rest and protected by RLS
    per user. PII is never logged.
- **App stores**: in-app account deletion, restore purchases, subscription terms shown
  before purchase, links to the privacy notice and the terms.
- **Tax content accuracy**: every rule cites its source. Parameters are reviewed each
  fiscal year, when the new RMF and its Anexos come out (usually Dec–Jan). A yearly CPA
  review of the golden suite.

## 9. Non-functional requirements

- Monorepo: `backend/` is a Go monolith (Clean Architecture, GORM); `mobile/` is Flutter; `docs/`.
- OpenAPI spec-first (`docs/api/openapi.yaml`). The Dart client is generated from it. One error envelope everywhere.
- PostgreSQL with ENABLE + FORCE RLS on every user-scoped table, keyed on `user_id`. Child
  tables carry `user_id` too, so policies stay simple. The app role does not bypass RLS.
- Every module ships behind a flag `module_<name>`, default OFF; when disabled, its routes return 404 server-side.
- Money is stored as integer centavos. Rounding follows SAT rules and happens inside the engine only.
- Period assignment uses the America/Mexico_City time zone.
- Mobile release builds are obfuscated (`--obfuscate --split-debug-info`). Debug symbols and source maps are never shipped.
- Availability 99.5% in v1. p95 < 300 ms for record CRUD and < 1 s for the dashboard.

## 10. Success metrics

| Metric | Target (first 90 days after launch) |
|---|---|
| Activation: onboarding done + first estimate within 24 h | ≥ 60% of registrations |
| Trial → paid conversion | ≥ 5% |
| D30 retention of paid users | ≥ 70% |
| Golden suite correctness | 100% (CI gate) |
| Users opening the breakdown at least monthly | ≥ 30% (trust signal) |
| Crash-free sessions | ≥ 99.5% |
| Reminder → app open rate | ≥ 20% |
| Taxpayers per paying user (monitor; informs fair-use, Q4) | Track p50 / p95 |

## 11. Open questions for the owner

| # | Question | PM recommendation |
|---|---|---|
| Q1 | Trial ends at 7 days **or** at the 6th income/expense, whichever comes first? | Yes, whichever comes first; created records count |
| Q2 | Do trial caps (5+5) apply **per user** or **per taxpayer**? | **Per user, across all taxpayers** (prevents multiplying free records by adding RFCs) |
| Q3 | After the trial without paying: read-only or full lock? | Read-only |
| Q4 | Fair-use cap on taxpayers per paid user (no per-RFC pricing, but abuse protection)? | Soft cap of 10 taxpayers, raised on request; revisit with the metric in §10 |
| Q5 | Price points for weekly / monthly / yearly (MXN)? Is weekly worth it (store fees, churn)? | Owner to set. Consider monthly + yearly at launch and weekly later |
| Q6 | Which regimes get a **full estimate at launch** vs. later? | Must: 612, 626 PF, 601, 626 PM. Should: 625, 606. Rest: informational at launch (after research) |
| Q7 | For 606, 625 and 601, are the taxpayer's options (blind vs. actual deduction, definitive withholding, coeficiente) a per-regime setting? | Yes, per taxpayer + regime, defaulting to the most common option |
| Q8 | Do users enter prior months / prior-year inputs at onboarding (needed for cumulative ISR and the coeficiente)? | Optional at onboarding; an estimate missing them is flagged "incomplete" |
| Q9 | Is the RFC **required** for every taxpayer, or optional? (We never verify it against SAT) | Required, format + check digit validated (it drives type and uniqueness) |
| Q10 | Legal role when a user enters someone else's data (responsable vs. encargado), and who reviews the terms and the aviso? | The owner hires a lawyer before public launch |
| Q11 | Who validates the golden test cases? | An external CPA, once per fiscal year |
| Q12 | Push provider: FCM for both platforms (APNs via FCM)? | Yes (Architect to confirm in an ADR) |
| Q13 | Launch on both App Store and Google Play on day one? | Both |
| Q14 | Does the dashboard need a combined "all my taxpayers" view? | Not in v1: per active taxpayer plus a switcher; a combined view is Could |
