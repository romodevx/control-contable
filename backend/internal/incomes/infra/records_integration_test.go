//go:build integration

package infra_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// latestRecordDate reads the store's latest live income date of the regime
// as user, in a unit of work like the taxpayers module's.
func (e *env) latestRecordDate(t *testing.T, user, taxpayerID string, regime fiscal.RegimeCode) fiscal.Date {
	t.Helper()
	var got fiscal.Date
	require.NoError(t, e.txm.ReadOnly(t.Context(), user, func(ctx context.Context) error {
		var err error
		got, err = infra.NewStore().LatestRecordDate(ctx, taxpayerID, regime)
		return err
	}))
	return got
}

// MODULE-taxpayers §2.6 / M2-14 G-1: ending a regime needs the latest day a
// live income of it holds the regime on, the later of its two dates
// (CheckRegime checks both). A deleted income holds nothing (its restore
// re-checks the window); another regime's, another taxpayer's or another
// user's neither.
func TestLatestRecordDate_IsTheLatestDateOfALiveIncomeOfTheRegime(t *testing.T) {
	e := newEnv(t)
	pm := e.taxpayer(t, userA, "moral", "A&B020304XY2", "601")
	other := e.taxpayer(t, userA, "moral", "J&M120615R79", "601")
	assert.True(t, e.latestRecordDate(t, userA, pm, "601").IsZero(), "no income yet")

	accrual := func(n int, invoiced, collected string) usecase.NewIncome {
		f := usecase.Fields{
			Regime: "601", Counterparty: domain.CounterpartyPublic, Amount: cents(100_00), IVARate: fiscal.IVAExempt,
			Dates: domain.Dates{InvoicedOn: day(t, invoiced)},
		}
		if collected != "" {
			f.Dates.CollectedOn = day(t, collected)
		}
		return usecase.NewIncome{ID: id(n), UserID: userA, TaxpayerID: pm, Fields: f}
	}
	require.NoError(t, e.as(t, userA, func(ctx context.Context) error {
		for _, in := range []usecase.NewIncome{
			accrual(1, "2026-03-10", ""),
			accrual(2, "2026-05-02", "2026-09-20"), // booked in May, collected in September
			accrual(3, "2026-08-31", ""),
		} {
			if err := e.store.Insert(ctx, in); err != nil {
				return err
			}
		}
		return nil
	}))
	assert.Equal(t, "2026-09-20", e.latestRecordDate(t, userA, pm, "601").String(), "the collection date holds the regime too")

	require.NoError(t, e.as(t, userA, func(ctx context.Context) error { return e.store.SetDeleted(ctx, id(2), true) }))
	assert.Equal(t, "2026-08-31", e.latestRecordDate(t, userA, pm, "601").String(), "a deleted income holds nothing")

	assert.True(t, e.latestRecordDate(t, userA, pm, "612").IsZero(), "another regime")
	assert.True(t, e.latestRecordDate(t, userA, other, "601").IsZero(), "another taxpayer")
	assert.True(t, e.latestRecordDate(t, userB, pm, "601").IsZero(), "RLS: another user sees nothing")
}

func TestLatestRecordDate_NeedsAUnitOfWork(t *testing.T) {
	_, err := infra.NewStore().LatestRecordDate(t.Context(), id(1), "601")

	require.Error(t, err)
}
