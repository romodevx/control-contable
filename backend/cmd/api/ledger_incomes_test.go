package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// fakeIncomeSums records the query and answers with rows or err.
type fakeIncomeSums struct {
	got  incomesinfra.LedgerQuery
	rows []incomesinfra.IncomeMonthSums
	err  error
}

func (f *fakeIncomeSums) IncomeSums(_ context.Context, q incomesinfra.LedgerQuery) ([]incomesinfra.IncomeMonthSums, error) {
	f.got = q
	return f.rows, f.err
}

func TestIncomeLedger_MapsEveryField(t *testing.T) {
	// Every field a distinct value, so a swapped or dropped field shows up.
	source := &fakeIncomeSums{rows: []incomesinfra.IncomeMonthSums{{
		Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.March, Cash: 1, Accrual: 2,
		ByRate: incomesinfra.IVAByRate{
			Rate16: incomesinfra.RateSums{Base: 3, IVA: 4}, Rate8: incomesinfra.RateSums{Base: 5, IVA: 6},
			Rate0Base: 7, ExemptBase: 8,
		},
		IVAWithheld: incomesinfra.Withheld{ByMoral: 9, ByPlatform: 10},
		ISRWithheld: incomesinfra.Withheld{ByMoral: 11, ByPlatform: 12},
	}}}
	q := taxusecase.LedgerQuery{TaxpayerID: "0199d4a2-1111-7000-8000-00000004f0a1", FirstYear: 2026, FiscalYear: 2027, Through: time.March}

	got, err := newIncomeLedger(source).IncomeSums(t.Context(), q)

	require.NoError(t, err)
	assert.Equal(t, incomesinfra.LedgerQuery{TaxpayerID: q.TaxpayerID, FirstYear: 2026, FiscalYear: 2027, Through: time.March}, source.got)
	assert.Equal(t, []taxusecase.IncomeMonthSums{{
		Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.March,
		Sums: domain.IncomeSums{
			Cash: 1, Accrual: 2,
			ByRate: domain.IVAByRate{
				Rate16: domain.RateSums{Base: 3, IVA: 4}, Rate8: domain.RateSums{Base: 5, IVA: 6},
				Rate0Base: 7, ExemptBase: 8,
			},
			IVAWithheld: domain.Withheld{ByMoral: 9, ByPlatform: 10},
			ISRWithheld: domain.Withheld{ByMoral: 11, ByPlatform: 12},
		},
	}}, got)
}

func TestIncomeLedger_PassesTheReaderErrorOn(t *testing.T) {
	boom := errors.New("boom")
	_, err := newIncomeLedger(&fakeIncomeSums{err: boom}).IncomeSums(t.Context(), taxusecase.LedgerQuery{})
	require.ErrorIs(t, err, boom)
}

func TestIncomeLedger_NoRows(t *testing.T) {
	got, err := newIncomeLedger(&fakeIncomeSums{}).IncomeSums(t.Context(), taxusecase.LedgerQuery{})
	require.NoError(t, err)
	assert.Empty(t, got)
}

// PriorDecember is an expense-only read (the withheld IVA credit): the
// incomes reader has no such range, so asking for it is a wiring bug, never
// silently a different range.
func TestIncomeLedger_RefusesThePriorDecember(t *testing.T) {
	source := &fakeIncomeSums{}
	_, err := newIncomeLedger(source).IncomeSums(t.Context(), taxusecase.LedgerQuery{
		TaxpayerID: "0199d4a2-1111-7000-8000-0000000e000a", FiscalYear: 2027, Through: time.January, PriorDecember: true,
	})
	require.ErrorIs(t, err, errIncomePriorDecember)
	assert.Zero(t, source.got, "nothing is read")
}
