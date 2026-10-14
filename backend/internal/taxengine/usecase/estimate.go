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
	IVA                domain.ConsolidatedIVA
	Totals             domain.Totals
	ParameterSet       taxparams.SetID
	RecordsFingerprint string
	EngineVersion      string
}

// Estimator computes a taxpayer's monthly estimate (Estimator, M5-22): the
// one entry point M5-23 (HTTP), M6 (dashboard), M8 (history) and M9
// (reminder amounts) use.
type Estimator struct {
	deps EstimateDeps
}

// NewEstimator returns the use case over deps.
func NewEstimator(deps EstimateDeps) *Estimator { return &Estimator{deps: deps} }

// regimeRIF is the bimonthly régimen 621: its payment follows the bimester.
const regimeRIF fiscal.RegimeCode = "621"

// ledgerRecords is what one estimate reads, in one unit of work.
type ledgerRecords struct {
	profile  TaxpayerProfile
	incomes  []IncomeMonthSums
	expenses []ExpenseMonthSums
	inputs   map[fiscal.RegimeCode]domain.TaxInputs
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
	id, err := uuid.Parse(taxpayerID)
	if err != nil {
		return TaxpayerEstimate{}, notFound(errors.Join(ErrTaxpayerNotFound, err))
	}
	p, err := uc.period(rawPeriod)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	sel, err := uc.selectParams(p)
	if err != nil {
		return TaxpayerEstimate{}, err
	}
	var (
		records ledgerRecords
		due     fiscal.Date
	)
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		var err error
		if records, err = uc.read(ctx, userID, id.String(), p); err != nil {
			return err
		}
		due, err = uc.dueDate(ctx, p, heldIn(records.profile.Regimes, p))
		return err
	})
	if err != nil {
		return TaxpayerEstimate{}, present("estimate month", err)
	}
	out, err := uc.compute(p, sel, records)
	if err != nil {
		return TaxpayerEstimate{}, fmt.Errorf("taxengine: estimate %s: %w", p, err)
	}
	out.TaxpayerID, out.DueDate = id.String(), due
	return out, nil
}

// period resolves rawPeriod against the clock's current month.
func (uc *Estimator) period(raw string) (domain.Period, error) {
	now := uc.deps.Clock.Now()
	current, err := domain.NewPeriod(now.Year(), now.Month())
	if err != nil {
		return domain.Period{}, fmt.Errorf("taxengine: current period: %w", err)
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

// selectParams picks the period's parameter set (ADR-0002 §3).
func (uc *Estimator) selectParams(p domain.Period) (taxparams.Selection, error) {
	sel, err := uc.deps.Params.Select(p.Year())
	if errors.Is(err, taxparams.ErrNoParameters) {
		return taxparams.Selection{}, parametersUnavailableError(fieldPeriod)
	}
	if err != nil {
		return taxparams.Selection{}, fmt.Errorf("taxengine: select parameters %d: %w", p.Year(), err)
	}
	return sel, nil
}

// read loads the profile, the year's aggregates and the tax inputs.
func (uc *Estimator) read(ctx context.Context, userID, taxpayerID string, p domain.Period) (ledgerRecords, error) {
	profile, err := uc.deps.Profiles.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return ledgerRecords{}, err
	}
	q := LedgerQuery{TaxpayerID: taxpayerID, FiscalYear: p.Year(), Through: p.Month()}
	incomes, err := uc.deps.Ledger.IncomeSums(ctx, q)
	if err != nil {
		return ledgerRecords{}, fmt.Errorf("income sums: %w", err)
	}
	expenses, err := uc.deps.Ledger.ExpenseSums(ctx, q)
	if err != nil {
		return ledgerRecords{}, fmt.Errorf("expense sums: %w", err)
	}
	inputs, err := uc.deps.Inputs.TaxInputs(ctx, taxpayerID, p.Year())
	if err != nil {
		return ledgerRecords{}, fmt.Errorf("tax inputs: %w", err)
	}
	return ledgerRecords{profile: profile, incomes: incomes, expenses: expenses, inputs: inputs}, nil
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
	for _, res := range current.regimes {
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
		Regimes:            current.regimes,
		IVA:                iva,
		Totals:             totals,
		ParameterSet:       taxparams.SetID{FiscalYear: sel.Params.FiscalYear(), Version: sel.Params.Version()},
		RecordsFingerprint: fingerprint,
		EngineVersion:      domain.EngineVersion,
	}, nil
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
