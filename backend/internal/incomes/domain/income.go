package domain

import (
	"errors"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Contract field names, as violations and warnings report them.
const (
	FieldAmountCents      = "amount_cents"
	FieldIVARate          = "iva_rate"
	FieldIVACents         = "iva_cents"
	FieldCounterpartyType = "counterparty_type"
	FieldISRWithheldCents = "isr_withheld_cents"
	FieldIVAWithheldCents = "iva_withheld_cents"
	FieldInvoicedOn       = "invoiced_on"
	FieldCollectedOn      = "collected_on"
	FieldPlatformActivity = "platform_activity"
	FieldCFDIUUID         = "cfdi_uuid"
)

// WarningCode is a non-blocking notice (contract Warning.code, MODULE-incomes
// §2.8). Codes are append-only.
type WarningCode string

// The income warnings. DuplicateCFDIUUID needs the stored incomes, so the
// use case adds it.
const (
	WarningWithholdingDiffersFromRate WarningCode = "withholding_differs_from_rate"
	WarningDuplicateCFDIUUID          WarningCode = "duplicate_cfdi_uuid"
)

// Warning is a notice on one field. It never echoes an amount, a description
// or a CFDI UUID.
type Warning struct {
	Code  WarningCode
	Field string
}

// Draft is an income as the user entered it, after the adapter decoded the
// request. Zero values are absent fields: a zero Date is null, an empty
// PlatformActivity lets a single 625 activity be the default.
type Draft struct {
	Amount           money.Cents // subtotal before IVA, MXN
	IVARate          IVARate
	IVA              money.Cents
	Counterparty     Counterparty
	Withheld         Withholdings
	Dates            Dates
	PlatformActivity PlatformActivity
}

// Totals are the live summary of an income ("IVA · Retenciones · Neto").
type Totals struct {
	Subtotal money.Cents // amount before IVA
	IVA      money.Cents
	Total    money.Cents // Subtotal + IVA
	Withheld money.Cents // ISR + IVA withheld
	Net      money.Cents // Total − Withheld: "Neto a recibir"
}

// Assessment is a valid income's derived facts: what the use case stores
// besides the draft (booked_on, the resolved 625 activity), what the
// response shows (pending_collection, warnings) and the client's prefills.
type Assessment struct {
	BookedOn          Date
	PendingCollection bool
	PlatformActivity  PlatformActivity
	Suggested         Suggested
	Totals            Totals
	Warnings          []Warning
}

// Assess applies MODULE-incomes §2 to a draft for payee, with today the
// current Mexico City day (Clock port) and catalog the dated rates.
//
// Validation (validation_failed) runs first and reports every field; only a
// draft that passes it meets the business rules (business_rule_violation),
// which are also all reported. Both come back as Violations. A parameter
// fault is ErrInvalidParameter. The regime window (CheckRegime), the
// idempotency key and the trial cap belong to the use case.
func Assess(d Draft, p Payee, today Date, catalog RateCatalog) (Assessment, error) {
	if err := checkContext(p, today, catalog); err != nil {
		return Assessment{}, err
	}
	activity, vs := validate(d, p, today)
	if len(vs) > 0 {
		return Assessment{}, vs
	}
	return applyRules(d, p, activity, catalog)
}

// checkContext rejects server-side inputs the rules cannot use.
func checkContext(p Payee, today Date, catalog RateCatalog) error {
	switch {
	case !slices.Contains(bases, p.Basis):
		return fmt.Errorf("income basis %q: %w", p.Basis, ErrInvalidParameter)
	case !slices.Contains(personTypes, p.PersonType):
		return fmt.Errorf("person type %q: %w", p.PersonType, ErrInvalidParameter)
	case today.IsZero():
		return fmt.Errorf("no current date: %w", ErrInvalidParameter)
	case catalog == nil:
		return fmt.Errorf("no rate catalog: %w", ErrInvalidParameter)
	}
	return nil
}

// validate is the schema-level pass. It returns the effective 625 activity.
func validate(d Draft, p Payee, today Date) (PlatformActivity, Violations) {
	var vs Violations
	checkAmount(&vs, d.Amount)
	checkEnum(&vs, FieldIVARate, d.IVARate, ivaRates)
	checkBounds(&vs, FieldIVACents, d.IVA)
	checkEnum(&vs, FieldCounterpartyType, d.Counterparty, counterparties)
	checkBounds(&vs, FieldISRWithheldCents, d.Withheld.ISR)
	checkBounds(&vs, FieldIVAWithheldCents, d.Withheld.IVA)
	checkDate(&vs, FieldInvoicedOn, d.Dates.InvoicedOn, p.Basis, today)
	checkDate(&vs, FieldCollectedOn, d.Dates.CollectedOn, p.Basis, today)
	activity := resolveActivity(&vs, d.PlatformActivity, p)
	return activity, vs
}

func checkAmount(vs *Violations, amount money.Cents) {
	switch {
	case amount < 1:
		vs.add(FieldAmountCents, ErrMustBePositive)
	case amount > MaxCents:
		vs.add(FieldAmountCents, ErrOutOfRange)
	}
}

func checkBounds(vs *Violations, field string, c money.Cents) {
	if c < 0 || c > MaxCents {
		vs.add(field, ErrOutOfRange)
	}
}

// checkDate requires the basis's date and keeps every date within
// [MinRecordDate, today]: future dates are refused.
func checkDate(vs *Violations, field string, on Date, basis IncomeBasis, today Date) {
	switch {
	case on.IsZero() && field == basis.RequiredDateField():
		vs.add(field, ErrRequired)
	case on.IsZero():
	case on.Before(MinRecordDate) || on.After(today):
		vs.add(field, ErrOutOfRange)
	}
}

// resolveActivity: platform_activity belongs to 625 only, defaults to the
// taxpayer's activity when they hold exactly one, else is required.
func resolveActivity(vs *Violations, activity PlatformActivity, p Payee) PlatformActivity {
	if p.Regime != RegimePlatforms {
		if activity != "" {
			vs.add(FieldPlatformActivity, ErrNotForRegime)
		}
		return ""
	}
	if activity == "" && len(p.PlatformActivities) == 1 {
		return p.PlatformActivities[0]
	}
	checkEnum(vs, FieldPlatformActivity, activity, activities)
	return activity
}

// applyRules is the business pass over a validated draft.
func applyRules(d Draft, p Payee, activity PlatformActivity, catalog RateCatalog) (Assessment, error) {
	var vs Violations
	if err := checkIVAInto(&vs, d, catalog); err != nil {
		return Assessment{}, err
	}
	if activity != "" && !slices.Contains(p.PlatformActivities, activity) {
		vs.add(FieldPlatformActivity, ErrActivityNotHeld)
	}
	policy := PolicyFor(p, d.Counterparty, activity)
	checkWithheld(&vs, FieldISRWithheldCents, d.Withheld.ISR, d.Amount, policy.ISR.Allowed)
	checkWithheld(&vs, FieldIVAWithheldCents, d.Withheld.IVA, d.IVA, policy.IVA.Allowed)
	if len(vs) > 0 {
		return Assessment{}, vs
	}
	booked := p.Basis.BookingDate(d.Dates)
	suggested, err := policy.Suggest(d.Amount, d.IVA, booked, catalog)
	if err != nil {
		return Assessment{}, err
	}
	return Assessment{
		BookedOn:          booked,
		PendingCollection: p.Basis.PendingCollection(d.Dates),
		PlatformActivity:  activity,
		Suggested:         suggested,
		Totals:            totalsOf(d),
		Warnings:          withholdingWarnings(suggested, d.Withheld),
	}, nil
}

// checkIVAInto records the IVA rule's violation on its field; a parameter
// fault is returned instead.
func checkIVAInto(vs *Violations, d Draft, catalog RateCatalog) error {
	err := CheckIVA(d.Amount, d.IVARate, d.IVA, d.Dates.RateDate(), catalog)
	switch {
	case errors.Is(err, ErrRateNotValidOnDate):
		vs.add(FieldIVARate, ErrRateNotValidOnDate)
	case errors.Is(err, ErrIVAAmountMismatch):
		vs.add(FieldIVACents, ErrIVAAmountMismatch)
	case err != nil:
		return err
	}
	return nil
}

// checkWithheld: a withholding the policy does not allow must be zero, and
// one it allows never exceeds its base.
func checkWithheld(vs *Violations, field string, withheld, base money.Cents, allowed bool) {
	switch {
	case withheld > 0 && !allowed:
		vs.add(field, ErrWithholdingNotApplicable)
	case withheld > base:
		vs.add(field, ErrWithholdingExceedsBase)
	}
}

func withholdingWarnings(s Suggested, w Withholdings) []Warning {
	var warnings []Warning
	if s.ISR.differs(w.ISR) {
		warnings = append(warnings, Warning{Code: WarningWithholdingDiffersFromRate, Field: FieldISRWithheldCents})
	}
	if s.IVA.differs(w.IVA) {
		warnings = append(warnings, Warning{Code: WarningWithholdingDiffersFromRate, Field: FieldIVAWithheldCents})
	}
	return warnings
}

// totalsOf sums a validated draft: every amount is within [0, MaxCents] and
// the withholdings within their bases, so plain int64 arithmetic is exact.
func totalsOf(d Draft) Totals {
	total := d.Amount + d.IVA
	withheld := d.Withheld.ISR + d.Withheld.IVA
	return Totals{Subtotal: d.Amount, IVA: d.IVA, Total: total, Withheld: withheld, Net: total - withheld}
}
