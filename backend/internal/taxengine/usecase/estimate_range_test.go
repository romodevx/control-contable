package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The range form of the estimator (M6-02, MODULE-dashboard §2.4): the
// dashboard's window of up to twelve months, across two fiscal years, read
// in one unit of work with one ledger statement per table.

// rangeToday is 10 March 2027: a window can reach back into 2026.
var rangeToday = time.Date(2027, time.March, 10, 12, 0, 0, 0, clock.Location())

func cashIncome(year int, month time.Month, cash money.Cents) usecase.IncomeMonthSums {
	return usecase.IncomeMonthSums{Regime: fiscal.RegimeBusiness, Year: year, Month: month, Sums: domain.IncomeSums{Cash: cash}}
}

func paidExpense(year int, month time.Month, deductible, other, cashOver money.Cents) usecase.ExpenseMonthSums {
	return usecase.ExpenseMonthSums{Regime: fiscal.RegimeBusiness, Year: year, Month: month, Sums: domain.ExpenseSums{
		DeductiblePaid: deductible, NonDeductiblePaid: other, CashOverLimitPaid: cashOver,
	}}
}

func newRangeFixture() (*estFixture, *estStrategy) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	f.clock = clock.NewFixed(rangeToday)
	f.ledger.incomes = []usecase.IncomeMonthSums{
		cashIncome(2026, time.November, 100_00), cashIncome(2026, time.December, 50_00), cashIncome(2027, time.January, 200_00),
	}
	f.ledger.expenses = []usecase.ExpenseMonthSums{paidExpense(2027, time.January, 30_00, 20_00, 10_00)}
	return f, &estStrategy{key: est612, treatment: domain.TreatmentFull}
}

func TestEstimateRangeReadsOnceAcrossTwoFiscalYears(t *testing.T) {
	f, s612 := newRangeFixture()

	got, err := f.estimator(s612).EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2026-11"), mustPeriod("2027-02"))

	require.NoError(t, err)
	var periods []string
	for _, e := range got {
		periods = append(periods, e.Period.String())
		assert.Equal(t, taxpayerA, e.TaxpayerID)
		assert.False(t, e.DueDate.IsZero(), "%s has its due date", e.Period)
	}
	assert.Equal(t, []string{"2026-11", "2026-12", "2027-01", "2027-02"}, periods, "oldest first, one per month")
	want := usecase.LedgerQuery{TaxpayerID: taxpayerA, FirstYear: 2026, FiscalYear: 2027, Through: time.February}
	assert.Equal(t, []usecase.LedgerQuery{want}, f.ledger.queries, "one income statement for the whole window")
	assert.Equal(t, []usecase.LedgerQuery{want}, f.ledger.expenseQueries, "one expense statement for the whole window")
	assert.Equal(t, []yearSpan{{2026, 2027}}, f.inputs.spans, "one tax-inputs statement for both fiscal years")
	assert.Equal(t, []string{"ro"}, f.tx.modes, "one read-only unit of work")
	assert.Len(t, f.profiles.users, 1, "the profile is read once")
}

// Each fiscal year accumulates its own records only: January 2027 never
// sees November 2026.
func TestEstimateRangeKeepsFiscalYearsApart(t *testing.T) {
	f, s612 := newRangeFixture()

	_, err := f.estimator(s612).EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2026-11"), mustPeriod("2027-01"))

	require.NoError(t, err)
	byPeriod := map[string]domain.Input{}
	for _, in := range s612.ran {
		byPeriod[in.Period.String()] = in // the last run of a period is its own month (the previous-month IVA runs first)
	}
	jan := byPeriod["2027-01"].Sums
	assert.Equal(t, money.Cents(200_00), jan.Month(time.January).Income.Cash)
	all, err := jan.Sum(time.January, time.January)
	require.NoError(t, err)
	assert.Equal(t, money.Cents(200_00), all.Income.Cash, "2026's records stay in 2026")
	nov := byPeriod["2026-11"].Sums
	assert.Equal(t, money.Cents(100_00), nov.Month(time.November).Income.Cash)
}

// The records the estimate read in each month: income collected (cash,
// before IVA) and expenses paid (every deductibility class): the
// dashboard's income vs expense figures come from the same rows.
func TestEstimateRangeCarriesEachMonthsRecordTotals(t *testing.T) {
	f, s612 := newRangeFixture()

	got, err := f.estimator(s612).EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2026-11"), mustPeriod("2027-02"))

	require.NoError(t, err)
	require.Len(t, got, 4)
	assert.Equal(t, usecase.RecordTotals{IncomeCollected: 100_00}, got[0].Records)
	assert.Equal(t, usecase.RecordTotals{IncomeCollected: 50_00}, got[1].Records)
	assert.Equal(t, usecase.RecordTotals{IncomeCollected: 200_00, ExpensesPaid: 60_00}, got[2].Records)
	assert.Equal(t, usecase.RecordTotals{}, got[3].Records, "a month without records is zero")
}

// EstimateMonth is EstimateRange of one month: the two never disagree.
func TestEstimateMonthIsTheRangeOfOneMonth(t *testing.T) {
	f, s612 := newRangeFixture()
	est := f.estimator(s612)

	month, err := est.EstimateMonth(t.Context(), userA, taxpayerA, "2027-01")
	require.NoError(t, err)
	window, err := est.EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2026-12"), mustPeriod("2027-01"))
	require.NoError(t, err)

	assert.Equal(t, month, window[1])
}

func TestEstimateRangeRefusesABadWindow(t *testing.T) {
	f, s612 := newRangeFixture()
	est := f.estimator(s612)

	_, err := est.EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2027-02"), mustPeriod("2027-01"))
	require.ErrorIs(t, err, usecase.ErrInvalidRange)

	_, err = est.EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2027-01"), mustPeriod("2027-04"))
	fields := requireCode(t, err, apperr.CodeValidationFailed)
	assert.Equal(t, usecase.DetailOutOfRange, fields[0].Code, "a future month")

	_, err = est.EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2025-12"), mustPeriod("2026-02"))
	fields = requireCode(t, err, apperr.CodeBusinessRuleViolation)
	assert.Equal(t, usecase.DetailParametersUnavailable, fields[0].Code)

	_, err = est.EstimateRange(t.Context(), userA, "not-a-uuid", mustPeriod("2026-12"), mustPeriod("2027-01"))
	requireCode(t, err, apperr.CodeNotFound)
	assert.Empty(t, f.tx.modes, "nothing is read for a refused window")
}

// FU4 (M5-31 G-2): a missing port fails the boot, not the first request.
func TestNewEstimatorValidatesItsDeps(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica)
	registry, err := domain.NewRegistry()
	require.NoError(t, err)
	for name, drop := range map[string]func(*usecase.EstimateDeps){
		"tx":       func(d *usecase.EstimateDeps) { d.Tx = nil },
		"profiles": func(d *usecase.EstimateDeps) { d.Profiles = nil },
		"ledger":   func(d *usecase.EstimateDeps) { d.Ledger = nil },
		"inputs":   func(d *usecase.EstimateDeps) { d.Inputs = nil },
		"texts":    func(d *usecase.EstimateDeps) { d.Texts = nil },
		"holidays": func(d *usecase.EstimateDeps) { d.Holidays = nil },
		"params":   func(d *usecase.EstimateDeps) { d.Params = nil },
		"clock":    func(d *usecase.EstimateDeps) { d.Clock = nil },
	} {
		t.Run(name, func(t *testing.T) {
			deps := f.deps(registry)
			drop(&deps)
			_, err := usecase.NewEstimator(deps)
			require.ErrorIs(t, err, usecase.ErrMissingDependency)
		})
	}
}
