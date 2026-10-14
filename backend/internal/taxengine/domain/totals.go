package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Totals are the taxpayer-level figures of one month (contract
// EstimateTotals). They carry no status.
type Totals struct {
	// IVA is the consolidated monthly IVA (ConsolidateIVA).
	IVA TaxResult
	// ISR is the sum of the payable per-regime ISR: a regime's favor
	// balance stays with that regime, never offset against another (OQ-T10).
	ISR TaxResult
	// WithholdingsToRemit is the ISR and IVA the taxpayer withheld from its
	// suppliers, summed over its regimes: remitted in full, never netted.
	WithholdingsToRemit money.Cents
	// TotalPayable is ISR payable + IVA payable + withholdings to remit.
	TotalPayable money.Cents
}

// NewTotals applies the totals rule (ADR-0002 §1, M5-22): each amount is
// counted exactly once. Each regime reports three disjoint things, its own
// ISR, its own IVA components (already netted into iva) and its
// withholdings to remit (ISR and IVA, already summed into WithheldToRemit),
// so the withholding lines are never re-added and WithheldToRemit never
// enters ISR or IVA. Fails with money.ErrOverflow.
func NewTotals(regimes []RegimeEstimate, iva TaxResult) (Totals, error) {
	var (
		s       summer
		isr     money.Cents
		toRemit money.Cents
	)
	for _, r := range regimes {
		if r.ISR != nil {
			isr = s.add(isr, r.ISR.Payable())
		}
		toRemit = s.add(toRemit, r.WithheldToRemit)
	}
	isrExact, err := isr.Exact()
	if s.err != nil || err != nil {
		return Totals{}, fmt.Errorf("totals: %w", firstError(s.err, err))
	}
	out := Totals{IVA: iva, ISR: NewTaxResult(isrExact), WithholdingsToRemit: toRemit}
	out.TotalPayable = s.add(s.add(out.ISR.Payable(), iva.Payable()), toRemit)
	if s.err != nil {
		return Totals{}, fmt.Errorf("totals: %w", s.err)
	}
	return out, nil
}

// firstError returns the first non-nil error.
func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
