package annual

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// PaymentSource is where a provisional payment comes from (MODULE-annual
// §2.5), in precedence order.
type PaymentSource string

// The payment sources, highest precedence first.
const (
	// PaymentFiled: the user confirmed or typed what the acuse says.
	PaymentFiled PaymentSource = "filed"
	// PaymentTaxInput: the M5 tax input prior_provisional_payments.
	PaymentTaxInput PaymentSource = "tax_input"
	// PaymentEstimatedByApp: the app's own monthly estimate, max(0, isr).
	PaymentEstimatedByApp PaymentSource = "estimated_by_app"
)

// paymentPrecedence is every source in precedence order.
var paymentPrecedence = []PaymentSource{PaymentFiled, PaymentTaxInput, PaymentEstimatedByApp}

// Valid reports a spec payment source.
func (s PaymentSource) Valid() bool { return slices.Contains(paymentPrecedence, s) }

// annualFiledPaymentsKey names the annual_filed_payments table as a line
// source: a value the user entered.
const annualFiledPaymentsKey domain.InputKey = "annual_filed_payments"

// lineSource is the breakdown source of a payment line.
func (s PaymentSource) lineSource() domain.Source {
	switch s {
	case PaymentFiled:
		return domain.SourceTaxInput(annualFiledPaymentsKey)
	case PaymentTaxInput:
		return domain.SourceTaxInput(domain.InputPriorProvisionalPayments)
	}
	return domain.SourceComputed()
}

// Payment is one provisional ISR payment of a regime, already resolved by
// precedence for its month. Month 0 is a year total as its source states
// it. FromIncompleteEstimate marks an app estimate whose monthly estimate
// was itself incomplete: the section is then incomplete (§2.6).
type Payment struct {
	Month                  int
	Amount                 money.Cents
	Source                 PaymentSource
	FromIncompleteEstimate bool
}

// PaymentsReader is the strategies' input port for the provisional
// payments of one taxpayer's fiscal year: ANN-14 implements it with the
// precedence of §2.5 (filed, then the tax input, then the app's estimate).
type PaymentsReader interface {
	ProvisionalPayments(regime fiscal.RegimeCode) ([]Payment, error)
}

const lastMonth = 12

func (p Payment) validate() error {
	switch {
	case p.Amount < 0:
		return fmt.Errorf("payment %s: amount must not be negative: %w", p.Amount, ErrInvalidYear)
	case !p.Source.Valid():
		return fmt.Errorf("payment source %q: %w", p.Source, ErrInvalidYear)
	case p.Month < 0 || p.Month > lastMonth:
		return fmt.Errorf("payment month %d: %w", p.Month, ErrInvalidYear)
	case p.FromIncompleteEstimate && (p.Source != PaymentEstimatedByApp || p.Month == 0):
		return fmt.Errorf("only a monthly app estimate can be incomplete: %w", ErrInvalidYear)
	}
	return nil
}

// oneSourcePerMonth rejects payments the port did not resolve by
// precedence: a month credited twice, or a year total (month 0) beside
// monthly payments, would be credited twice. O(n) over 13 slots.
func oneSourcePerMonth(payments []Payment) error {
	var seen [lastMonth + 1]bool
	yearTotal, monthly := false, false
	for _, p := range payments {
		if p.Month < 0 || p.Month > lastMonth {
			continue // Payment.validate reports it
		}
		if seen[p.Month] {
			return fmt.Errorf("payment month %d credited twice: %w", p.Month, ErrInvalidYear)
		}
		seen[p.Month] = true
		yearTotal, monthly = yearTotal || p.Month == 0, monthly || p.Month > 0
	}
	if yearTotal && monthly {
		return fmt.Errorf("a year total beside monthly payments: %w", ErrInvalidYear)
	}
	return nil
}

// paymentGroup is one regime's payments of one source: a breakdown line.
type paymentGroup struct {
	regime fiscal.RegimeCode
	source PaymentSource
	total  money.Cents
	count  int
}

// regimePayments are a regime's payments read through the port and grouped
// by source in precedence order; missing names each month whose default
// payment came from an incomplete estimate (filed_payment:<regime>:<YYYY-MM>).
type regimePayments struct {
	groups    []paymentGroup
	missing   []string
	estimated bool
}

func readPayments(r PaymentsReader, fiscalYear int, regime fiscal.RegimeCode) (regimePayments, error) {
	payments, err := r.ProvisionalPayments(regime)
	if err != nil {
		return regimePayments{}, fmt.Errorf("annual: provisional payments %s: %w", regime, err)
	}
	if err := oneSourcePerMonth(payments); err != nil {
		return regimePayments{}, fmt.Errorf("annual: %s: %w", regime, err)
	}
	var out regimePayments
	for _, source := range paymentPrecedence {
		g := paymentGroup{regime: regime, source: source}
		for _, p := range payments {
			if err := p.validate(); err != nil {
				return regimePayments{}, fmt.Errorf("annual: %s: %w", regime, err)
			}
			if p.Source != source {
				continue
			}
			if g.total, err = g.total.Add(p.Amount); err != nil {
				return regimePayments{}, fmt.Errorf("annual: %s payments: %w", regime, err)
			}
			g.count++
			if p.FromIncompleteEstimate {
				out.missing = append(out.missing, missingFiledPayment(regime, fiscalYear, p.Month))
			}
		}
		if g.count > 0 {
			out.groups = append(out.groups, g)
			out.estimated = out.estimated || source == PaymentEstimatedByApp
		}
	}
	slices.Sort(out.missing)
	return out, nil
}

// subtractPayments traces each group as a credit line naming its regime, source
// and number of payments, and returns their total.
func (t *annualTrace) subtractPayments(groups []paymentGroup, code domain.LineCode, label domain.LabelKey, basis []domain.Citation) money.Cents {
	var total money.Cents
	for _, g := range groups {
		t.l.Subtract(annualStep(code, label, g.source.lineSource(), basis, string(g.regime), string(g.source), strconv.Itoa(g.count)), t.exact(g.total))
		total = t.sum(total, g.total)
	}
	return total
}
