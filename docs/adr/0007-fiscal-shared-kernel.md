# ADR-0007 — A fiscal shared kernel in `internal/kernel/fiscal`

| Field | Value |
|---|---|
| Status | ACCEPTED (tech-lead review of PR #67, M3-03 / M4-04) |
| Date | 2026-10-13 |
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
