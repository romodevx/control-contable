package usecase_test

import (
	"context"
	"errors"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The fixtures of the estimate use case are prefixed est: the tax-inputs
// fixtures share this test package.

var errEstPortDown = errors.New("fake: port down")

// estUserOf reads the fake unit of work's user: a port called outside it
// sees "".
func estUserOf(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	return user
}

// estProfiles is the taxpayers module's profile reader.
type estProfiles struct {
	profile usecase.TaxpayerProfile
	err     error
	users   []string
}

func (f *estProfiles) Profile(ctx context.Context, userID, taxpayerID string) (usecase.TaxpayerProfile, error) {
	f.users = append(f.users, estUserOf(ctx)+"|"+userID+"|"+taxpayerID)
	return f.profile, f.err
}

// estLedger serves fixed aggregate rows and records the queries.
type estLedger struct {
	incomes  []usecase.IncomeMonthSums
	expenses []usecase.ExpenseMonthSums
	err      error
	queries  []usecase.LedgerQuery
	users    []string
}

func (f *estLedger) IncomeSums(ctx context.Context, q usecase.LedgerQuery) ([]usecase.IncomeMonthSums, error) {
	f.queries = append(f.queries, q)
	f.users = append(f.users, estUserOf(ctx))
	return f.incomes, f.err
}

func (f *estLedger) ExpenseSums(ctx context.Context, q usecase.LedgerQuery) ([]usecase.ExpenseMonthSums, error) {
	f.users = append(f.users, estUserOf(ctx))
	return f.expenses, f.err
}

// estInputs serves the tax inputs of every regime.
type estInputs struct {
	values map[fiscal.RegimeCode]domain.TaxInputs
	err    error
	users  []string
}

func (f *estInputs) TaxInputs(ctx context.Context, _ string, _ int) (map[fiscal.RegimeCode]domain.TaxInputs, error) {
	f.users = append(f.users, estUserOf(ctx))
	return f.values, f.err
}

// estHolidays is the holidays calendar.
type estHolidays struct {
	days map[string]bool
	err  error
}

func (f estHolidays) IsHoliday(_ context.Context, d fiscal.Date) (bool, error) {
	return f.days[d.String()], f.err
}

// estParams selects one synthetic parameter set (the fake strategies read
// none of it).
type estParams struct {
	provisional bool
	err         error
}

func (f estParams) Select(fiscalYear int) (taxparams.Selection, error) {
	if f.err != nil {
		return taxparams.Selection{}, f.err
	}
	return taxparams.Selection{Params: estParamSet(), RequestedYear: fiscalYear, Provisional: f.provisional}, nil
}

func estParamSet() *taxparams.Params {
	d := func(s string) fiscal.Date {
		date, err := fiscal.ParseDate(s)
		if err != nil {
			panic(err)
		}
		return date
	}
	p, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear: 2026, Version: 3, PublishedOn: d("2026-10-13"),
		Sources: []taxparams.Source{{
			ID: "synthetic", Title: "Synthetic", URL: "https://example.test/dof", DOFDate: d("2025-12-31"), Reference: "none",
		}},
	})
	if err != nil {
		panic(err)
	}
	return p
}

// estStrategy is a fake strategy: it returns the estimate of the period's
// month (or a zero full estimate), requires what it is told, and records
// every input it ran on.
type estStrategy struct {
	key       domain.StrategyKey
	treatment domain.Treatment
	required  []domain.RequiredInput
	byMonth   map[time.Month]domain.RegimeEstimate
	err       error
	ran       []domain.Input
	facts     []domain.Facts
}

func (s *estStrategy) Key() domain.StrategyKey { return s.key }

func (s *estStrategy) Treatment() domain.Treatment { return s.treatment }

func (s *estStrategy) RequiredInputs(_ domain.Options, _ domain.Period, f domain.Facts) []domain.RequiredInput {
	s.facts = append(s.facts, f)
	return s.required
}

func (s *estStrategy) Estimate(in domain.Input, _ *taxparams.Params) (domain.RegimeEstimate, error) {
	s.ran = append(s.ran, in)
	if s.err != nil {
		return domain.RegimeEstimate{}, s.err
	}
	est, ok := s.byMonth[in.Period.Month()]
	if !ok {
		est = domain.RegimeEstimate{Coverage: s.treatment}
	}
	est.Key = s.key
	return est, nil
}

func estKey(code fiscal.RegimeCode, pt fiscal.PersonType) domain.StrategyKey {
	return domain.StrategyKey{Regime: code, PersonType: pt}
}

func estISR(c money.Cents) *domain.TaxResult {
	r := domain.NewTaxResult(money.Exact(c) * money.ExactPerCent)
	return &r
}

// estHold is a regime held from start (YYYY-MM), open-ended.
func estHold(code fiscal.RegimeCode, start string, treatment domain.Treatment) usecase.RegimeHold {
	return usecase.RegimeHold{Code: code, Treatment: treatment, Start: mustPeriod(start)}
}

// estFixture wires the estimator over fakes; today is 14 October 2026.
type estFixture struct {
	tx       *fakeTx
	profiles *estProfiles
	ledger   *estLedger
	inputs   *estInputs
	holidays estHolidays
	params   estParams
	clock    clock.Clock
}

func newEstFixture(pt fiscal.PersonType, holds ...usecase.RegimeHold) *estFixture {
	return &estFixture{
		tx:       &fakeTx{},
		profiles: &estProfiles{profile: usecase.TaxpayerProfile{PersonType: pt, Regimes: holds}},
		ledger:   &estLedger{},
		inputs:   &estInputs{},
		clock:    clock.NewFixed(today),
	}
}

func (f *estFixture) estimator(strategies ...domain.Strategy) *usecase.Estimator {
	registry, err := domain.NewRegistry(strategies...)
	if err != nil {
		panic(err)
	}
	return usecase.NewEstimator(usecase.EstimateDeps{
		Tx: f.tx, Profiles: f.profiles, Ledger: f.ledger, Inputs: f.inputs, Holidays: f.holidays,
		Strategies: registry, Params: f.params, Clock: f.clock,
	})
}
