package usecase

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// EstimateDeps are the ports the estimate use case needs. Tx must be the
// manager the taxpayers profile reader runs on, so the profile joins the
// estimate's unit of work (tx.ErrForeignTransaction otherwise).
type EstimateDeps struct {
	Tx         tx.Manager
	Profiles   TaxpayerProfileReader
	Ledger     LedgerReader
	Inputs     TaxInputReader
	Texts      RegimeTextReader
	Holidays   HolidayCalendar
	Strategies domain.Registry
	Params     ParamsCatalog
	Clock      clock.Clock
}

// RegimeResult is one regime's estimate as the taxpayer sees it: the
// strategy's figures, the catalog treatment and the derived status.
type RegimeResult struct {
	domain.RegimeEstimate
	// Treatment is the catalog's support level (contract treatment).
	Treatment domain.Treatment
	Status    domain.Status
	// Text is the catalog's name, explanation and input labels.
	Text RegimeText
}

// TaxpayerEstimate is the informational estimate of one month (contract
// TaxpayerEstimate, MODULE-tax-engine §2.1), derived on every read.
type TaxpayerEstimate struct {
	TaxpayerID  string
	PersonType  fiscal.PersonType
	Period      domain.Period
	DueDate     fiscal.Date
	Legend      string
	Disclaimers []domain.Disclaimer
	// Regimes are the regimes held in the period, in catalog (code) order.
	Regimes []RegimeResult
	// IVA is the consolidated monthly IVA with its trace; Totals.IVA is its
	// result.
	IVA          domain.ConsolidatedIVA
	Totals       domain.Totals
	ParameterSet taxparams.SetID
	// ProvisionalParameters: the period's fiscal year has no published set
	// and ParameterSet is an earlier year's (ADR-0002 §3).
	ProvisionalParameters bool
	RecordsFingerprint    string
	EngineVersion         string
	// Records are the period's live records the estimate read, summed over
	// every regime (informational ones included): no contract field of the
	// estimate, the dashboard's income vs expense view (M6-02).
	Records RecordTotals
}

// RecordTotals are one month's live records before IVA: incomes collected
// (collected_on) and expenses paid (paid_on, every deductibility class).
type RecordTotals struct {
	IncomeCollected money.Cents
	ExpensesPaid    money.Cents
}

// Estimator computes a taxpayer's monthly estimate (Estimator, M5-22): the
// one entry point M5-23 (HTTP), M6 (dashboard), M8 (history) and M9
// (reminder amounts) use.
type Estimator struct {
	deps EstimateDeps
}

// NewEstimator returns the use case over deps. A missing port is
// ErrMissingDependency, so a wiring mistake fails the boot rather than the
// first request (FU4, M5-31 G-2). Strategies is a value: an empty registry
// estimates every regime as unsupported, never a nil dereference.
func NewEstimator(deps EstimateDeps) (*Estimator, error) {
	ports := []struct {
		name    string
		missing bool
	}{
		{"Tx", deps.Tx == nil}, {"Profiles", deps.Profiles == nil}, {"Ledger", deps.Ledger == nil},
		{"Inputs", deps.Inputs == nil}, {"Texts", deps.Texts == nil}, {"Holidays", deps.Holidays == nil},
		{"Params", deps.Params == nil}, {"Clock", deps.Clock == nil},
	}
	for _, p := range ports {
		if p.missing {
			return nil, fmt.Errorf("%w: EstimateDeps.%s", ErrMissingDependency, p.name)
		}
	}
	return &Estimator{deps: deps}, nil
}

// regimeRIF is the bimonthly régimen 621: its payment follows the bimester.
const regimeRIF fiscal.RegimeCode = "621"

// ledgerRecords is what one estimate reads, in one unit of work.
type ledgerRecords struct {
	profile  TaxpayerProfile
	incomes  []IncomeMonthSums
	expenses []ExpenseMonthSums
	inputs   map[fiscal.RegimeCode]domain.TaxInputs
	texts    map[fiscal.RegimeCode]RegimeText
}

// EstimateMonth estimates rawPeriod ("YYYY-MM"; "" is the current month in
// Mexico City) for the caller's taxpayer (getTaxpayerEstimate).
//
// Order: taxpayer id (404) → period format (400 invalid) → a future month
// (400 out_of_range) → before the first published year (422
// parameters_unavailable) → then, in one read-only unit of work, the
// profile (404), the ledger aggregates, the tax inputs and the due date's
// holidays. The computation is pure: each held regime's strategy through
// the registry, IVA netted per taxpayer (OQ-T2, no carry-forward OQ-T5),
// ISR summed per regime without cross offsets (OQ-T10), withholdings to
// remit counted once, disclaimers, and the records fingerprint.
func (uc *Estimator) EstimateMonth(ctx context.Context, userID, taxpayerID, rawPeriod string) (TaxpayerEstimate, error) {
	if _, err := uuid.Parse(taxpayerID); err != nil {
		return TaxpayerEstimate{}, notFound(errors.Join(ErrTaxpayerNotFound, err))
	}
	p, err := uc.period(rawPeriod)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	out, err := uc.EstimateRange(ctx, userID, taxpayerID, p, p)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	return out[0], nil
}

// EstimateRange estimates every month of [from, to], oldest first (the
// dashboard's trend, M6-02, MODULE-dashboard §2.4). It reads once, in one
// read-only unit of work: the profile (404), one ledger statement per
// table from January of from's fiscal year through to, the tax inputs of
// each fiscal year and the catalog texts; then runs the pure engine per
// month, each fiscal year on its own records. Each estimate is exactly
// what EstimateMonth answers for its month (EstimateMonth is the range of
// one month).
//
// A malformed taxpayer id is 404; to after the current month is 400
// out_of_range; from before the first published year is 422
// parameters_unavailable; from after to is ErrInvalidRange (a caller bug).
func (uc *Estimator) EstimateRange(ctx context.Context, userID, taxpayerID string, from, to domain.Period) ([]TaxpayerEstimate, error) {
	id, err := uuid.Parse(taxpayerID)
	if err != nil {
		return nil, notFound(errors.Join(ErrTaxpayerNotFound, err))
	}
	months, err := uc.window(from, to)
	if err != nil {
		return nil, err
	}
	selections := map[int]taxparams.Selection{}
	for year := from.Year(); year <= to.Year(); year++ {
		if selections[year], err = uc.selectParams(year); err != nil {
			return nil, err
		}
	}
	var (
		records rangeRecords
		dues    = make([]fiscal.Date, len(months))
	)
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		var err error
		if records, err = uc.read(ctx, userID, id.String(), from, to); err != nil {
			return err
		}
		for i, p := range months {
			if dues[i], err = uc.dueDate(ctx, p, heldIn(records.profile.Regimes, p)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, present("estimate range", err)
	}
	out := make([]TaxpayerEstimate, 0, len(months))
	for i, p := range months {
		year := records.year(p.Year())
		est, err := uc.compute(p, selections[p.Year()], year)
		if err != nil {
			return nil, fmt.Errorf("taxengine: estimate %s: %w", p, err)
		}
		if est.Records, err = recordTotals(year, p.Month()); err != nil {
			return nil, fmt.Errorf("taxengine: records of %s: %w", p, err)
		}
		est.TaxpayerID, est.DueDate = id.String(), dues[i]
		out = append(out, est)
	}
	return out, nil
}

// window checks [from, to] against the clock and the published parameters
// and lists its months, oldest first.
func (uc *Estimator) window(from, to domain.Period) ([]domain.Period, error) {
	if from.IsZero() || to.IsZero() || monthIndex(from) > monthIndex(to) {
		return nil, fmt.Errorf("%w: %s..%s", ErrInvalidRange, from, to)
	}
	current, err := uc.currentPeriod()
	if err != nil {
		return nil, err
	}
	switch {
	case monthIndex(to) > monthIndex(current):
		return nil, futurePeriodError()
	case from.Year() < firstParamsYear:
		return nil, parametersUnavailableError(fieldPeriod)
	}
	months := make([]domain.Period, 0, monthIndex(to)-monthIndex(from)+1)
	for index := monthIndex(from); index <= monthIndex(to); index++ {
		p, err := domain.NewPeriod(index/monthsPerYear, time.Month(index%monthsPerYear+1))
		if err != nil {
			return nil, err
		}
		months = append(months, p)
	}
	return months, nil
}

// recordTotals sums the month's rows of one fiscal year over every regime.
func recordTotals(r ledgerRecords, month time.Month) (RecordTotals, error) {
	var (
		out  RecordTotals
		errs []error
	)
	add := func(total *money.Cents, amount money.Cents) {
		sum, err := total.Add(amount)
		errs = append(errs, err)
		*total = sum
	}
	for _, row := range r.incomes {
		if row.Month == month {
			add(&out.IncomeCollected, row.Sums.Cash)
		}
	}
	for _, row := range r.expenses {
		if row.Month == month {
			add(&out.ExpensesPaid, row.Sums.DeductiblePaid)
			add(&out.ExpensesPaid, row.Sums.NonDeductiblePaid)
			add(&out.ExpensesPaid, row.Sums.CashOverLimitPaid)
		}
	}
	return out, errors.Join(errs...)
}

// period resolves rawPeriod against the clock's current month.
func (uc *Estimator) period(raw string) (domain.Period, error) {
	current, err := uc.currentPeriod()
	if err != nil {
		return domain.Period{}, err
	}
	if raw == "" {
		return current, nil
	}
	p, err := domain.ParsePeriod(raw)
	switch {
	case err != nil:
		return domain.Period{}, invalidPeriodError(err)
	case monthIndex(p) > monthIndex(current):
		return domain.Period{}, futurePeriodError()
	case p.Year() < firstParamsYear:
		return domain.Period{}, parametersUnavailableError(fieldPeriod)
	}
	return p, nil
}

// currentPeriod is the clock's month in Mexico City.
func (uc *Estimator) currentPeriod() (domain.Period, error) {
	now := uc.deps.Clock.Now()
	current, err := domain.NewPeriod(now.Year(), now.Month())
	if err != nil {
		return domain.Period{}, fmt.Errorf("taxengine: current period: %w", err)
	}
	return current, nil
}

// selectParams picks the fiscal year's parameter set (ADR-0002 §3).
func (uc *Estimator) selectParams(fiscalYear int) (taxparams.Selection, error) {
	sel, err := uc.deps.Params.Select(fiscalYear)
	if errors.Is(err, taxparams.ErrNoParameters) {
		return taxparams.Selection{}, parametersUnavailableError(fieldPeriod)
	}
	if err != nil {
		return taxparams.Selection{}, fmt.Errorf("taxengine: select parameters %d: %w", fiscalYear, err)
	}
	return sel, nil
}

// rangeRecords is what an estimate range reads, in one unit of work: the
// ledger rows of every fiscal year in it and each year's tax inputs.
type rangeRecords struct {
	profile  TaxpayerProfile
	incomes  []IncomeMonthSums
	expenses []ExpenseMonthSums
	inputs   map[int]map[fiscal.RegimeCode]domain.TaxInputs
	texts    map[fiscal.RegimeCode]RegimeText
}

// year narrows the range to one fiscal year's records.
func (r rangeRecords) year(fiscalYear int) ledgerRecords {
	return ledgerRecords{
		profile:  r.profile,
		incomes:  slices.DeleteFunc(slices.Clone(r.incomes), func(row IncomeMonthSums) bool { return row.Year != fiscalYear }),
		expenses: slices.DeleteFunc(slices.Clone(r.expenses), func(row ExpenseMonthSums) bool { return row.Year != fiscalYear }),
		inputs:   r.inputs[fiscalYear],
		texts:    r.texts,
	}
}

// read loads the profile, the aggregates of January of from's fiscal year
// through to (one statement per table), each fiscal year's tax inputs and
// the catalog texts.
func (uc *Estimator) read(ctx context.Context, userID, taxpayerID string, from, to domain.Period) (rangeRecords, error) {
	profile, err := uc.deps.Profiles.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return rangeRecords{}, err
	}
	q := LedgerQuery{TaxpayerID: taxpayerID, FiscalYear: to.Year(), Through: to.Month()}
	if from.Year() != to.Year() {
		q.FirstYear = from.Year()
	}
	incomes, err := uc.deps.Ledger.IncomeSums(ctx, q)
	if err != nil {
		return rangeRecords{}, fmt.Errorf("income sums: %w", err)
	}
	expenses, err := uc.deps.Ledger.ExpenseSums(ctx, q)
	if err != nil {
		return rangeRecords{}, fmt.Errorf("expense sums: %w", err)
	}
	inputs := map[int]map[fiscal.RegimeCode]domain.TaxInputs{}
	for year := from.Year(); year <= to.Year(); year++ {
		if inputs[year], err = uc.deps.Inputs.TaxInputs(ctx, taxpayerID, year); err != nil {
			return rangeRecords{}, fmt.Errorf("tax inputs %d: %w", year, err)
		}
	}
	texts, err := uc.deps.Texts.RegimeTexts(ctx, taxpayerID)
	if err != nil {
		return rangeRecords{}, fmt.Errorf("regime texts: %w", err)
	}
	return rangeRecords{profile: profile, incomes: incomes, expenses: expenses, inputs: inputs, texts: texts}, nil
}

// compute derives the estimate from what was read: no I/O.
func (uc *Estimator) compute(p domain.Period, sel taxparams.Selection, r ledgerRecords) (TaxpayerEstimate, error) {
	years, err := yearSumsThrough(p, r)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	current, err := uc.runMonth(p, years, r, sel.Params)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	previousInFavor, err := uc.previousInFavor(p, r, sel.Params)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	iva, err := domain.ConsolidateIVA(current.contribs, previousInFavor)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	estimates := make([]domain.RegimeEstimate, 0, len(current.regimes))
	for i := range current.regimes {
		res := &current.regimes[i]
		res.Text = r.texts[res.Key.Regime]
		if contributesIVA(res.RegimeEstimate) {
			res.CPATopics = withTopics(res.CPATopics, iva.CPATopics)
		}
		estimates = append(estimates, res.RegimeEstimate)
	}
	totals, err := domain.NewTotals(estimates, iva.Result)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	fingerprint, err := fingerprintOf(p, sel, years, r)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	return TaxpayerEstimate{
		PersonType: r.profile.PersonType,
		Period:     p,
		Legend:     domain.Legend,
		Disclaimers: domain.EstimateDisclaimers(domain.DisclaimerFacts{
			Regimes: estimates, IVA: iva, ProvisionalParameters: sel.Provisional,
		}),
		Regimes:               current.regimes,
		IVA:                   iva,
		Totals:                totals,
		ParameterSet:          taxparams.SetID{FiscalYear: sel.Params.FiscalYear(), Version: sel.Params.Version()},
		ProvisionalParameters: sel.Provisional,
		RecordsFingerprint:    fingerprint,
		EngineVersion:         domain.EngineVersion,
	}, nil
}

// contributesIVA reports a regime whose IVA components enter the
// consolidated IVA with a non-zero amount (ConsolidatedIVA.Contributors).
func contributesIVA(e domain.RegimeEstimate) bool {
	return e.IVA != nil && *e.IVA != (domain.IVAComponents{})
}

// withTopics adds the consolidated IVA's topics to a contributing regime's,
// each once: the regime's IVA is netted by a rule no CPA has validated, so
// the regime is under review too (contract TX-NET-01, OQ-T2).
func withTopics(own, consolidated []domain.CPATopic) []domain.CPATopic {
	out := slices.Clone(own)
	for _, topic := range consolidated {
		if !slices.Contains(out, topic) {
			out = append(out, topic)
		}
	}
	return out
}

// monthRun is every held regime's estimate of one period and its share of
// the consolidated IVA.
type monthRun struct {
	regimes  []RegimeResult
	contribs []domain.IVAContribution
}

// runMonth runs the strategy of every regime held in p on years, the
// aggregates January..p.
func (uc *Estimator) runMonth(p domain.Period, years map[fiscal.RegimeCode]domain.YearSums, r ledgerRecords,
	params *taxparams.Params,
) (monthRun, error) {
	var run monthRun
	for _, hold := range heldIn(r.profile.Regimes, p) {
		in, err := regimeInput(hold, r.profile.PersonType, p, years, r.inputs)
		if err != nil {
			return monthRun{}, err
		}
		est, err := uc.estimateRegime(hold, in, activeMonths(hold, p), params)
		if err != nil {
			return monthRun{}, err
		}
		run.regimes = append(run.regimes, RegimeResult{RegimeEstimate: est, Treatment: hold.Treatment, Status: est.Status()})
		if est.IVA != nil {
			run.contribs = append(run.contribs, domain.IVAContribution{
				Regime: hold.Code, Components: *est.IVA, ExemptBase: in.Sums.Month(p.Month()).Income.ByRate.ExemptBase,
			})
		}
	}
	return run, nil
}

// estimateRegime runs the regime's strategy through the registry; a regime
// the catalog marks informational shows only its explanation, whatever is
// registered, so its records never change the numbers (MODULE-incomes OQ-4).
func (uc *Estimator) estimateRegime(hold RegimeHold, in domain.Input, active []time.Month, params *taxparams.Params) (domain.RegimeEstimate, error) {
	if hold.Treatment == domain.TreatmentInformational {
		return domain.NewInformationalStrategy(in.Key).Estimate(in, params)
	}
	return uc.deps.Strategies.Estimate(in, active, params)
}

// previousInFavor is the previous month's consolidated IVA in favor, or 0:
// shown, never carried forward (OQ-T5). January has no previous month in
// the fiscal year's aggregates.
func (uc *Estimator) previousInFavor(p domain.Period, r ledgerRecords, params *taxparams.Params) (money.Cents, error) {
	if p.Month() == time.January {
		return 0, nil
	}
	prev, err := domain.NewPeriod(p.Year(), p.Month()-1)
	if err != nil {
		return 0, err
	}
	years, err := yearSumsThrough(prev, r)
	if err != nil {
		return 0, err
	}
	run, err := uc.runMonth(prev, years, r, params)
	if err != nil {
		return 0, fmt.Errorf("previous month %s: %w", prev, err)
	}
	iva, err := domain.ConsolidateIVA(run.contribs, 0)
	if err != nil || iva.Result.Outcome() != domain.OutcomeInFavor {
		return 0, err
	}
	return -iva.Result.Amount, nil
}

// fingerprintOf hashes every regime held in some month of the year up to
// p: the previous month's regimes count too (its IVA line depends on them).
func fingerprintOf(p domain.Period, sel taxparams.Selection, years map[fiscal.RegimeCode]domain.YearSums,
	r ledgerRecords,
) (string, error) {
	in := domain.FingerprintInput{
		PersonType: r.profile.PersonType, Period: p, Provisional: sel.Provisional,
		ParameterSet: taxparams.SetID{FiscalYear: sel.Params.FiscalYear(), Version: sel.Params.Version()}.String(),
	}
	for _, hold := range r.profile.Regimes {
		active := activeMonths(hold, p)
		if len(active) == 0 {
			continue
		}
		regime, err := regimeInput(hold, r.profile.PersonType, p, years, r.inputs)
		if err != nil {
			return "", err
		}
		in.Regimes = append(in.Regimes, domain.FingerprintRegime{Input: regime, Treatment: hold.Treatment, ActiveMonths: active})
	}
	return domain.Fingerprint(in), nil
}

// yearSumsThrough folds the rows of January..p into each regime's sums.
func yearSumsThrough(p domain.Period, r ledgerRecords) (map[fiscal.RegimeCode]domain.YearSums, error) {
	m := p.Month()
	incomes := slices.DeleteFunc(slices.Clone(r.incomes), func(row IncomeMonthSums) bool { return row.Month > m })
	expenses := slices.DeleteFunc(slices.Clone(r.expenses), func(row ExpenseMonthSums) bool { return row.Month > m })
	return YearSumsByRegime(LedgerQuery{FiscalYear: p.Year(), Through: m}, incomes, expenses)
}

// regimeInput assembles the strategy input of hold for p; a regime with no
// records gets zero sums for January..p.
func regimeInput(hold RegimeHold, pt fiscal.PersonType, p domain.Period, years map[fiscal.RegimeCode]domain.YearSums,
	inputs map[fiscal.RegimeCode]domain.TaxInputs,
) (domain.Input, error) {
	sums, ok := years[hold.Code]
	if !ok {
		var err error
		if sums, err = domain.NewYearSums(p.Month()); err != nil {
			return domain.Input{}, err
		}
	}
	return domain.Input{
		Key:    domain.StrategyKey{Regime: hold.Code, PersonType: pt},
		Period: p, Options: hold.Options, Sums: sums, Inputs: inputs[hold.Code],
	}, nil
}

// heldIn returns the regimes whose window contains p, in catalog (code)
// order (§2.1); archived regimes count for periods up to their end.
func heldIn(regimes []RegimeHold, p domain.Period) []RegimeHold {
	held := slices.DeleteFunc(slices.Clone(regimes), func(h RegimeHold) bool { return !holds(h, monthIndex(p)) })
	slices.SortFunc(held, func(a, b RegimeHold) int { return cmp.Compare(a.Code, b.Code) })
	return held
}

// activeMonths are the months of p's fiscal year, January..p, in which the
// regime was held (Facts.ActiveMonths).
func activeMonths(h RegimeHold, p domain.Period) []time.Month {
	var months []time.Month
	for m := time.January; m <= p.Month(); m++ {
		if holds(h, monthIndex(p)-int(p.Month()-m)) {
			months = append(months, m)
		}
	}
	return months
}

// holds reports whether h's window contains the month at index.
func holds(h RegimeHold, index int) bool {
	return monthIndex(h.Start) <= index && (h.End.IsZero() || index <= monthIndex(h.End))
}

// monthIndex numbers months consecutively across years, so windows compare
// as integers.
func monthIndex(p domain.Period) int { return p.Year()*monthsPerYear + int(p.Month()) - 1 }

// monthsPerYear is the number of monthly tax periods in a year.
const monthsPerYear = 12
