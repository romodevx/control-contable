package domain

import (
	"errors"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Contract field names, as violations and warnings report them.
const (
	FieldCategoryCode     = "category_code"
	FieldAmountCents      = "amount_cents"
	FieldIVARate          = "iva_rate"
	FieldIVACents         = "iva_cents"
	FieldPaymentMethod    = "payment_method"
	FieldPaidOn           = "paid_on"
	FieldHasCFDI          = "has_cfdi"
	FieldCFDIUUID         = "cfdi_uuid"
	FieldDeductible       = "deductible"
	FieldISRWithheldCents = "isr_withheld_cents"
	FieldIVAWithheldCents = "iva_withheld_cents"
)

// WarningCode is a non-blocking notice (contract Warning.code,
// MODULE-expenses §2.8). Codes are append-only.
type WarningCode string

// The expense warnings. DuplicateCFDIUUID needs the stored expenses, so the
// use case adds it.
const (
	WarningCashOverThreshold        WarningCode = "cash_over_threshold"
	WarningNoCFDI                   WarningCode = "no_cfdi"
	WarningNotISRDeductibleInRegime WarningCode = "not_isr_deductible_in_regime"
	WarningCategoryNotDeductible    WarningCode = "category_not_deductible"
	WarningDuplicateCFDIUUID        WarningCode = "duplicate_cfdi_uuid"
)

// Warning is a notice on one field. It never echoes an amount, a description
// or a CFDI UUID.
type Warning struct {
	Code  WarningCode
	Field string
}

// Draft is an expense as the user entered it, after the adapter decoded the
// request. A zero Date or CFDIUUID is an absent field.
type Draft struct {
	Category      CategoryCode
	Amount        money.Cents // subtotal before IVA, MXN
	IVARate       IVARate
	IVA           money.Cents
	PaymentMethod PaymentMethod
	PaidOn        Date // the booking date: every regime deducts on a paid basis
	HasCFDI       bool
	CFDIUUID      CFDIUUID
	Deductible    Override
	Withheld      Withholdings
}

// Totals are the live summary of an expense.
type Totals struct {
	Subtotal money.Cents // amount before IVA
	IVA      money.Cents
	Total    money.Cents // Subtotal + IVA
	Withheld money.Cents // ISR + IVA withheld from the supplier
	Net      money.Cents // Total − Withheld: paid to the supplier
}

// Assessment is a valid expense's derived facts: the stored effective
// deductible and its source, the hint the form shows, the creditable-IVA
// candidate M4-08 aggregates, the client's prefills, totals and warnings.
type Assessment struct {
	BookedOn      Date
	Hint          Hint
	Deductible    bool
	Source        DeductibleSource
	CashOverLimit bool
	// CreditableIVA is the IVA that may be credited (LIVA 5, ⚠️ CPA): with a
	// CFDI, and deductible for ISR — or, in RESICO PF, where ISR deductions
	// do not exist, deductible but for the regime (not a cash payment over
	// the limit). M5 applies the final rule.
	CreditableIVA money.Cents
	Suggested     Suggested
	Totals        Totals
	Warnings      []Warning
}

// Assess applies MODULE-expenses §2 to a draft paid by p, under the
// category's rule (category_regime_rules for p's regime and person type),
// with today the current Mexico City day (Clock port) and catalog the dated
// parameters.
//
// Validation (validation_failed) runs first and reports every field; only a
// draft that passes it meets the business rules (business_rule_violation).
// Both come back as Violations. A parameter fault is ErrInvalidParameter.
// The regime window, the category's existence, the idempotency key and the
// trial cap belong to the use case.
func Assess(d Draft, p Payer, rule CategoryRule, today Date, catalog RateCatalog) (Assessment, error) {
	if err := checkContext(d, p, rule, today, catalog); err != nil {
		return Assessment{}, err
	}
	if vs := validate(d, today); len(vs) > 0 {
		return Assessment{}, vs
	}
	return applyRules(d, p, rule, catalog)
}

// checkContext rejects server-side inputs the rules cannot use.
func checkContext(d Draft, p Payer, rule CategoryRule, today Date, catalog RateCatalog) error {
	switch {
	case !slices.Contains(personTypes, p.PersonType):
		return fmt.Errorf("person type %q: %w", p.PersonType, ErrInvalidParameter)
	case !slices.Contains(deductibilities, rule.Deductibility):
		return fmt.Errorf("category rule deductibility %q: %w", rule.Deductibility, ErrInvalidParameter)
	case rule.Category != d.Category:
		return fmt.Errorf("rule of category %q for %q: %w", rule.Category, d.Category, ErrInvalidParameter)
	case today.IsZero():
		return fmt.Errorf("no current date: %w", ErrInvalidParameter)
	case catalog == nil:
		return fmt.Errorf("no rate catalog: %w", ErrInvalidParameter)
	}
	return nil
}

// validate is the schema-level pass.
func validate(d Draft, today Date) Violations {
	var vs Violations
	if d.Category == "" {
		vs.add(FieldCategoryCode, ErrRequired)
	}
	checkAmount(&vs, d.Amount)
	checkEnum(&vs, FieldIVARate, d.IVARate, ivaRates)
	checkBounds(&vs, FieldIVACents, d.IVA)
	checkEnum(&vs, FieldPaymentMethod, d.PaymentMethod, paymentMethods)
	checkPaidOn(&vs, d.PaidOn, today)
	if !d.CFDIUUID.IsZero() && !d.HasCFDI {
		vs.add(FieldCFDIUUID, ErrHasCFDIMismatch)
	}
	if !slices.Contains(overrides, d.Deductible) {
		vs.add(FieldDeductible, ErrInvalid)
	}
	checkBounds(&vs, FieldISRWithheldCents, d.Withheld.ISR)
	checkBounds(&vs, FieldIVAWithheldCents, d.Withheld.IVA)
	return vs
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

// checkPaidOn requires paid_on within [MinRecordDate, today].
func checkPaidOn(vs *Violations, on, today Date) {
	switch {
	case on.IsZero():
		vs.add(FieldPaidOn, ErrRequired)
	case on.Before(MinRecordDate) || on.After(today):
		vs.add(FieldPaidOn, ErrOutOfRange)
	}
}

// applyRules is the business pass over a validated draft.
func applyRules(d Draft, p Payer, rule CategoryRule, catalog RateCatalog) (Assessment, error) {
	var vs Violations
	if err := checkIVAInto(&vs, d, catalog); err != nil {
		return Assessment{}, err
	}
	allowed := MayWithhold(p)
	checkWithheld(&vs, FieldISRWithheldCents, d.Withheld.ISR, d.Amount, allowed)
	checkWithheld(&vs, FieldIVAWithheldCents, d.Withheld.IVA, d.IVA, allowed)
	if len(vs) > 0 {
		return Assessment{}, vs
	}
	cashOver, err := cashOverLimit(d, rule, catalog)
	if err != nil {
		return Assessment{}, err
	}
	suggested, err := SuggestWithholdings(d.Category, p, d.Amount, d.IVA, d.PaidOn, catalog)
	if err != nil {
		return Assessment{}, err
	}
	hint := ResolveDeductibility(rule, p)
	deductible, source := effectiveDeductible(hint, d, cashOver)
	return Assessment{
		BookedOn:      d.PaidOn,
		Hint:          hint,
		Deductible:    deductible,
		Source:        source,
		CashOverLimit: cashOver,
		CreditableIVA: creditableIVA(d, p, deductible, cashOver),
		Suggested:     suggested,
		Totals:        totalsOf(d),
		Warnings:      warningsOf(hint, d, cashOver),
	}, nil
}

// checkIVAInto records the IVA rule's violation on its field (on paid_on);
// a parameter fault is returned instead.
func checkIVAInto(vs *Violations, d Draft, catalog RateCatalog) error {
	err := CheckIVA(d.Amount, d.IVARate, d.IVA, d.PaidOn, catalog)
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

// checkWithheld: a física withholds nothing (LIVA 1-A); a moral never above
// the base.
func checkWithheld(vs *Violations, field string, withheld, base money.Cents, allowed bool) {
	switch {
	case withheld > 0 && !allowed:
		vs.add(field, ErrWithholdingNotApplicable)
	case withheld > base:
		vs.add(field, ErrWithholdingExceedsBase)
	}
}

// cashOverLimit is MODULE-expenses §2.5 (⚠️ CPA): a cash payment whose total
// with IVA exceeds the dated limit, or any cash payment of a category that
// requires an electronic payment.
func cashOverLimit(d Draft, rule CategoryRule, catalog RateCatalog) (bool, error) {
	if d.PaymentMethod != PaymentCash {
		return false, nil
	}
	if rule.ElectronicPaymentOnly {
		return true, nil
	}
	limit, ok := catalog.Amount(KeyCashPaymentLimit, d.PaidOn)
	if !ok {
		return false, fmt.Errorf("%s on %s: %w", KeyCashPaymentLimit, d.PaidOn, ErrInvalidParameter)
	}
	return d.Amount+d.IVA > limit, nil
}

// effectiveDeductible is MODULE-expenses §2.4: the user's override wins;
// otherwise the hint's default, given a CFDI and no cash payment over the
// limit.
func effectiveDeductible(hint Hint, d Draft, cashOver bool) (bool, DeductibleSource) {
	switch d.Deductible {
	case ClaimDeductible:
		return true, SourceUser
	case ClaimNotDeductible:
		return false, SourceUser
	}
	return hint.DefaultDeductible && d.HasCFDI && !cashOver, SourceCategoryDefault
}

// creditableIVA is the IVA candidate for crediting (see Assessment).
func creditableIVA(d Draft, p Payer, deductible, cashOver bool) money.Cents {
	resicoPF := p.Regime == RegimeRESICO && p.PersonType == PersonFisica
	if d.HasCFDI && (deductible || (resicoPF && !cashOver)) {
		return d.IVA
	}
	return 0
}

func warningsOf(hint Hint, d Draft, cashOver bool) []Warning {
	var warnings []Warning
	if cashOver {
		warnings = append(warnings, Warning{Code: WarningCashOverThreshold, Field: FieldPaymentMethod})
	}
	if !d.HasCFDI {
		warnings = append(warnings, Warning{Code: WarningNoCFDI, Field: FieldHasCFDI})
	}
	if d.Deductible != ClaimDeductible {
		return warnings
	}
	switch hint.Deductibility {
	case NotDeductible:
		warnings = append(warnings, Warning{Code: WarningCategoryNotDeductible, Field: FieldDeductible})
	case ISRNotApplicable:
		warnings = append(warnings, Warning{Code: WarningNotISRDeductibleInRegime, Field: FieldDeductible})
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
