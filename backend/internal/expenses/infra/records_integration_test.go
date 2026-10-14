//go:build integration

package infra_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// latestPaidOn reads the store's latest live paid_on of the regime as
// userID, in a unit of work like the taxpayers module's.
func (e *env) latestPaidOn(t *testing.T, userID, taxpayerID string, regime fiscal.RegimeCode) fiscal.Date {
	t.Helper()
	var got fiscal.Date
	require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
		var err error
		got, err = infra.NewStore().LatestPaidOn(ctx, taxpayerID, regime)
		return err
	}))
	return got
}

// MODULE-taxpayers §2.6 / M2-14 G-1: ending a regime needs the latest
// period holding a live expense of it; a deleted expense does not hold the
// regime (its restore re-checks the window), another regime's or another
// user's expense neither.
func TestLatestPaidOn_IsTheLatestLiveExpenseOfTheRegime(t *testing.T) {
	e := newEnv(t)
	assert.True(t, e.latestPaidOn(t, userA, taxpayerA, "612").IsZero(), "no expense yet")

	for n, on := range map[int]string{1: "2026-03-10", 2: "2026-08-31", 3: "2026-09-20"} {
		_, err := e.uc.Create(t.Context(), software(t, n, on))
		require.NoError(t, err)
	}
	assert.Equal(t, "2026-09-20", e.latestPaidOn(t, userA, taxpayerA, "612").String())

	require.NoError(t, e.uc.Delete(t.Context(), userA, taxpayerA, expenseID(3)))
	assert.Equal(t, "2026-08-31", e.latestPaidOn(t, userA, taxpayerA, "612").String(), "a deleted expense holds nothing")

	assert.True(t, e.latestPaidOn(t, userA, taxpayerA, "606").IsZero(), "another regime of the taxpayer")
	assert.True(t, e.latestPaidOn(t, userA, taxpayerOld, "612").IsZero(), "another taxpayer")
	assert.True(t, e.latestPaidOn(t, userB, taxpayerA, "612").IsZero(), "RLS: another user sees nothing")
}

func TestLatestPaidOn_NeedsAUnitOfWork(t *testing.T) {
	_, err := infra.NewStore().LatestPaidOn(t.Context(), taxpayerA, "612")

	require.Error(t, err)
}
