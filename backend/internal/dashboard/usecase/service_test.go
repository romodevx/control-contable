package usecase_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	"github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func cents(c money.Cents) *money.Cents { return &c }

// MODULE-dashboard §3 (OQ-D3, OQ-D7): every combination of the engine and
// the record modules; nothing is computed, nor returned, for a module that
// is OFF.
func TestGetFlagMatrix(t *testing.T) {
	for _, tax := range []bool{true, false} {
		for _, inc := range []bool{true, false} {
			for _, exp := range []bool{true, false} {
				t.Run(fmt.Sprintf("tax=%v incomes=%v expenses=%v", tax, inc, exp), func(t *testing.T) {
					f := newFixture()
					f.flags = fakeFlags{flagTax: tax, flagIncomes: inc, flagExpenses: exp}
					f.estimator.byMonth = map[string]usecase.MonthTax{"2026-10": {IncomeCollected: 500_00, ExpensesPaid: 200_00}}
					f.records.incomes = map[string]money.Cents{"2026-10": 500_00}
					f.records.expenses = map[string]money.Cents{"2026-10": 200_00}

					got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

					require.NoError(t, err)
					require.Len(t, got.Trend, 10)
					assert.Equal(t, tax, got.Estimate != nil, "estimate iff the engine is ON")
					assert.Equal(t, tax, got.Legend != "", "legend iff estimate")
					assert.Equal(t, tax, len(f.estimator.calls) == 1, "the estimator runs iff the engine is ON")
					assert.Equal(t, !tax, len(f.taxpayers.calls) == 1, "without the engine the taxpayer is checked on its own")
					for _, p := range got.Trend {
						assert.Equal(t, tax, p.Tax != nil, "%s tax", p.Month)
						assert.Equal(t, inc, p.Income != nil, "%s income", p.Month)
						assert.Equal(t, exp, p.Expense != nil, "%s expense", p.Month)
					}
					assert.Equal(t, got.Trend[9].Figures, got.Current, "month is the last point")
					if inc {
						assert.Equal(t, cents(500_00), got.Current.Income)
					}
					if exp {
						assert.Equal(t, cents(200_00), got.Current.Expense)
					}
					assert.False(t, got.IsEmpty, "records exist, or no record module is ON")
					for _, call := range f.records.calls {
						assert.False(t, tax, "with the engine ON the figures come from its rows: %s", call)
						if !inc {
							assert.NotContains(t, call, "incomes", "never read an OFF module's records")
						}
						if !exp {
							assert.NotContains(t, call, "expenses", "never read an OFF module's records")
						}
					}
					assert.Equal(t, []string{"ro"}, f.tx.modes, "one read-only unit of work")
					assert.Equal(t, []string{userA + "|" + userA}, f.access.users, "access read in the unit of work")
				})
			}
		}
	}
}

// The window is oldest first with gaps as zeros; the engine's figures and
// statuses land on their months.
func TestGetTrendWithTheEngine(t *testing.T) {
	f := newFixture()
	f.estimator.byMonth = map[string]usecase.MonthTax{
		"2026-02": {
			IVAPayable: 0, ISRPayable: 0, WithholdingsToRemit: 125_00, TotalPayable: 125_00, IVAOutcome: usecase.OutcomeInFavor,
			Regimes: []domain.Status{domain.StatusIncomplete, domain.StatusComplete}, IncomeCollected: 25_000_00, ExpensesPaid: 29_000_00,
		},
		"2026-03": {
			IVAPayable: 2_133_33, ISRPayable: 150_96, TotalPayable: 2_284_29, IVAOutcome: usecase.OutcomePayable,
			Regimes: []domain.Status{domain.StatusComplete}, IncomeCollected: 30_000_00, ExpensesPaid: 10_000_00,
		},
	}

	got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA, Period: "2026-03"})

	require.NoError(t, err)
	assert.Equal(t, []string{userA + "|" + userA + "|" + taxpayerA + "|2026-01..2026-03"}, f.estimator.calls)
	require.Len(t, got.Trend, 3)
	assert.Equal(t, usecase.TrendPoint{
		Month:   domain.NewMonth(2026, time.January),
		Figures: usecase.Figures{Income: cents(0), Expense: cents(0)},
		Tax:     &usecase.TrendTax{IVAOutcome: usecase.OutcomeZero, Status: domain.StatusNone},
	}, got.Trend[0], "a month without records is a point of zeros")
	assert.Equal(t, &usecase.TrendTax{
		WithholdingsToRemit: 125_00, TotalPayable: 125_00, IVAOutcome: usecase.OutcomeInFavor, Status: domain.StatusIncomplete,
	}, got.Trend[1].Tax)
	assert.Equal(t, &usecase.TrendTax{
		IVAPayable: 2_133_33, ISRPayable: 150_96, TotalPayable: 2_284_29, IVAOutcome: usecase.OutcomePayable, Status: domain.StatusComplete,
	}, got.Trend[2].Tax)
	assert.Equal(t, usecase.Figures{Income: cents(30_000_00), Expense: cents(10_000_00)}, got.Current)
	assert.Equal(t, &hero{Period: "2026-03"}, got.Estimate)
	assert.Equal(t, "legend", got.Legend)
	assert.Equal(t, domain.NewMonth(2026, time.March), got.Period)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
}

func TestGetTrendWithoutTheEngine(t *testing.T) {
	f := newFixture()
	f.flags[flagTax] = false
	f.records.incomes = map[string]money.Cents{"2026-02": 25_000_00, "2026-03": 30_000_00}
	f.records.expenses = map[string]money.Cents{"2026-02": 29_000_00}

	got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA, Period: "2026-03", Months: 6})

	require.NoError(t, err)
	assert.Equal(t, []string{"incomes|" + userA + "|2026-01..2026-03", "expenses|" + userA + "|2026-01..2026-03"}, f.records.calls)
	assert.Equal(t, []usecase.TrendPoint{
		{Month: domain.NewMonth(2026, time.January), Figures: usecase.Figures{Income: cents(0), Expense: cents(0)}},
		{Month: domain.NewMonth(2026, time.February), Figures: usecase.Figures{Income: cents(25_000_00), Expense: cents(29_000_00)}},
		{Month: domain.NewMonth(2026, time.March), Figures: usecase.Figures{Income: cents(30_000_00), Expense: cents(0)}},
	}, got.Trend)
	assert.Nil(t, got.Estimate)
	assert.Empty(t, got.Legend)
}

// is_empty: no live record of an enabled module in any period. A record in
// the window answers it without another read.
func TestGetIsEmpty(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		windowIncome          money.Cents
		hasIncomes, hasExpens bool
		want                  bool
		existenceChecks       int
	}{
		{name: "nothing anywhere", want: true, existenceChecks: 2},
		{name: "an income outside the window", hasIncomes: true, existenceChecks: 1},
		{name: "an expense outside the window", hasExpens: true, existenceChecks: 2},
		{name: "a record in the window", windowIncome: 1_00, existenceChecks: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture()
			f.estimator.byMonth = map[string]usecase.MonthTax{"2026-05": {IncomeCollected: tc.windowIncome}}
			f.records.hasIncomes, f.records.hasExpenses = tc.hasIncomes, tc.hasExpens

			got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

			require.NoError(t, err)
			assert.Equal(t, tc.want, got.IsEmpty)
			assert.Len(t, f.records.calls, tc.existenceChecks)
		})
	}
}

// With one record module OFF, only the other one's records count.
func TestGetIsEmptyCountsOnlyEnabledModules(t *testing.T) {
	f := newFixture()
	f.flags[flagIncomes] = false
	f.records.hasIncomes = true

	got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

	require.NoError(t, err)
	assert.True(t, got.IsEmpty, "the incomes module is OFF: its records are invisible")
	assert.Equal(t, []string{"has expenses|" + userA}, f.records.calls)
}

func TestGetAccessSummary(t *testing.T) {
	ends := time.Date(2026, time.October, 17, 20, 0, 0, 0, time.UTC) // today (18:00 UTC) + 3 days 2 hours
	for _, tc := range []struct {
		name    string
		summary access.Summary
		want    usecase.AccessSummary
	}{
		{
			name: "trial",
			summary: access.Summary{
				Status: access.Status{Level: access.LevelTrial, TrialEndsAt: ends}, Usage: access.Usage{Incomes: 3, Expenses: 0},
			},
			want: usecase.AccessSummary{
				Level: access.LevelTrial, TrialEndsAt: &ends, TrialDaysLeft: intPtr(4), IncomesRemaining: intPtr(2), ExpensesRemaining: intPtr(5),
			},
		},
		{
			name:    "read only: counters 0, no trial fields",
			summary: access.Summary{Status: access.Status{Level: access.LevelReadOnly, TrialEndsAt: ends.AddDate(0, 0, -30)}},
			want:    usecase.AccessSummary{Level: access.LevelReadOnly, IncomesRemaining: intPtr(0), ExpensesRemaining: intPtr(0)},
		},
		{name: "pro", summary: access.Summary{Status: access.Status{Level: access.LevelPro}}, want: usecase.AccessSummary{Level: access.LevelPro}},
		{name: "comp", summary: access.Summary{Status: access.Status{Level: access.LevelComp}}, want: usecase.AccessSummary{Level: access.LevelComp}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture()
			f.access.summary = tc.summary

			got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Access)
		})
	}
}

func intPtr(n int) *int { return &n }

// MODULE-dashboard §5: the request's own errors come before any read.
func TestGetRefusesBadInput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		req    usecase.Request
		code   apperr.Code
		detail string
	}{
		{"malformed taxpayer id", usecase.Request{TaxpayerID: "not-a-uuid"}, apperr.CodeNotFound, ""},
		{"malformed period", usecase.Request{TaxpayerID: taxpayerA, Period: "2026-13"}, apperr.CodeValidationFailed, usecase.DetailPattern},
		{"future period", usecase.Request{TaxpayerID: taxpayerA, Period: "2026-11"}, apperr.CodeValidationFailed, usecase.DetailOutOfRange},
		{"before 2026", usecase.Request{TaxpayerID: taxpayerA, Period: "2025-12"}, apperr.CodeBusinessRuleViolation, usecase.DetailParametersUnavailable},
		{"months below 6", usecase.Request{TaxpayerID: taxpayerA, Months: 5}, apperr.CodeValidationFailed, usecase.DetailMinimum},
		{"months above 12", usecase.Request{TaxpayerID: taxpayerA, Months: 13}, apperr.CodeValidationFailed, usecase.DetailMaximum},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture()

			_, err := f.get(t.Context(), tc.req)

			code, fields := appCode(err)
			require.Equal(t, tc.code, code, "%v", err)
			if tc.detail != "" {
				require.Len(t, fields, 1)
				assert.Equal(t, tc.detail, fields[0].Code)
			}
			assert.Empty(t, f.tx.modes, "nothing is read")
		})
	}
}

// No period is the current month in Mexico City: 23:30 on 31 October in
// Mexico City is already November in UTC, still October here.
func TestGetDefaultsToTheCurrentMonth(t *testing.T) {
	f := newFixture()
	f.clock = clock.NewFixed(time.Date(2026, time.November, 1, 5, 30, 0, 0, time.UTC))

	got, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

	require.NoError(t, err)
	assert.Equal(t, "2026-10", got.Period.String())
	assert.Len(t, got.Trend, 10, "default months 12, clipped at 2026-01")
}

// Partial failure is not a mode (§5): any port failing fails the request.
func TestGetFailsWhole(t *testing.T) {
	for name, breakIt := range map[string]func(*fixture){
		"estimator": func(f *fixture) { f.estimator.err = errPortDown },
		"access":    func(f *fixture) { f.access.err = errPortDown },
		"records":   func(f *fixture) { f.flags[flagTax] = false; f.records.err = errPortDown },
		"existence": func(f *fixture) { f.records.err = errPortDown },
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture()
			breakIt(f)

			_, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

			require.ErrorIs(t, err, errPortDown)
		})
	}
}

// A foreign or unknown taxpayer is the ports' 404, in the hidden module's
// words: indistinguishable from module_dashboard OFF (§8, IDOR parity).
func TestGetPassesTheNotFoundOn(t *testing.T) {
	notFound := apperr.New(apperr.CodeNotFound, "The taxpayer was not found.")
	for name, breakIt := range map[string]func(*fixture){
		"engine ON":  func(f *fixture) { f.estimator.err = notFound },
		"engine OFF": func(f *fixture) { f.flags[flagTax] = false; f.taxpayers.err = notFound },
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture()
			breakIt(f)

			_, err := f.get(t.Context(), usecase.Request{TaxpayerID: taxpayerA})

			code, _ := appCode(err)
			assert.Equal(t, apperr.CodeNotFound, code)
			app, _ := errors.AsType[*apperr.Error](err)
			assert.Equal(t, "Not found.", app.Message, "the hidden module's message")
			require.ErrorIs(t, err, notFound, "the cause is kept for the logs")
			assert.Empty(t, f.access.users, "nothing else is read for a foreign taxpayer")
		})
	}
}

// The taxpayer id reaches the ports canonical (lower-case).
func TestGetCanonicalizesTheTaxpayerID(t *testing.T) {
	f := newFixture()
	f.flags[flagTax] = false

	_, err := f.get(t.Context(), usecase.Request{TaxpayerID: "0199D4A2-1111-7000-8000-00000000D00A"})

	require.NoError(t, err)
	assert.Equal(t, []string{userA + "|" + userA + "|" + taxpayerA}, f.taxpayers.calls)
}

func TestNewServiceValidatesItsDeps(t *testing.T) {
	for name, drop := range map[string]func(*usecase.Deps[hero]){
		"tx":        func(d *usecase.Deps[hero]) { d.Tx = nil },
		"flags":     func(d *usecase.Deps[hero]) { d.Flags = nil },
		"taxpayers": func(d *usecase.Deps[hero]) { d.Taxpayers = nil },
		"estimator": func(d *usecase.Deps[hero]) { d.Estimator = nil },
		"records":   func(d *usecase.Deps[hero]) { d.Records = nil },
		"access":    func(d *usecase.Deps[hero]) { d.Access = nil },
		"clock":     func(d *usecase.Deps[hero]) { d.Clock = nil },
	} {
		t.Run(name, func(t *testing.T) {
			deps := newFixture().deps()
			drop(&deps)
			_, err := usecase.NewService(deps)
			require.ErrorIs(t, err, usecase.ErrMissingDependency)
		})
	}
}
