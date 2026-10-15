package annual

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// AppEstimate is the app's monthly ISR estimate of a regime for a month;
// Incomplete when that monthly estimate was itself incomplete.
type AppEstimate struct {
	ISR        money.Cents
	Incomplete bool
}

// PaymentSources is the input port of one regime's provisional payments of
// a fiscal year: each source's amount for a month, when it has one. The use
// case (ANN-14) fills it from annual_filed_payments, the M5 tax inputs and
// one Estimator range call.
type PaymentSources interface {
	Filed(m time.Month) (money.Cents, bool)
	TaxInput(m time.Month) (money.Cents, bool)
	AppEstimate(m time.Month) (AppEstimate, bool)
}

// ResolvedPayments are a regime's payments of the year, one per month by
// precedence, and the inputs that would make them verifiable:
// filed_payment:<regime>:<YYYY-MM> for each month whose payment is a default
// from an incomplete monthly estimate, or that has no source at all
// (MODULE-annual §2.6; the section is incomplete).
type ResolvedPayments struct {
	Payments []Payment
	Missing  []string
}

// ResolvePayments applies the MODULE-annual §2.5 precedence (filed, then
// the tax input, then the app's estimate as max(0, isr)) to each of months:
// the regime's active months of fiscalYear, ascending. O(len(months)).
func ResolvePayments(regime fiscal.RegimeCode, fiscalYear int, months []time.Month, src PaymentSources) (ResolvedPayments, error) {
	if err := ascendingMonths(months); err != nil {
		return ResolvedPayments{}, err
	}
	var out ResolvedPayments
	for _, m := range months {
		p, verifiable, err := resolveMonth(m, src)
		if err != nil {
			return ResolvedPayments{}, err
		}
		if p.Source != "" {
			out.Payments = append(out.Payments, p)
		}
		if !verifiable {
			out.Missing = append(out.Missing, missingFiledPayment(regime, fiscalYear, int(m)))
		}
	}
	return out, nil
}

// missingFiledPayment names a month whose payment cannot be verified.
func missingFiledPayment(regime fiscal.RegimeCode, fiscalYear, month int) string {
	return fmt.Sprintf("filed_payment:%s:%04d-%02d", regime, fiscalYear, month)
}

func ascendingMonths(months []time.Month) error {
	for i, m := range months {
		if m < time.January || m > time.December {
			return fmt.Errorf("month %d: %w", int(m), ErrInvalidYear)
		}
		if i > 0 && m <= months[i-1] {
			return fmt.Errorf("months must be ascending and unique (%d after %d): %w", int(m), int(months[i-1]), ErrInvalidYear)
		}
	}
	return nil
}

// resolveMonth returns month m's payment (no Source when there is none) and
// whether it is verifiable.
func resolveMonth(m time.Month, src PaymentSources) (Payment, bool, error) {
	if v, ok := src.Filed(m); ok {
		return typedPayment(m, v, PaymentFiled)
	}
	if v, ok := src.TaxInput(m); ok {
		return typedPayment(m, v, PaymentTaxInput)
	}
	if e, ok := src.AppEstimate(m); ok {
		p := Payment{Month: int(m), Amount: max(e.ISR, 0), Source: PaymentEstimatedByApp, FromIncompleteEstimate: e.Incomplete}
		return p, !e.Incomplete, nil
	}
	return Payment{}, false, nil
}

func typedPayment(m time.Month, v money.Cents, s PaymentSource) (Payment, bool, error) {
	if v < 0 {
		return Payment{}, false, fmt.Errorf("%s payment of month %d is negative: %w", s, int(m), ErrInvalidYear)
	}
	return Payment{Month: int(m), Amount: v, Source: s}, true, nil
}

// PaymentTotal is the sum of the payments of one source and how many there
// are.
type PaymentTotal struct {
	Source PaymentSource
	Amount money.Cents
	Months int
}

// Totals sums the payments per source, in precedence order, omitting a
// source with no payment; money.ErrOverflow past int64.
func (r ResolvedPayments) Totals() ([]PaymentTotal, error) {
	var out []PaymentTotal
	for _, s := range paymentPrecedence {
		t := PaymentTotal{Source: s}
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
func (r ResolvedPayments) EstimatedMonths() int {
	n := 0
	for _, p := range r.Payments {
		if p.Source == PaymentEstimatedByApp {
			n++
		}
	}
	return n
}
