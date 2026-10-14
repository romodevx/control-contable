package domain

import (
	"errors"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
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

// The income warnings. DuplicateCFDIUUID needs the stored incomes and
// ProvisionalParameters the parameter set a date reads, so the use case
// adds them.
const (
	WarningWithholdingDiffersFromRate WarningCode = "withholding_differs_from_rate"
	WarningDuplicateCFDIUUID          WarningCode = "duplicate_cfdi_uuid"
	// WarningProvisionalParameters: the rates checked on a date of the
	// income come from a provisional parameter set (the year is not
	// published yet, or the set is marked provisional).
	WarningProvisionalParameters WarningCode = "provisional_parameters"
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
	IVARate          fiscal.IVARate
	IVA              money.Cents
	Counterparty     Counterparty
	Withheld         fiscal.Withholdings
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
	BookedOn          fiscal.Date
	PendingCollection bool
	PlatformActivity  PlatformActivity
	Suggested         fiscal.Suggested
	Totals            Totals
	Warnings          []Warning
}

// Assess applies MODULE-incomes §2 to a draft for payee, with today the
// current Mexico City day (Clock port) and catalog the dated rates.
//
// Validation (validation_failed) runs first and reports every field; only a
// draft that passes it meets the business rules (business_rule_violation),
// which are also all reported. Both come back as fiscal.Violations. A
// parameter fault is fiscal.ErrInvalidParameter. The regime window
// (CheckRegime), the idempotency key and the trial cap belong to the use
// case.
func Assess(d Draft, p Payee, today fiscal.Date, catalog fiscal.RateCatalog) (Assessment, error) {
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
func checkContext(p Payee, today fiscal.Date, catalog fiscal.RateCatalog) error {
	switch {
	case !p.Basis.Valid():
		return fmt.Errorf("income basis %q: %w", p.Basis, fiscal.ErrInvalidParameter)
	case !p.PersonType.Valid():
		return fmt.Errorf("person type %q: %w", p.PersonType, fiscal.ErrInvalidParameter)
	case today.IsZero():
		return fmt.Errorf("no current date: %w", fiscal.ErrInvalidParameter)
	case catalog == nil:
		return fmt.Errorf("no rate catalog: %w", fiscal.ErrInvalidParameter)
	}
	return nil
}

// validate is the schema-level pass. It returns the effective 625 activity.
func validate(d Draft, p Payee, today fiscal.Date) (PlatformActivity, fiscal.Violations) {
	var vs fiscal.Violations
	checkAmount(&vs, d.Amount)
	fiscal.CheckEnum(&vs, FieldIVARate, d.IVARate)
	checkBounds(&vs, FieldIVACents, d.IVA)
	fiscal.CheckEnum(&vs, FieldCounterpartyType, d.Counterparty)
	checkBounds(&vs, FieldISRWithheldCents, d.Withheld.ISR)
	checkBounds(&vs, FieldIVAWithheldCents, d.Withheld.IVA)
	checkDate(&vs, FieldInvoicedOn, d.Dates.InvoicedOn, p.Basis, today)
	checkDate(&vs, FieldCollectedOn, d.Dates.CollectedOn, p.Basis, today)
	activity := resolveActivity(&vs, d.PlatformActivity, p)
	return activity, vs
}

func checkAmount(vs *fiscal.Violations, amount money.Cents) {
	switch {
	case amount < 1:
		vs.Add(FieldAmountCents, fiscal.ErrMustBePositive)
	case amount > fiscal.MaxCents:
		vs.Add(FieldAmountCents, fiscal.ErrOutOfRange)
	}
}

func checkBounds(vs *fiscal.Violations, field string, c money.Cents) {
	if c < 0 || c > fiscal.MaxCents {
		vs.Add(field, fiscal.ErrOutOfRange)
	}
}

// checkDate requires the basis's date and keeps every date within
// [fiscal.MinRecordDate, today]: future dates are refused.
func checkDate(vs *fiscal.Violations, field string, on fiscal.Date, basis IncomeBasis, today fiscal.Date) {
	switch {
	case on.IsZero() && field == basis.RequiredDateField():
		vs.Add(field, fiscal.ErrRequired)
	case on.IsZero():
	case on.Before(fiscal.MinRecordDate) || on.After(today):
		vs.Add(field, fiscal.ErrOutOfRange)
	}
}

// resolveActivity: platform_activity belongs to 625 only, defaults to the
// taxpayer's activity when they hold exactly one, else is required.
func resolveActivity(vs *fiscal.Violations, activity PlatformActivity, p Payee) PlatformActivity {
	if p.Regime != fiscal.RegimePlatforms {
		if activity != "" {
			vs.Add(FieldPlatformActivity, ErrNotForRegime)
		}
		return ""
	}
	if activity == "" && len(p.PlatformActivities) == 1 {
		return p.PlatformActivities[0]
	}
	fiscal.CheckEnum(vs, FieldPlatformActivity, activity)
	return activity
}

// applyRules is the business pass over a validated draft.
func applyRules(d Draft, p Payee, activity PlatformActivity, catalog fiscal.RateCatalog) (Assessment, error) {
	var vs fiscal.Violations
	if err := checkIVAInto(&vs, d, catalog); err != nil {
		return Assessment{}, err
	}
	if activity != "" && !slices.Contains(p.PlatformActivities, activity) {
		vs.Add(FieldPlatformActivity, ErrActivityNotHeld)
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
func checkIVAInto(vs *fiscal.Violations, d Draft, catalog fiscal.RateCatalog) error {
	err := fiscal.CheckIVA(d.Amount, d.IVARate, d.IVA, d.Dates.RateDate(), catalog)
	switch {
	case errors.Is(err, fiscal.ErrRateNotValidOnDate):
		vs.Add(FieldIVARate, fiscal.ErrRateNotValidOnDate)
	case errors.Is(err, fiscal.ErrIVAAmountMismatch):
		vs.Add(FieldIVACents, fiscal.ErrIVAAmountMismatch)
	case err != nil:
		return err
	}
	return nil
}

// checkWithheld: a withholding the policy does not allow must be zero, and
// one it allows never exceeds its base.
func checkWithheld(vs *fiscal.Violations, field string, withheld, base money.Cents, allowed bool) {
	switch {
	case withheld > 0 && !allowed:
		vs.Add(field, fiscal.ErrWithholdingNotApplicable)
	case withheld > base:
		vs.Add(field, fiscal.ErrWithholdingExceedsBase)
	}
}

func withholdingWarnings(s fiscal.Suggested, w fiscal.Withholdings) []Warning {
	var warnings []Warning
	if s.ISR.Differs(w.ISR) {
		warnings = append(warnings, Warning{Code: WarningWithholdingDiffersFromRate, Field: FieldISRWithheldCents})
	}
	if s.IVA.Differs(w.IVA) {
		warnings = append(warnings, Warning{Code: WarningWithholdingDiffersFromRate, Field: FieldIVAWithheldCents})
	}
	return warnings
}

// totalsOf sums a validated draft: every amount is within [0, fiscal.MaxCents] and
// the withholdings within their bases, so plain int64 arithmetic is exact.
func totalsOf(d Draft) Totals {
	total := d.Amount + d.IVA
	withheld := d.Withheld.ISR + d.Withheld.IVA
	return Totals{Subtotal: d.Amount, IVA: d.IVA, Total: total, Withheld: withheld, Net: total - withheld}
}
