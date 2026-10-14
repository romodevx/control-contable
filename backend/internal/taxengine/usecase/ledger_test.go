package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

var marchQuery = usecase.LedgerQuery{TaxpayerID: "tp", FiscalYear: 2026, Through: time.March}

func expense(regime fiscal.RegimeCode, month time.Month, deductible money.Cents) usecase.ExpenseMonthSums {
	return usecase.ExpenseMonthSums{Regime: regime, Month: month, Sums: domain.ExpenseSums{DeductiblePaid: deductible}}
}

func income(regime fiscal.RegimeCode, month time.Month, cash money.Cents) usecase.IncomeMonthSums {
	return usecase.IncomeMonthSums{Regime: regime, Month: month, Sums: domain.IncomeSums{Cash: cash}}
}

func TestYearSumsByRegime(t *testing.T) {
	years, err := usecase.YearSumsByRegime(marchQuery,
		[]usecase.IncomeMonthSums{income(fiscal.RegimeBusiness, time.January, 1000), income(fiscal.RegimeBusiness, time.March, 3000)},
		[]usecase.ExpenseMonthSums{
			expense(fiscal.RegimeBusiness, time.January, 100), expense(fiscal.RegimeBusiness, time.February, 200),
			expense(fiscal.RegimeRent, time.March, 50),
		},
	)
	require.NoError(t, err)
	require.Len(t, years, 2)

	business := years[fiscal.RegimeBusiness]
	assert.Equal(t, time.March, business.Through())
	assert.Equal(t, domain.MonthSums{
		Income: domain.IncomeSums{Cash: 1000}, Expenses: domain.ExpenseSums{DeductiblePaid: 100},
	}, business.Month(time.January), "income and expense rows of a month meet")
	assert.Equal(t, domain.MonthSums{Expenses: domain.ExpenseSums{DeductiblePaid: 200}}, business.Month(time.February))
	ytd, err := business.Sum(time.January, marchQuery.Through)
	require.NoError(t, err)
	assert.Equal(t, domain.MonthSums{
		Income: domain.IncomeSums{Cash: 4000}, Expenses: domain.ExpenseSums{DeductiblePaid: 300},
	}, ytd, "year to date from January")

	rent := years[fiscal.RegimeRent]
	assert.Equal(t, time.March, rent.Through())
	assert.Equal(t, domain.MonthSums{}, rent.Month(time.January), "a month without records is zero")
	assert.Equal(t, domain.MonthSums{Expenses: domain.ExpenseSums{DeductiblePaid: 50}}, rent.Month(time.March))
}

// The income reader's rows (M3-07) fold the same way: an accrual regime accumulates its ISR
// income from January, while its IVA is read for the period's month only (LIVA 5-D).
func TestYearSumsByRegime_AccrualIncomeYTDAndMonthlyIVA(t *testing.T) {
	moral := fiscal.RegimeCode("601")
	row := func(month time.Month, accrual, cash, base, iva money.Cents) usecase.IncomeMonthSums {
		return usecase.IncomeMonthSums{Regime: moral, Month: month, Sums: domain.IncomeSums{
			Accrual: accrual, Cash: cash, ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: base, IVA: iva}},
			ISRWithheld: domain.Withheld{ByPlatform: cash / 40},
		}}
	}
	years, err := usecase.YearSumsByRegime(marchQuery,
		[]usecase.IncomeMonthSums{row(time.January, 0, 3000_00, 3000_00, 480_00), row(time.March, 15000_00, 1000_00, 1000_00, 160_00)},
		nil,
	)
	require.NoError(t, err)

	ytd, err := years[moral].Sum(time.January, marchQuery.Through)
	require.NoError(t, err)
	assert.Equal(t, money.Cents(15000_00), ytd.Income.Accrual)
	assert.Equal(t, money.Cents(4000_00), ytd.Income.Cash)
	assert.Equal(t, money.Cents(100_00), ytd.Income.ISRWithheld.ByPlatform)
	assert.Equal(t, domain.RateSums{Base: 1000_00, IVA: 160_00}, years[moral].Month(time.March).Income.ByRate.Rate16,
		"IVA is never year to date")
}

func TestYearSumsByRegime_NoRows(t *testing.T) {
	years, err := usecase.YearSumsByRegime(marchQuery, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, years, "a regime without records is absent: the zero YearSums")
}

func TestYearSumsByRegime_Rejects(t *testing.T) {
	cases := map[string]struct {
		q        usecase.LedgerQuery
		incomes  []usecase.IncomeMonthSums
		expenses []usecase.ExpenseMonthSums
	}{
		"expense month after the period": {marchQuery, nil, []usecase.ExpenseMonthSums{expense(fiscal.RegimeBusiness, time.April, 1)}},
		"income month after the period":  {marchQuery, []usecase.IncomeMonthSums{income(fiscal.RegimeBusiness, time.April, 1)}, nil},
		"month zero":                     {marchQuery, nil, []usecase.ExpenseMonthSums{expense(fiscal.RegimeBusiness, 0, 1)}},
		"repeated expense month": {marchQuery, nil, []usecase.ExpenseMonthSums{
			expense(fiscal.RegimeBusiness, time.January, 1), expense(fiscal.RegimeBusiness, time.January, 2),
		}},
		"repeated income month": {marchQuery, []usecase.IncomeMonthSums{
			income(fiscal.RegimeBusiness, time.January, 1), income(fiscal.RegimeBusiness, time.January, 2),
		}, nil},
		"through out of range": {usecase.LedgerQuery{Through: 13}, nil, nil},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := usecase.YearSumsByRegime(c.q, c.incomes, c.expenses)
			require.ErrorIs(t, err, domain.ErrInvalidMonth)
		})
	}
}
