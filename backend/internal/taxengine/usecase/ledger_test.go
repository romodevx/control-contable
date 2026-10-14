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
