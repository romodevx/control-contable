package annual

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/explain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// ErrMissingDependency reports an EstimateDeps port left nil: a wiring
// mistake fails the boot, never the first request.
var ErrMissingDependency = errors.New("annual: missing dependency")

// MonthlyRange is the M5 Estimator's range form (usecase.Estimator): the
// app's monthly estimates, the default provisional payments (§2.5). It
// joins the annual estimate's unit of work, so the composition root gives
// both the same tx.Manager.
type MonthlyRange interface {
	EstimateRange(ctx context.Context, userID, taxpayerID string, from, to domain.Period) ([]usecase.TaxpayerEstimate, error)
}

// EstimateDeps are the ports of the annual estimate (ADR-0009).
type EstimateDeps struct {
	Tx         tx.Manager
	Profiles   usecase.TaxpayerProfileReader
	Ledger     usecase.LedgerReader
	TaxInputs  usecase.TaxInputReader
	YearInputs YearInputsReader
	Filed      FiledPaymentsReader
	Monthly    MonthlyRange
	// MonthlyStrategies is the monthly registry: it tells which regimes'
	// monthly estimates credit the typed prior payments (chained defaults).
	MonthlyStrategies domain.Registry
	Sections          SectionRegistry
	Holidays          usecase.HolidayCalendar
	Params            usecase.ParamsCatalog
	Clock             clock.Clock
}

// Estimator computes the annual estimate (getAnnualEstimate,
// getAnnualBreakdown) and the filed-payments columns (getFiledPayments),
// derived on every read, never stored (ADR-0002 §5).
type Estimator struct{ deps EstimateDeps }

// NewEstimator returns the use case over deps.
func NewEstimator(deps EstimateDeps) (*Estimator, error) {
	ports := []struct {
		name    string
		missing bool
	}{
		{"Tx", deps.Tx == nil}, {"Profiles", deps.Profiles == nil}, {"Ledger", deps.Ledger == nil},
		{"TaxInputs", deps.TaxInputs == nil}, {"YearInputs", deps.YearInputs == nil}, {"Filed", deps.Filed == nil},
		{"Monthly", deps.Monthly == nil}, {"Holidays", deps.Holidays == nil}, {"Params", deps.Params == nil},
		{"Clock", deps.Clock == nil},
	}
	for _, p := range ports {
		if p.missing {
			return nil, fmt.Errorf("%w: EstimateDeps.%s", ErrMissingDependency, p.name)
		}
	}
	return &Estimator{deps: deps}, nil
}

// yearWindow is the fiscal year an estimate reads: through is the last
// month counted (December for a closed year, the last closed month of the
// current one, 0 in its January).
type yearWindow struct {
	fiscalYear int
	through    time.Month
	inProgress bool
}

// Estimate computes the annual estimate of fiscalYear for the caller's
// taxpayer. Order: taxpayer id (404) → a future year (400 out_of_range) →
// before 2026 or without the year's own parameters (422) → then, in one
// read-only unit of work, the profile (404), one ledger pass for the year,
// the tax inputs, the annual inputs, the filed payments, at most one
// Estimator range (only when a month needs a default payment) and the
// deadlines' holidays. The computation is pure.
func (uc *Estimator) Estimate(ctx context.Context, userID, rawTaxpayerID string, fiscalYear int) (Estimate, error) {
	id, w, sel, err := uc.resolve(rawTaxpayerID, fiscalYear)
	if err != nil {
		return Estimate{}, err
	}
	var r readings
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		r, err = uc.read(ctx, userID, id, w, sel.Params, "")
		return err
	})
	if err != nil {
		return Estimate{}, present("annual estimate", err)
	}
	est, err := uc.build(id, w, sel, r)
	if err != nil {
		return Estimate{}, fmt.Errorf("annual: estimate %d: %w", fiscalYear, err)
	}
	return est, nil
}

// Breakdown is the trace of one computed section of the estimate: the
// estimate's refusals first, then a kind the estimate does not have is
// 422 section_not_present.
func (uc *Estimator) Breakdown(ctx context.Context, userID, rawTaxpayerID string, fiscalYear int, kind string) (Breakdown, error) {
	est, err := uc.Estimate(ctx, userID, rawTaxpayerID, fiscalYear)
	if err != nil {
		return Breakdown{}, err
	}
	k := adomain.SectionKind(kind)
	i := slices.IndexFunc(est.Sections, func(s Section) bool { return s.Kind == k })
	if !IsComputedKind(k) || i < 0 {
		return Breakdown{}, sectionNotPresentError()
	}
	section := est.Sections[i]
	disclaimers, err := sectionDisclaimers(est, section)
	if err != nil {
		return Breakdown{}, fmt.Errorf("annual: breakdown %s: %w", k, err)
	}
	return Breakdown{
		TaxpayerID: est.TaxpayerID, FiscalYear: est.FiscalYear, Section: section, Legend: est.Legend,
		Disclaimers: disclaimers, ParameterSet: est.ParameterSet, EngineVersion: est.EngineVersion, RangeUsed: est.RangeUsed,
	}, nil
}

// Payments returns the twelve months of regime's provisional payments with
// every source's value (getFiledPayments' columns): the Estimator range
// always runs, so the app's estimate shows beside a confirmed month. The
// caller checks the regime first (FiledPayments.Get).
func (uc *Estimator) Payments(ctx context.Context, userID, rawTaxpayerID string, regime fiscal.RegimeCode, fiscalYear int) (RegimePayments, error) {
	id, w, sel, err := uc.resolve(rawTaxpayerID, fiscalYear)
	if err != nil {
		return RegimePayments{}, err
	}
	var r readings
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		r, err = uc.read(ctx, userID, id, w, sel.Params, regime)
		return err
	})
	if err != nil {
		return RegimePayments{}, present("annual payments", err)
	}
	if got, ok := r.book[regime]; ok {
		return got, nil
	}
	return ResolveRegimePayments(fiscalYear, RegimeSources{Regime: regime, Inputs: r.inputs[regime]})
}

// resolve checks the taxpayer id and the year and selects the parameters.
func (uc *Estimator) resolve(rawTaxpayerID string, fiscalYear int) (string, yearWindow, taxparams.Selection, error) {
	id, err := taxpayerID(rawTaxpayerID)
	if err != nil {
		return "", yearWindow{}, taxparams.Selection{}, err
	}
	now := uc.deps.Clock.Now().In(clock.Location())
	switch {
	case fiscalYear > now.Year():
		return "", yearWindow{}, taxparams.Selection{}, futureYearError(FieldFiscalYearPath)
	case fiscalYear < firstParamsYear:
		return "", yearWindow{}, taxparams.Selection{}, parametersUnavailableError(FieldFiscalYearPath)
	}
	sel, err := uc.deps.Params.Select(fiscalYear)
	switch {
	case errors.Is(err, taxparams.ErrNoParameters) || (err == nil && sel.Fallback):
		// The annual reads the year's own tariff and deadlines: an earlier
		// year's set cannot stand in.
		return "", yearWindow{}, taxparams.Selection{}, parametersUnavailableError(FieldFiscalYearPath)
	case err != nil:
		return "", yearWindow{}, taxparams.Selection{}, fmt.Errorf("annual: parameters %d: %w", fiscalYear, err)
	}
	w := yearWindow{fiscalYear: fiscalYear, through: time.December}
	if fiscalYear == now.Year() {
		w.through, w.inProgress = now.Month()-1, true
	}
	return id, w, sel, nil
}

// readings are what one estimate reads in its unit of work.
type readings struct {
	profile   usecase.TaxpayerProfile
	held      []usecase.RegimeHold
	incomes   []usecase.IncomeMonthSums
	expenses  []usecase.ExpenseMonthSums
	inputs    map[fiscal.RegimeCode]domain.TaxInputs
	year      YearInputs
	book      PaymentBook
	rangeUsed bool
	deadlines map[explain.DeadlineKind]fiscal.Date
}

// read loads everything the estimate needs. forRegime forces the
// Estimator range for that regime's columns (Payments); "" runs it only
// when a month has neither a filed payment nor a tax input.
func (uc *Estimator) read(ctx context.Context, userID, id string, w yearWindow, p *taxparams.Params, forRegime fiscal.RegimeCode) (readings, error) {
	var (
		r   readings
		err error
	)
	if r.profile, err = uc.deps.Profiles.Profile(ctx, userID, id); err != nil {
		return readings{}, err
	}
	r.held = heldIn(r.profile.Regimes, w.fiscalYear)
	if w.through > 0 {
		q := usecase.LedgerQuery{TaxpayerID: id, FiscalYear: w.fiscalYear, Through: w.through}
		if r.incomes, err = uc.deps.Ledger.IncomeSums(ctx, q); err != nil {
			return readings{}, fmt.Errorf("income sums: %w", err)
		}
		if r.expenses, err = uc.deps.Ledger.ExpenseSums(ctx, q); err != nil {
			return readings{}, fmt.Errorf("expense sums: %w", err)
		}
	}
	byYear, err := uc.deps.TaxInputs.TaxInputs(ctx, id, w.fiscalYear, w.fiscalYear)
	if err != nil {
		return readings{}, fmt.Errorf("tax inputs: %w", err)
	}
	r.inputs = byYear[w.fiscalYear]
	if r.year, err = uc.deps.YearInputs.YearInputs(ctx, id, w.fiscalYear); err != nil {
		return readings{}, fmt.Errorf("annual inputs: %w", err)
	}
	if err := uc.readPayments(ctx, userID, id, w, &r, forRegime); err != nil {
		return readings{}, err
	}
	if r.deadlines, err = uc.deadlines(ctx, p, w.fiscalYear); err != nil {
		return readings{}, err
	}
	return r, nil
}

// readPayments resolves every paying regime's months by precedence, with
// one Estimator range call when a default is needed.
func (uc *Estimator) readPayments(ctx context.Context, userID, id string, w yearWindow, r *readings, forRegime fiscal.RegimeCode) error {
	filed, err := uc.deps.Filed.FiledPayments(ctx, id, w.fiscalYear)
	if err != nil {
		return fmt.Errorf("filed payments: %w", err)
	}
	var sources []RegimeSources
	needsRange := false
	for _, h := range r.held {
		if !adomain.PaysProvisionalISR(h.Code, definitiveOption(h)) {
			continue
		}
		src := RegimeSources{
			Regime: h.Code, Active: activeMonths(h, w), Filed: filed[h.Code], Inputs: r.inputs[h.Code],
			Cumulative: uc.cumulative(h, r.profile.PersonType, w.fiscalYear),
		}
		needsRange = needsRange || h.Code == forRegime || needsDefault(src)
		sources = append(sources, src)
	}
	if needsRange && w.through > 0 {
		if err := uc.appEstimates(ctx, userID, id, w, sources); err != nil {
			return err
		}
		r.rangeUsed = true
	}
	r.book = make(PaymentBook, len(sources))
	for _, src := range sources {
		got, err := ResolveRegimePayments(w.fiscalYear, src)
		if err != nil {
			return err
		}
		r.book[src.Regime] = got
	}
	return nil
}

// appEstimates fills each source's App from one Estimator range,
// January..through.
func (uc *Estimator) appEstimates(ctx context.Context, userID, id string, w yearWindow, sources []RegimeSources) error {
	from, err := domain.NewPeriod(w.fiscalYear, time.January)
	if err != nil {
		return err
	}
	to, err := domain.NewPeriod(w.fiscalYear, w.through)
	if err != nil {
		return err
	}
	months, err := uc.deps.Monthly.EstimateRange(ctx, userID, id, from, to)
	if err != nil {
		return err
	}
	for i := range sources {
		sources[i].App = map[time.Month]MonthlyISR{}
		for _, est := range months {
			if isr, ok := monthlyISR(est, sources[i].Regime); ok {
				sources[i].App[est.Period.Month()] = isr
			}
		}
	}
	return nil
}

// monthlyISR is regime's ISR in one monthly estimate. Its missing prior
// provisional payments do not make it incomplete: the annual chains the
// payments it credits instead (ResolveRegimePayments).
func monthlyISR(est usecase.TaxpayerEstimate, regime fiscal.RegimeCode) (MonthlyISR, bool) {
	i := slices.IndexFunc(est.Regimes, func(r usecase.RegimeResult) bool { return r.Key.Regime == regime })
	if i < 0 || est.Regimes[i].ISR == nil {
		return MonthlyISR{}, false
	}
	r := est.Regimes[i]
	incomplete := slices.ContainsFunc(r.MissingInputs, func(m domain.MissingInput) bool {
		return m.Key != domain.InputPriorProvisionalPayments
	})
	return MonthlyISR{ISR: r.ISR.Amount, Incomplete: incomplete}, true
}

// needsDefault reports an active month with neither a filed payment nor a
// tax input.
func needsDefault(src RegimeSources) bool {
	return slices.ContainsFunc(src.Active, func(m time.Month) bool {
		if _, ok := src.Filed[m]; ok {
			return false
		}
		_, ok := src.Inputs.MonthlyMoney(domain.InputPriorProvisionalPayments, m)
		return !ok
	})
}

// cumulative reports a regime whose monthly strategy requires the prior
// provisional payments (its estimate is YTD tax less them).
func (uc *Estimator) cumulative(h usecase.RegimeHold, pt fiscal.PersonType, fiscalYear int) bool {
	s, _ := uc.deps.MonthlyStrategies.Lookup(domain.StrategyKey{Regime: h.Code, PersonType: pt})
	december, err := domain.NewPeriod(fiscalYear, time.December)
	if err != nil {
		return false
	}
	all := make([]time.Month, 0, monthsPerYear)
	for m := time.January; m <= time.December; m++ {
		all = append(all, m)
	}
	required := s.RequiredInputs(h.Options, december, domain.Facts{ActiveMonths: all})
	return slices.ContainsFunc(required, func(r domain.RequiredInput) bool { return r.Key == domain.InputPriorProvisionalPayments })
}

// deadlines are the legal annual dates of the year, each moved to the next
// business day (CFF 12 ⚠️, "sujeto a confirmación").
func (uc *Estimator) deadlines(ctx context.Context, p *taxparams.Params, fiscalYear int) (map[explain.DeadlineKind]fiscal.Date, error) {
	holidays, err := uc.deps.Holidays.Holidays(ctx)
	if err != nil {
		return nil, fmt.Errorf("annual: holidays: %w", err)
	}
	out := map[explain.DeadlineKind]fiscal.Date{}
	for _, kind := range []explain.DeadlineKind{explain.DeadlinePF, explain.DeadlinePM, explain.DeadlineNonprofitInformative} {
		legal, err := explain.Deadline(p, fiscalYear, kind)
		if err != nil {
			return nil, err
		}
		if out[kind], _, err = fiscal.NextBusinessDay(legal, holidays); err != nil {
			return nil, fmt.Errorf("annual: deadline %s: %w", kind, err)
		}
	}
	return out, nil
}

// heldIn are the regimes whose window intersects the fiscal year, in code
// order (archived ones for their active months).
func heldIn(regimes []usecase.RegimeHold, fiscalYear int) []usecase.RegimeHold {
	held := slices.DeleteFunc(slices.Clone(regimes), func(h usecase.RegimeHold) bool {
		return h.Start.Year() > fiscalYear || (!h.End.IsZero() && h.End.Year() < fiscalYear)
	})
	slices.SortFunc(held, func(a, b usecase.RegimeHold) int { return cmp.Compare(a.Code, b.Code) })
	return held
}

// activeMonths are the months of the window the regime is held in.
func activeMonths(h usecase.RegimeHold, w yearWindow) []time.Month {
	held := HeldRegime{Code: h.Code, Start: h.Start, End: h.End}
	var out []time.Month
	for m := time.January; m <= w.through; m++ {
		if held.Holds(w.fiscalYear, m) {
			out = append(out, m)
		}
	}
	return out
}

// definitiveOption is the 625 option of LISR 113-B.
func definitiveOption(h usecase.RegimeHold) bool {
	v, _ := h.Options.Bool(adomain.OptionPlatformDefinitive)
	return v
}

// sumMoney adds centavos, keeping the first overflow.
func sumMoney(values ...money.Cents) (money.Cents, error) {
	var total money.Cents
	for _, v := range values {
		var err error
		if total, err = total.Add(v); err != nil {
			return 0, err
		}
	}
	return total, nil
}
