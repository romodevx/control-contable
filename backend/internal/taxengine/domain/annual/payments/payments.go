// Package payments resolves the provisional ISR payments an annual section
// credits (MODULE-annual §2.5, PLAN ANN-10): per month, what the user
// confirmed as filed, else the M5 tax input prior_provisional_payments,
// else the app's own monthly estimate (max(0, isr)). The three sources come
// through the Sources input port, so the annual strategies stay pure and the
// use case (ANN-14) fills it from annual_filed_payments, the tax inputs and
// one Estimator range call.
//
// Money is int64 centavos; there is no float and no clock in this package.
package payments

import (
	"fmt"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Error is a payments error; its values are constants matched with
// errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// ErrInvalidPayments reports months outside 1..12, repeated or unordered,
// or a negative filed or typed payment.
const ErrInvalidPayments Error = "annual payments: invalid provisional payments"

// Source is where a month's payment comes from (contract source of a
// payment line).
type Source string

// The sources, in precedence order.
const (
	SourceFiled          Source = "filed"
	SourceTaxInput       Source = "tax_input"
	SourceEstimatedByApp Source = "estimated_by_app"
)

// precedence lists the sources in the order Resolve tries them.
var precedence = []Source{SourceFiled, SourceTaxInput, SourceEstimatedByApp}

// Valid reports a spec payment source.
func (s Source) Valid() bool { return slices.Contains(precedence, s) }

// AppEstimate is the app's monthly ISR estimate of a regime for a month;
// Incomplete when that monthly estimate was itself incomplete.
type AppEstimate struct {
	ISR        money.Cents
	Incomplete bool
}

// Sources is the input port of one regime's provisional payments of a
// fiscal year: each source's amount for a month, when it has one.
type Sources interface {
	Filed(m time.Month) (money.Cents, bool)
	TaxInput(m time.Month) (money.Cents, bool)
	AppEstimate(m time.Month) (AppEstimate, bool)
}

// Payment is one credited provisional payment. Month 0 is a year total as
// a source states it (a hand-authored golden, a CPA figure).
type Payment struct {
	Month  time.Month
	Amount money.Cents
	Source Source
}

// Resolved is a regime's payments of the year and the inputs that would
// make them verifiable: filed_payment:<regime>:<YYYY-MM> for each month
// whose payment is a default from an incomplete monthly estimate, or that
// has no source at all (MODULE-annual §2.6; the section is incomplete).
type Resolved struct {
	Payments []Payment
	Missing  []string
}

// Resolve applies the precedence to each of months (the regime's active
// months of fiscalYear, ascending). O(len(months)).
func Resolve(regime fiscal.RegimeCode, fiscalYear int, months []time.Month, src Sources) (Resolved, error) {
	if err := validMonths(months); err != nil {
		return Resolved{}, err
	}
	var out Resolved
	for _, m := range months {
		p, complete, err := resolveMonth(m, src)
		if err != nil {
			return Resolved{}, err
		}
		if p.Source != "" {
			out.Payments = append(out.Payments, p)
		}
		if !complete {
			out.Missing = append(out.Missing, fmt.Sprintf("filed_payment:%s:%04d-%02d", regime, fiscalYear, int(m)))
		}
	}
	return out, nil
}

func validMonths(months []time.Month) error {
	for i, m := range months {
		if m < time.January || m > time.December {
			return fmt.Errorf("month %d: %w", int(m), ErrInvalidPayments)
		}
		if i > 0 && m <= months[i-1] {
			return fmt.Errorf("months must be ascending and unique (%d after %d): %w", int(m), int(months[i-1]), ErrInvalidPayments)
		}
	}
	return nil
}

// resolveMonth returns month m's payment (no Source when there is none) and
// whether it is verifiable.
func resolveMonth(m time.Month, src Sources) (Payment, bool, error) {
	if v, ok := src.Filed(m); ok {
		return typed(m, v, SourceFiled)
	}
	if v, ok := src.TaxInput(m); ok {
		return typed(m, v, SourceTaxInput)
	}
	if e, ok := src.AppEstimate(m); ok {
		return Payment{Month: m, Amount: max(e.ISR, 0), Source: SourceEstimatedByApp}, !e.Incomplete, nil
	}
	return Payment{}, false, nil
}

func typed(m time.Month, v money.Cents, s Source) (Payment, bool, error) {
	if v < 0 {
		return Payment{}, false, fmt.Errorf("%s payment of month %d is negative: %w", s, int(m), ErrInvalidPayments)
	}
	return Payment{Month: m, Amount: v, Source: s}, true, nil
}

// Total is the sum of the payments of one source and how many there are.
type Total struct {
	Source Source
	Amount money.Cents
	Months int
}

// Totals sums the payments per source, in precedence order, omitting a
// source with no payment; money.ErrOverflow past int64.
func (r Resolved) Totals() ([]Total, error) {
	var out []Total
	for _, s := range precedence {
		t := Total{Source: s}
		for _, p := range r.Payments {
			if p.Source != s {
				continue
			}
			sum, err := t.Amount.Add(p.Amount)
			if err != nil {
				return nil, fmt.Errorf("annual payments %s: %w", s, err)
			}
			t.Amount, t.Months = sum, t.Months+1
		}
		if t.Months > 0 {
			out = append(out, t)
		}
	}
	return out, nil
}

// EstimatedMonths is the number of payments taken from the app's estimate
// (the payments_estimated disclaimer's {n}).
func (r Resolved) EstimatedMonths() int {
	n := 0
	for _, p := range r.Payments {
		if p.Source == SourceEstimatedByApp {
			n++
		}
	}
	return n
}
