# ADR-0007 — A fiscal shared kernel in `internal/kernel/fiscal`

| Field | Value |
|---|---|
| Status | ACCEPTED (tech-lead review of PR #67, M3-03 / M4-04) |
| Date | 2026-10-13 |
| Amended | 2026-10-15 by M9-01: the due-date rule joins the kernel ([Amendment 1](#amendment-1-m9-01-2026-10-15-the-due-date-rule)); implemented by M9-04 |
| Related | [ADR-0001](0001-stack.md) Dependency Rule, [ADR-0002](0002-tax-engine.md) §3–§4, [MODULE-incomes](../project-specs/MODULE-incomes.md) §2.5, [MODULE-expenses](../project-specs/MODULE-expenses.md) §2, P0-29 FU4, `backend/internal/archtest` |
| Owner | tech-lead / architect |

## Context

PR #67 added the income (M3-03) and expense (M4-04) domains. Because modules never import
each other (archtest `cross-module`) and `platform/` takes no business logic (P0-29 FU4), the
PR copied the same code into both: `Date`, `CFDIUUID`, `PersonType`, `RegimeCode`, `IVARate`,
`ExpectedIVA`/`CheckIVA`/`IVATolerance`, the IVA withholding `Share`, suggestions, the
`RateKey` names and the whole error model. About 600 lines were duplicated, and the tests
were duplicated too.

These are not two models that happen to look alike. A record's date, its IVA treatment and
tolerance, the legal 2/3 share and the contract error codes must mean the same thing to
incomes, to expenses and to the M5 tax engine, which nets both sides into one IVA result
(ADR-0002 §1). If a copy drifts (for example a tolerance change in one module only), the
monthly estimate is wrong without anyone noticing. In DDD terms this is a **shared kernel**:
a small model that several bounded contexts agree to share and change together.

## Options

1. **Keep the copies.** This keeps the slices independent, but it violates DRY on knowledge
   that is the same by law, and every drift is a silent fiscal bug.
2. **Put it in `platform/`.** This breaks P0-29 FU4. Platform is cross-cutting technical
   infrastructure, not business rules.
3. **One module owns it and the others import it.** This breaks `cross-module` and couples
   the slices, so a module could no longer be removed on its own.
4. **A dedicated shared kernel, `internal/kernel/fiscal`, with its own archtest rule.
   Chosen.**

## Decision

1. `backend/internal/kernel/` is a new location kind in archtest (`kindKernel`). Every
   package of every module may import it, including `domain`.
2. Rule `kernel-imports`: a kernel package imports only the standard library minus
   frameworks, the pure platform packages (`money`) and other kernel packages. It never
   imports a module or impure platform, and the domain's external allowlist does not apply.
   Platform may not import the kernel (`platform-modules`). depguard (`kernel` list) and
   forbidigo (no float, no `time.Now`) cover `internal/kernel` the same way they cover a
   domain.
3. `internal/kernel/fiscal` holds **only** concepts that more than one bounded context
   reads:
   - `Date`, `MinRecordDate`, `CFDIUUID`;
   - `PersonType`, `RegimeCode` (the codes some module spells out), `IVARate`;
   - `RateCatalog` (`Rate` only), `RateKey`, and the keys read by more than one context
     (IVA rates, the PM fees/rent ISR withholding);
   - `ExpectedIVA`, `CheckIVA`, `IVATolerance`, `MaxCents`, `ApplyRate`;
   - the withholding kinds: a legal `Share` of the IVA, `SuggestAtRate`, and
     `Suggestion`/`Suggested`/`Withholdings`;
   - the error model (`Error`, `Class`, `Violation(s)`) and the shared contract codes.
4. Module-specific rules stay in the module. Incomes keeps the §2.6 applicability table, the
   platform keys and the income basis. Expenses keeps deductibility, the cash limit and the
   supplier rules.
5. **Extension without modification (OCP).** `Error` is a comparable `{code, class}`. A
   module declares its own codes as package variables with `fiscal.ValidationError` or
   `fiscal.BusinessRuleError`, so the kernel's `Class` never needs a switch over module
   codes. Module enums join `fiscal.ParseEnum` and `fiscal.CheckEnum` by implementing
   `Valid() bool`.
6. **Ports stay consumer-side.** Incomes uses `fiscal.RateCatalog` as is, since it needs
   exactly `Rate`. Expenses declares its own `RateCatalog`, which embeds `fiscal.RateCatalog`
   and adds `Amount` (the cash-payment limit). M5-08 implements both, and the composition
   root wires them.
7. Error strings now read `fiscal: <code>`. The code (`Detail()`) and the class are what
   the adapter maps, and both are unchanged.

## Consequences

- One copy of each fiscal rule, with 100 % coverage in the kernel. The research worked
  examples for IVA charged and IVA paid now sit in a single table.
- A change to the kernel is a change to every context. Kernel PRs need tech-lead review,
  and the M5 tax engine imports the same package instead of a third copy.
- The kernel must stay small. Before adding something, ask: "Does more than one bounded
  context read this, and must they agree?" If only one context reads it, it stays in that
  module.
- Modules are still vertical slices. Disabling or removing one module does not touch the
  kernel or the other modules.

## Amendment 1 (M9-01, 2026-10-15): the due-date rule

**Context.** The rule "day 17 of the following month, moved past weekends and holidays (CFF
art. 12 ¶1, ¶5)" lives in two copies inside `internal/taxengine`: `usecase/duedate.go` (the
monthly `due_date`, with the 621 bimester offset) and `usecase/annual/estimate.go`
`businessDay` (the annual deadlines). Both ask a `HolidayCalendar` port that the composition
root answers with `noHolidays` (the M5 "holiday gap", PLAN #98 FU2). M9 adds two more readers
in another bounded context: `getTaxCalendar` and the reminder job (MODULE-reminders §2.1).
The estimate shown on Home, the History row, the calendar entry and the push notification
must name the **same** date: a user told "vence el lunes 19" by one screen and "sábado 17" by
another stops trusting both. That passes this ADR's test: more than one bounded context reads
it, and they must agree.

**Options.**

1. *Reminders call the tax engine for dates.* Breaks `cross-module` (archtest): reminders would
   import taxengine, and disabling `module_tax_engine` would break the calendar, which the
   spec keeps working without it (MODULE-reminders §3).
2. *A third copy in reminders.* Every holiday or Friday-extension decision (OQ-R1) would have
   to land in three places; a drift is a wrong date on a lock screen.
3. *`platform/calendar`.* `platform/` takes no business rules (P0-29 FU4); CFF art. 12 is one.
4. ***Kernel functions, data through a value (chosen).***

**Decision.** Item 3 of the Decision gains the due-date rule, as pure functions with no I/O and
no `time.Now`:

- `HolidaySet`: an immutable value built from the `holidays` rows (`moves_due_date = true`) of
  the seeded years, with `Covers(year) bool`. Outside a covered year it answers weekends only;
  the caller decides whether that is acceptable (the calendar endpoint answers `422
  calendar_unavailable` for an unseeded `year`; the December entry and the reminder job fall
  back to weekends only and emit `reminders_calendar_unseeded`).
- `NextBusinessDay(d Date, h HolidaySet) (Date, MovedReason, error)`: the CFF 12 move, bounded
  (today's `maxDueShift` 14 → error). `MovedReason` is `weekend` | `holiday` | none, the
  contract's `TaxCalendarMovedReason`. Used by the monthly and the annual deadline.
- `MonthlyDueDate(p Period, bimonthly bool, h HolidaySet) (nominal, due Date, MovedReason,
  error)`: day 17 of the month after `p` (after the bimester when `bimonthly`), then
  `NextBusinessDay`.

The `holidays` table (M9-03) is read by an adapter in each consumer that builds a `HolidaySet`
(ports stay consumer-side, Decision item 6); the kernel never reads it. `noHolidays` and both
taxengine copies are deleted by M9-04, with an archtest/grep guard failing on `time.Saturday`
outside the kernel. Not in the kernel: the push sender port, the reminder job and the
`notification_log` (only reminders reads them; the FCM sender port is declared in the
reminders module, ADR-0001 Firebase Admin row).

**Complexity.** `HolidaySet` is a hash set of dates: membership O(1). `NextBusinessDay` scans
at most `maxDueShift` days, so a calendar of 12 entries is O(12 · 14) set lookups, constant in
practice; building the set is O(rows), about 10 per seeded year.

**Consequences.** A golden table pins the rule: the 24 periods 2025-12..2027-11 (every due date
in the seeded years, research §A.4.1) plus the annual deadlines of fiscal years 2025 and 2026.
No 2026–2027 date changes when `noHolidays` is replaced (no day 17 meets a holiday in those
years), so the estimate goldens and `EngineVersion` stay as they are; if any estimate golden
changes, M9-04 bumps `EngineVersion` per the engine version policy. The Friday extension
(OQ-R1) stays out until a CPA confirms; adding it is a change to this one function and every
reader follows. Contract 0.13.0 marks every calendar entry `subject_to_confirmation` meanwhile.
