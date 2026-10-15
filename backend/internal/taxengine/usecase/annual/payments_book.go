package annual

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// MonthlyISR is the app's monthly ISR estimate of a regime for one month, as
// the monthly engine answered it: possibly negative (the strategy's own
// saldo a favor) and incomplete when an input other than the prior
// provisional payments was missing.
type MonthlyISR struct {
	ISR        money.Cents
	Incomplete bool
}

// RegimeSources are the three payment sources of one regime's fiscal year
// (MODULE-annual §2.5) and the months they apply to.
type RegimeSources struct {
	Regime fiscal.RegimeCode
	// Active are the months held and closed (through data_through),
	// ascending: the months a provisional payment is due for.
	Active []time.Month
	// Filed are the confirmed payments (annual_filed_payments).
	Filed map[time.Month]money.Cents
	// Inputs are the regime's M5 tax inputs of the year
	// (prior_provisional_payments per month).
	Inputs domain.TaxInputs
	// App are the monthly estimates the Estimator range answered.
	App map[time.Month]MonthlyISR
	// Cumulative reports a regime whose monthly estimate credits the typed
	// prior provisional payments (612, 601, RESICO PM): its defaults are
	// chained (ResolveRegimePayments).
	Cumulative bool
}

// AppColumn is the app's estimate of one month as the annual credits it:
// max(0, isr), chained for a cumulative regime.
type AppColumn struct {
	Present    bool
	Cents      money.Cents
	Incomplete bool
}

// Amount is the estimate, or nil when the month has none.
func (a AppColumn) Amount() *money.Cents {
	if !a.Present {
		return nil
	}
	v := a.Cents
	return &v
}

// MonthPayment is one month of a regime's provisional payments: each
// source's value and the one credited (Source "" when none is).
type MonthPayment struct {
	Month     time.Month
	Filed     *money.Cents
	TaxInput  *money.Cents
	App       AppColumn
	Effective *money.Cents
	Source    adomain.PaymentSource
}

// RegimePayments are one regime's twelve months (January first) and the
// payments the annual sections credit, resolved by precedence.
type RegimePayments struct {
	Regime   fiscal.RegimeCode
	Months   []MonthPayment
	Resolved adomain.ResolvedPayments
}

// ResolveRegimePayments applies the §2.5 precedence to each active month:
// the filed value, else the tax input, else the app's estimate as
// max(0, isr).
//
// A cumulative regime's monthly estimate is YTD tax − the typed prior
// payments (missing months count as zero), so summing the raw estimates
// would pay the year's tax several times. Its default is therefore
// chained: isr(m) + Σ typed prior payments before m − Σ payments the
// annual credits before m, which is exactly the monthly figure had the
// credited payments been typed (the strategies subtract the prior payments
// linearly, before any rounding). O(12).
func ResolveRegimePayments(fiscalYear int, src RegimeSources) (RegimePayments, error) {
	out := RegimePayments{Regime: src.Regime, Months: make([]MonthPayment, 0, monthsPerYear)}
	active := make(map[time.Month]bool, len(src.Active))
	for _, m := range src.Active {
		active[m] = true
	}
	var credited money.Cents
	for m := time.January; m <= time.December; m++ {
		month, err := resolveMonth(src, m, active[m], credited)
		if err != nil {
			return RegimePayments{}, fmt.Errorf("annual: payments %s %d-%02d: %w", src.Regime, fiscalYear, int(m), err)
		}
		if month.Effective != nil {
			if credited, err = credited.Add(*month.Effective); err != nil {
				return RegimePayments{}, fmt.Errorf("annual: payments %s: %w", src.Regime, err)
			}
		}
		out.Months = append(out.Months, month)
	}
	resolved, err := adomain.ResolvePayments(src.Regime, fiscalYear, src.Active, bookSources(out.Months))
	if err != nil {
		return RegimePayments{}, fmt.Errorf("annual: payments %s: %w", src.Regime, err)
	}
	out.Resolved = resolved
	return out, nil
}

// resolveMonth fills month m's columns; credited is what the annual
// credits before m.
func resolveMonth(src RegimeSources, m time.Month, active bool, credited money.Cents) (MonthPayment, error) {
	out := MonthPayment{Month: m}
	if v, ok := src.Filed[m]; ok {
		out.Filed = &v
	}
	if v, ok := src.Inputs.MonthlyMoney(domain.InputPriorProvisionalPayments, m); ok {
		out.TaxInput = &v
	}
	if !active {
		return out, nil
	}
	if est, ok := src.App[m]; ok {
		app, err := chained(src, m, est, credited)
		if err != nil {
			return MonthPayment{}, err
		}
		out.App = app
	}
	switch {
	case out.Filed != nil:
		out.Effective, out.Source = out.Filed, adomain.PaymentFiled
	case out.TaxInput != nil:
		out.Effective, out.Source = out.TaxInput, adomain.PaymentTaxInput
	case out.App.Present:
		out.Effective, out.Source = out.App.Amount(), adomain.PaymentEstimatedByApp
	}
	return out, nil
}

// chained is the month's default as the annual credits it.
func chained(src RegimeSources, m time.Month, est MonthlyISR, credited money.Cents) (AppColumn, error) {
	isr := est.ISR
	if src.Cumulative {
		typed, err := src.Inputs.SumBefore(domain.InputPriorProvisionalPayments, m)
		if err != nil {
			return AppColumn{}, err
		}
		if isr, err = isr.Add(typed); err != nil {
			return AppColumn{}, err
		}
		if isr, err = isr.Add(-credited); err != nil {
			return AppColumn{}, err
		}
	}
	return AppColumn{Present: true, Cents: max(isr, 0), Incomplete: est.Incomplete}, nil
}

// bookSources is adomain.PaymentSources over resolved months.
type bookSources []MonthPayment

func (b bookSources) Filed(m time.Month) (money.Cents, bool) { return presentCents(b[m-1].Filed) }

func (b bookSources) TaxInput(m time.Month) (money.Cents, bool) { return presentCents(b[m-1].TaxInput) }

func (b bookSources) AppEstimate(m time.Month) (adomain.AppEstimate, bool) {
	a := b[m-1].App
	return adomain.AppEstimate{ISR: a.Cents, Incomplete: a.Incomplete}, a.Present
}

func presentCents(p *money.Cents) (money.Cents, bool) {
	if p == nil {
		return 0, false
	}
	return *p, true
}

// PaymentBook is every paying regime's payments of the year: the strategies'
// adomain.PaymentsReader.
type PaymentBook map[fiscal.RegimeCode]RegimePayments

// ProvisionalPayments implements adomain.PaymentsReader: the regime's
// resolved payments, none for a regime without any.
func (b PaymentBook) ProvisionalPayments(regime fiscal.RegimeCode) ([]adomain.Payment, error) {
	return b[regime].Resolved.Payments, nil
}

// SourceCounts are how many regime-months came from each source (contract
// PaymentSourceCounts).
type SourceCounts struct {
	Filed, TaxInput, EstimatedByApp int
}

// Counts sums the sources of regimes' resolved payments.
func (b PaymentBook) Counts(regimes ...fiscal.RegimeCode) SourceCounts {
	var out SourceCounts
	for _, r := range regimes {
		for _, p := range b[r].Resolved.Payments {
			switch p.Source {
			case adomain.PaymentFiled:
				out.Filed++
			case adomain.PaymentTaxInput:
				out.TaxInput++
			case adomain.PaymentEstimatedByApp:
				out.EstimatedByApp++
			}
		}
	}
	return out
}
