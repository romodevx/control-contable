package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// IVATolerance is how far iva_cents may sit from round(amount × rate)
// (MODULE-incomes §2.5, OQ-3, shared by MODULE-expenses): CFDIs compute IVA
// per line, so totals drift a few centavos; the tolerance absorbs it without
// letting a typo through.
const IVATolerance money.Cents = 100

// MaxCents is the contract's NonNegativeCents maximum. Every amount below it
// keeps rate and share products far inside money.Exact's range.
const MaxCents money.Cents = 99_999_999_999_999

// ivaRateKeys maps each taxable treatment to its dated catalog rate.
var ivaRateKeys = map[IVARate]RateKey{
	IVARate16: KeyIVAGeneral,
	IVARate8:  KeyIVABorder,
	IVARate0:  KeyIVAZero,
}

// ExpectedIVA returns round(amount × rate) for treatment on day on. The
// rate must be in force that day (ErrRateNotValidOnDate): the border 8 % lapses
// with its decree, and no date is hard-coded. An exempt expense has no IVA
// and needs no rate.
func ExpectedIVA(amount money.Cents, treatment IVARate, on Date, catalog RateCatalog) (money.Cents, error) {
	if treatment == IVAExempt {
		return 0, nil
	}
	key, ok := ivaRateKeys[treatment]
	if !ok {
		return 0, ErrInvalid
	}
	r, ok := catalog.Rate(key, on)
	if !ok {
		return 0, ErrRateNotValidOnDate
	}
	return applyRate(amount, r)
}

// CheckIVA checks iva against the treatment on day on: a 0 % or exempt
// expense has no IVA at all; a taxable one must be within IVATolerance of
// ExpectedIVA (ErrIVAAmountMismatch). It is a check, not stored rounding: the
// stored IVA is what the user entered.
func CheckIVA(amount money.Cents, treatment IVARate, iva money.Cents, on Date, catalog RateCatalog) error {
	expected, err := ExpectedIVA(amount, treatment, on, catalog)
	if err != nil {
		return err
	}
	if !treatment.Taxable() {
		if iva != 0 {
			return ErrIVAAmountMismatch
		}
		return nil
	}
	if !withinTolerance(iva, expected) {
		return ErrIVAAmountMismatch
	}
	return nil
}

// withinTolerance reports |got − want| ≤ IVATolerance; a difference too large
// for int64 is outside it.
func withinTolerance(got, want money.Cents) bool {
	diff, err := got.Sub(want)
	return err == nil && diff <= IVATolerance && diff >= -IVATolerance
}

// applyRate returns base × r rounded once to the centavo, half away from
// zero. Only an absurd catalog rate can overflow, a parameter fault.
func applyRate(base money.Cents, r money.Rate) (money.Cents, error) {
	e, err := base.MulRate(r)
	if err != nil {
		return 0, fmt.Errorf("apply rate %s: %w: %w", r, ErrInvalidParameter, err)
	}
	return e.Round(), nil
}
