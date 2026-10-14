package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// fakeExpenseSums records the query and answers with rows or err.
type fakeExpenseSums struct {
	got  expensesinfra.LedgerQuery
	rows []expensesinfra.ExpenseMonthSums
	err  error
}

func (f *fakeExpenseSums) ExpenseSums(_ context.Context, q expensesinfra.LedgerQuery) ([]expensesinfra.ExpenseMonthSums, error) {
	f.got = q
	return f.rows, f.err
}

func TestExpenseLedger_MapsEveryField(t *testing.T) {
	// Every field a distinct value, so a swapped or dropped field shows up.
	source := &fakeExpenseSums{rows: []expensesinfra.ExpenseMonthSums{{
		Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.March,
		DeductiblePaid: 1, NonDeductiblePaid: 2, CashOverLimitPaid: 3,
		IVAPaid: expensesinfra.IVAByRate{
			Rate16: expensesinfra.RateSums{Base: 4, IVA: 5}, Rate8: expensesinfra.RateSums{Base: 6, IVA: 7},
			Rate0Base: 8, ExemptBase: 9,
		},
		CreditableIVAPaid: 10, PredialPaid: 11, ISRWithheldToRemit: 12, IVAWithheldToRemit: 13,
	}}}
	q := taxusecase.LedgerQuery{TaxpayerID: "0199d4a2-1111-7000-8000-0000000e000a", FirstYear: 2026, FiscalYear: 2027, Through: time.March}

	got, err := newExpenseLedger(source).ExpenseSums(t.Context(), q)

	require.NoError(t, err)
	assert.Equal(t, expensesinfra.LedgerQuery{TaxpayerID: q.TaxpayerID, FirstYear: 2026, FiscalYear: 2027, Through: time.March}, source.got)
	assert.Equal(t, []taxusecase.ExpenseMonthSums{{
		Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.March,
		Sums: domain.ExpenseSums{
			DeductiblePaid: 1, NonDeductiblePaid: 2, CashOverLimitPaid: 3,
			IVAPaid: domain.IVAByRate{
				Rate16: domain.RateSums{Base: 4, IVA: 5}, Rate8: domain.RateSums{Base: 6, IVA: 7},
				Rate0Base: 8, ExemptBase: 9,
			},
			CreditableIVAPaid: 10, PredialPaid: 11, ISRWithheldToRemit: 12, IVAWithheldToRemit: 13,
		},
	}}, got)
}

func TestExpenseLedger_PassesTheReaderErrorOn(t *testing.T) {
	boom := errors.New("boom")
	_, err := newExpenseLedger(&fakeExpenseSums{err: boom}).ExpenseSums(t.Context(), taxusecase.LedgerQuery{})
	require.ErrorIs(t, err, boom)
}

func TestExpenseLedger_NoRows(t *testing.T) {
	got, err := newExpenseLedger(&fakeExpenseSums{}).ExpenseSums(t.Context(), taxusecase.LedgerQuery{})
	require.NoError(t, err)
	assert.Empty(t, got)
}
