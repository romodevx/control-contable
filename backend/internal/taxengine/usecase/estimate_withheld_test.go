package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// LIVA 5 fr. IV (MODULE-withholdings §4.2 item 2): the IVA a persona moral
// withheld from a supplier is credited in the next month's return, so
// January credits the previous December's and the read reaches one month
// into the previous fiscal year. The fixtures are prefixed wh.

// whWithheld is a 601 expense row whose withheld IVA is creditable.
func whWithheld(year int, month time.Month, withheld money.Cents) usecase.ExpenseMonthSums {
	return usecase.ExpenseMonthSums{Regime: "601", Year: year, Month: month, Sums: domain.ExpenseSums{
		DeductiblePaid: 1000_00, CreditableIVAWithheld: withheld,
	}}
}

// whFixture is a 601 persona moral held since January 2026, today 10 March
// 2027, with expenses.
func whFixture(expenses ...usecase.ExpenseMonthSums) (*estFixture, *estStrategy) {
	f := newEstFixture(fiscal.PersonMoral, estHold("601", "2026-01", domain.TreatmentFull))
	f.clock = clock.NewFixed(rangeToday)
	f.ledger.expenses = expenses
	return f, &estStrategy{key: est601, treatment: domain.TreatmentFull}
}

// lastRun is the input the strategy last ran on for period.
func lastRun(t *testing.T, s *estStrategy, period string) domain.Input {
	t.Helper()
	var in domain.Input
	found := false
	for _, ran := range s.ran {
		if ran.Period.String() == period {
			in, found = ran, true
		}
	}
	require.True(t, found, "the strategy never ran on %s", period)
	return in
}

func TestEstimateJanuaryCreditsThePreviousDecembersWithheldIVA(t *testing.T) {
	f, s601 := whFixture(whWithheld(2026, time.December, 5333_33), whWithheld(2027, time.January, 1_00))

	got, err := f.estimator(s601).EstimateMonth(t.Context(), userA, taxpayerA, "2027-01")

	require.NoError(t, err)
	jan := lastRun(t, s601, "2027-01").Sums
	assert.Equal(t, money.Cents(5333_33), jan.CreditableWithheldPriorMonth(time.January))
	assert.Equal(t, money.Cents(1_00), jan.Month(time.January).Expenses.CreditableIVAWithheld, "January's own waits for February")
	assert.Equal(t, money.Cents(1000_00), jan.Month(time.January).Expenses.DeductiblePaid, "December never enters 2027's sums")
	assert.Equal(t, usecase.RecordTotals{ExpensesPaid: 1000_00}, got.Records, "the records are January's only")

	want := usecase.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2027, Through: time.January}
	assert.Equal(t, []usecase.LedgerQuery{want}, f.ledger.queries, "incomes read the fiscal year only")
	want.PriorDecember = true
	assert.Equal(t, []usecase.LedgerQuery{want}, f.ledger.expenseQueries, "expenses reach one month back, same statement")
}

// A regime whose only row is the previous December still credits it.
func TestEstimateJanuaryCreditsDecemberWithoutJanuaryRecords(t *testing.T) {
	f, s601 := whFixture(whWithheld(2026, time.December, 400_00))

	_, err := f.estimator(s601).EstimateMonth(t.Context(), userA, taxpayerA, "2027-01")

	require.NoError(t, err)
	assert.Equal(t, money.Cents(400_00), lastRun(t, s601, "2027-01").Sums.CreditableWithheldPriorMonth(time.January))
}

// 2026 is the first fiscal year with parameters, and no record is dated
// before it (BusinessDate): January 2026 reads no previous December.
func TestEstimateJanuary2026ReadsNoPreviousDecember(t *testing.T) {
	f, s601 := whFixture()
	f.clock = clock.NewFixed(today)

	_, err := f.estimator(s601).EstimateMonth(t.Context(), userA, taxpayerA, "2026-01")

	require.NoError(t, err)
	assert.Equal(t, []usecase.LedgerQuery{{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.January}},
		f.ledger.expenseQueries)
}

// A window across the fiscal years credits December 2026 in January 2027
// from the rows it already read; each month equals EstimateMonth's.
func TestEstimateRangeCreditsDecemberInJanuaryAcrossTheYears(t *testing.T) {
	f, s601 := whFixture(whWithheld(2026, time.November, 7_00), whWithheld(2026, time.December, 9_00))
	est := f.estimator(s601)

	window, err := est.EstimateRange(t.Context(), userA, taxpayerA, mustPeriod("2026-12"), mustPeriod("2027-01"))
	require.NoError(t, err)
	assert.Equal(t, money.Cents(9_00), lastRun(t, s601, "2027-01").Sums.CreditableWithheldPriorMonth(time.January))
	assert.Equal(t, money.Cents(7_00), lastRun(t, s601, "2026-12").Sums.CreditableWithheldPriorMonth(time.December))

	month, err := est.EstimateMonth(t.Context(), userA, taxpayerA, "2027-01")
	require.NoError(t, err)
	assert.Equal(t, month, window[1])
}

// The previous December is part of January's input, so its fingerprint
// moves with it (M8 marks the month as changed).
func TestEstimateFingerprintCoversThePreviousDecember(t *testing.T) {
	fingerprint := func(december money.Cents) string {
		t.Helper()
		f, s601 := whFixture(whWithheld(2026, time.December, december))
		got, err := f.estimator(s601).EstimateMonth(t.Context(), userA, taxpayerA, "2027-01")
		require.NoError(t, err)
		return got.RecordsFingerprint
	}

	assert.NotEqual(t, fingerprint(100_00), fingerprint(100_01))
}
