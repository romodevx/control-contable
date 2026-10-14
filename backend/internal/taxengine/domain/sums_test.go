package domain_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// sums fills every field with v, so a field left out of Add shows up.
func sums(v money.Cents) domain.MonthSums {
	return domain.MonthSums{
		Income: domain.IncomeSums{
			Cash: v, Accrual: v,
			ByRate: domain.IVAByRate{
				Rate16: domain.RateSums{Base: v, IVA: v}, Rate8: domain.RateSums{Base: v, IVA: v},
				Rate0Base: v, ExemptBase: v,
			},
			IVAWithheld: domain.Withheld{ByMoral: v, ByPlatform: v},
			ISRWithheld: domain.Withheld{ByMoral: v, ByPlatform: v},
		},
		Expenses: domain.ExpenseSums{
			DeductiblePaid: v, NonDeductiblePaid: v, CashOverLimitPaid: v,
			IVAPaid: domain.IVAByRate{
				Rate16: domain.RateSums{Base: v, IVA: v}, Rate8: domain.RateSums{Base: v, IVA: v},
				Rate0Base: v, ExemptBase: v,
			},
			CreditableIVAPaid: v, PredialPaid: v,
			ISRWithheldToRemit: v, IVAWithheldToRemit: v,
		},
	}
}

func TestYearSums(t *testing.T) {
	y, err := domain.NewYearSums(time.March,
		domain.MonthEntry{Month: time.January, Sums: sums(1)},
		domain.MonthEntry{Month: time.March, Sums: sums(10)},
	)
	require.NoError(t, err)
	assert.Equal(t, time.March, y.Through())
	assert.Equal(t, sums(1), y.Month(time.January))
	assert.Equal(t, domain.MonthSums{}, y.Month(time.February), "a month without records")
	assert.Equal(t, domain.MonthSums{}, y.Month(time.April), "after the period")
	assert.Equal(t, domain.MonthSums{}, y.Month(0))

	ytd, err := y.Sum(time.January, time.March)
	require.NoError(t, err)
	assert.Equal(t, sums(11), ytd)
	bimester, err := y.Sum(time.February, time.March)
	require.NoError(t, err)
	assert.Equal(t, sums(10), bimester)

	for _, r := range [][2]time.Month{{0, time.March}, {time.February, time.January}, {time.January, time.April}} {
		_, err := y.Sum(r[0], r[1])
		require.ErrorIs(t, err, domain.ErrInvalidMonth, "%v", r)
	}
}

func TestNewYearSumsRejects(t *testing.T) {
	tests := []struct {
		name    string
		through time.Month
		entries []domain.MonthEntry
	}{
		{name: "through zero", through: 0},
		{name: "through 13", through: 13},
		{name: "month zero", through: time.March, entries: []domain.MonthEntry{{Month: 0}}},
		{name: "after through", through: time.March, entries: []domain.MonthEntry{{Month: time.April}}},
		{name: "repeated", through: time.March, entries: []domain.MonthEntry{{Month: time.March}, {Month: time.March}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewYearSums(tt.through, tt.entries...)
			require.ErrorIs(t, err, domain.ErrInvalidMonth)
		})
	}
}

func TestMonthSumsAddOverflows(t *testing.T) {
	for _, field := range []func(*domain.MonthSums){
		func(s *domain.MonthSums) { s.Income.Cash = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Income.ISRWithheld.ByPlatform = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Income.ByRate.Rate8.IVA = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Income.ByRate.ExemptBase = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Expenses.IVAWithheldToRemit = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Expenses.NonDeductiblePaid = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Expenses.CashOverLimitPaid = math.MaxInt64 },
		func(s *domain.MonthSums) { s.Expenses.IVAPaid.Rate8.IVA = math.MaxInt64 },
	} {
		big := domain.MonthSums{}
		field(&big)
		_, err := big.Add(sums(1))
		require.ErrorIs(t, err, money.ErrOverflow)
	}
	y, err := domain.NewYearSums(time.February,
		domain.MonthEntry{Month: time.January, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: math.MaxInt64}}},
		domain.MonthEntry{Month: time.February, Sums: sums(1)},
	)
	require.NoError(t, err)
	_, err = y.Sum(time.January, time.February)
	require.ErrorIs(t, err, money.ErrOverflow)
}
