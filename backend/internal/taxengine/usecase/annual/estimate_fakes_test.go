package annual_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// The annual estimate's fixtures are prefixed ann: the input use cases'
// fixtures share this test package.

// annProfiles serves one profile per taxpayer of userA.
type annProfiles struct {
	byID map[string]usecase.TaxpayerProfile
}

func (f *annProfiles) Profile(ctx context.Context, userID, taxpayerID string) (usecase.TaxpayerProfile, error) {
	p, ok := f.byID[taxpayerID]
	if !ok || userID != userA || ctxUser(ctx) != userA {
		return usecase.TaxpayerProfile{}, apperr.New(apperr.CodeNotFound, "Not found.")
	}
	return p, nil
}

// annLedger serves the year's rows and records the queries.
type annLedger struct {
	incomes  []usecase.IncomeMonthSums
	expenses []usecase.ExpenseMonthSums
	queries  []usecase.LedgerQuery
}

func (f *annLedger) IncomeSums(_ context.Context, q usecase.LedgerQuery) ([]usecase.IncomeMonthSums, error) {
	f.queries = append(f.queries, q)
	var out []usecase.IncomeMonthSums
	for _, r := range f.incomes {
		if r.Year == q.FiscalYear && r.Month <= q.Through {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *annLedger) ExpenseSums(_ context.Context, q usecase.LedgerQuery) ([]usecase.ExpenseMonthSums, error) {
	var out []usecase.ExpenseMonthSums
	for _, r := range f.expenses {
		if r.Year == q.FiscalYear && r.Month <= q.Through {
			out = append(out, r)
		}
	}
	return out, nil
}

// annTaxInputs serves each regime's tax inputs of the year.
type annTaxInputs struct {
	values map[fiscal.RegimeCode][]domain.TaxInputValue
}

func (f *annTaxInputs) TaxInputs(_ context.Context, _ string, first, last int) (map[int]map[fiscal.RegimeCode]domain.TaxInputs, error) {
	out := map[int]map[fiscal.RegimeCode]domain.TaxInputs{}
	for year := first; year <= last; year++ {
		out[year] = map[fiscal.RegimeCode]domain.TaxInputs{}
		for code, vs := range f.values {
			in, err := domain.NewTaxInputs(vs...)
			if err != nil {
				return nil, err
			}
			out[year][code] = in
		}
	}
	return out, nil
}

// annYearInputs serves the annual-only inputs.
type annYearInputs struct{ in annual.YearInputs }

func (f *annYearInputs) YearInputs(context.Context, string, int) (annual.YearInputs, error) {
	return f.in, nil
}

// annFiled serves the confirmed payments.
type annFiled struct {
	byRegime map[fiscal.RegimeCode]map[time.Month]money.Cents
}

func (f *annFiled) FiledPayments(context.Context, string, int) (map[fiscal.RegimeCode]map[time.Month]money.Cents, error) {
	return f.byRegime, nil
}

// annMonthly is the Estimator range: each regime's ISR per month and the
// months whose estimate misses an input other than the prior payments.
type annMonthly struct {
	isr        map[fiscal.RegimeCode]map[time.Month]money.Cents
	incomplete map[fiscal.RegimeCode][]time.Month
	calls      []string
}

func (f *annMonthly) EstimateRange(ctx context.Context, userID, _ string, from, to domain.Period) ([]usecase.TaxpayerEstimate, error) {
	f.calls = append(f.calls, ctxUser(ctx)+"|"+userID+"|"+from.String()+".."+to.String())
	var out []usecase.TaxpayerEstimate
	for m := from.Month(); m <= to.Month(); m++ {
		p, err := domain.NewPeriod(from.Year(), m)
		if err != nil {
			return nil, err
		}
		est := usecase.TaxpayerEstimate{Period: p}
		for code, byMonth := range f.isr {
			r := domain.RegimeEstimate{
				Key: domain.StrategyKey{Regime: code}, ISR: &domain.TaxResult{Amount: byMonth[m]},
				MissingInputs: []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January}}},
			}
			for _, im := range f.incomplete[code] {
				if im == m {
					r.MissingInputs = append(r.MissingInputs, domain.MissingInput{Key: domain.InputCoeficienteUtilidad})
				}
			}
			est.Regimes = append(est.Regimes, usecase.RegimeResult{RegimeEstimate: r})
		}
		out = append(out, est)
	}
	return out, nil
}

// annHolidays are the official non-business days.
type annHolidays struct{ days []string }

func (f annHolidays) IsHoliday(_ context.Context, d fiscal.Date) (bool, error) {
	for _, day := range f.days {
		if d.String() == day {
			return true, nil
		}
	}
	return false, nil
}

// annWorld is one annual estimate's ports.
type annWorld struct {
	tx        *fakeTx
	profiles  *annProfiles
	ledger    *annLedger
	taxInputs *annTaxInputs
	inputs    *annYearInputs
	filed     *annFiled
	monthly   *annMonthly
	holidays  annHolidays
	now       time.Time
}

// closedYear is a day of 2027: 2026 is closed.
var closedYear = time.Date(2027, time.February, 10, 12, 0, 0, 0, clock.Location())

func newAnnWorld() *annWorld {
	return &annWorld{
		tx: &fakeTx{},
		profiles: &annProfiles{byID: map[string]usecase.TaxpayerProfile{
			fisicaA: {PersonType: fiscal.PersonFisica, Regimes: []usecase.RegimeHold{{Code: "612", Start: mustPeriod("2022-01")}}},
			moralA: {PersonType: fiscal.PersonMoral, Regimes: []usecase.RegimeHold{
				{Code: "601", Start: mustPeriod("2024-01")}, {Code: "603", Start: mustPeriod("2024-01")},
			}},
		}},
		ledger:    &annLedger{},
		taxInputs: &annTaxInputs{values: map[fiscal.RegimeCode][]domain.TaxInputValue{}},
		inputs:    &annYearInputs{},
		filed:     &annFiled{byRegime: map[fiscal.RegimeCode]map[time.Month]money.Cents{}},
		monthly:   &annMonthly{isr: map[fiscal.RegimeCode]map[time.Month]money.Cents{}},
		now:       closedYear,
	}
}

func (w *annWorld) estimator(t *testing.T) *annual.Estimator {
	t.Helper()
	cat, err := params.Embedded()
	require.NoError(t, err)
	monthly, err := domain.NewRegistry(taxengine.Strategies()...)
	require.NoError(t, err)
	sections, err := annual.NewSectionRegistry(taxengine.AnnualSections(annuallimits.New(cat))...)
	require.NoError(t, err)
	uc, err := annual.NewEstimator(annual.EstimateDeps{
		Tx: w.tx, Profiles: w.profiles, Ledger: w.ledger, TaxInputs: w.taxInputs, YearInputs: w.inputs,
		Filed: w.filed, Monthly: w.monthly, MonthlyStrategies: monthly, Sections: sections,
		Holidays: w.holidays, Params: cat, Clock: clock.NewFixed(w.now),
	})
	require.NoError(t, err)
	return uc
}

// income records a regime's month of income collected (and accrued) with
// the ISR a persona moral withheld.
func (w *annWorld) income(regime fiscal.RegimeCode, year int, m time.Month, amount, withheld money.Cents) {
	w.ledger.incomes = append(w.ledger.incomes, usecase.IncomeMonthSums{
		Regime: regime, Year: year, Month: m,
		Sums: domain.IncomeSums{Cash: amount, Accrual: amount, ISRWithheld: domain.Withheld{ByMoral: withheld}},
	})
}

// expense records a regime's month of deductible expenses paid.
func (w *annWorld) expense(regime fiscal.RegimeCode, year int, m time.Month, amount money.Cents) {
	w.ledger.expenses = append(w.ledger.expenses, usecase.ExpenseMonthSums{
		Regime: regime, Year: year, Month: m, Sums: domain.ExpenseSums{DeductiblePaid: amount},
	})
}

// fileYear confirms every month of the year: total in December, zero
// before.
func (w *annWorld) fileYear(regime fiscal.RegimeCode, total money.Cents) {
	months := map[time.Month]money.Cents{}
	for m := time.January; m < time.December; m++ {
		months[m] = 0
	}
	months[time.December] = total
	w.filed.byRegime[regime] = months
}

// an01 is research AN-01 (MODULE-annual §3) for fisicaA: 612 collected
// 900,000, deductions 400,000, 30,000 withheld, a medical deduction of
// 40,000.
func (w *annWorld) an01() {
	w.income("612", 2026, time.June, 900_000_00, 30_000_00)
	w.expense("612", 2026, time.June, 400_000_00)
	w.inputs.in.Deductions = []adomain.Deduction{{
		Category: adomain.CategoryMedical, Amount: 40_000_00, PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiarySelf,
	}}
}
