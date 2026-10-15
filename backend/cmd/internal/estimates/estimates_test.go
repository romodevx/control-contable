package estimates

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

type noopTx struct{}

func (noopTx) ReadWrite(ctx context.Context, _ string, work tx.Work) error { return work(ctx) }
func (noopTx) ReadOnly(ctx context.Context, _ string, work tx.Work) error  { return work(ctx) }

type noHolidays struct{}

func (noHolidays) Holidays(context.Context) (fiscal.HolidaySet, error) {
	return fiscal.NewHolidaySet(nil, nil)
}

// The monthly Estimator every composition root builds the same way: every
// strategy, the embedded parameters, the shared ledger readers. A missing
// port fails the boot.
func TestNewMonthly(t *testing.T) {
	deps := MonthlyDeps{
		Tx: noopTx{}, Taxpayers: fakeProfiles{}, Holidays: noHolidays{},
		Clock: clock.NewFixed(time.Date(2026, time.May, 14, 10, 0, 0, 0, clock.Location())),
	}
	estimator, err := NewMonthly(deps)
	require.NoError(t, err)
	assert.NotNil(t, estimator)

	for name, mutate := range map[string]func(*MonthlyDeps){
		"Tx":        func(d *MonthlyDeps) { d.Tx = nil },
		"Taxpayers": func(d *MonthlyDeps) { d.Taxpayers = nil },
		"Holidays":  func(d *MonthlyDeps) { d.Holidays = nil },
		"Clock":     func(d *MonthlyDeps) { d.Clock = nil },
	} {
		t.Run(name, func(t *testing.T) {
			d := deps
			mutate(&d)
			_, err := NewMonthly(d)
			require.ErrorIs(t, err, taxusecase.ErrMissingDependency)
		})
	}
}

func TestNewLedgerServesBothTables(t *testing.T) {
	var _ taxusecase.LedgerReader = NewLedger()
}
