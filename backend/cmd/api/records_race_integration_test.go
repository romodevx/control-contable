//go:build integration

package main

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	expusecase "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// PR #93 review FU-1: ending a regime concurrently with a record write
// (create, restore or update, incomes and expenses) must never leave a live record
// in a period the regime no longer covers. The end path locks the
// taxpayers row, then the taxpayer_regimes row, FOR NO KEY UPDATE before it
// reads the latest record; every write's CheckRegime takes FOR SHARE on the
// same rows in the write's own transaction (the taxpayers row is the one a
// racing write waits on first). These tests stop the end path right after its
// record lookup, start the write, and require it to wait on the lock and
// then read the ended window (422 regime_not_active).

// raceRecordDay is the record's day: October 2026, after the end period.
const (
	raceRecordDay = "2026-10-05"
	// raceInWindowDay is a day the ended window still covers: an update
	// races by moving a record from it to raceRecordDay.
	raceInWindowDay = "2026-09-05"
	raceEndPeriod   = "2026-09"
	// raceAfterEnd is the first day the ended window no longer covers.
	raceAfterEnd = "2026-10-01"
	// raceWaitTimeout bounds every wait of these tests.
	raceWaitTimeout = 15 * time.Second
)

// pausingRecords is the record lookup with a pause after it: the end path
// holds its locks while the test starts the write.
type pausingRecords struct {
	inner   tpusecase.RegimeRecords
	looked  chan struct{}
	release chan struct{}
}

func (p pausingRecords) LatestRecordPeriod(ctx context.Context, taxpayerID, code string) (period tpdomain.Period, err error) {
	period, err = p.inner.LatestRecordPeriod(ctx, taxpayerID, code)
	close(p.looked)
	<-p.release
	return period, err
}

type raceEnv struct {
	app       *sql.DB
	taxpayers *tpusecase.Taxpayers
	regimes   *tpusecase.Regimes
	incomes   *incomes.Incomes
	expenses  *expusecase.Expenses
	records   pausingRecords
}

// newRaceEnv wires the real taxpayers, incomes and expenses use cases over
// one cc_app unit of work, as newRecordModules does, with the pausing
// record lookup.
func newRaceEnv(t *testing.T) *raceEnv {
	t.Helper()
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	require.NoError(t, pgtest.AsUser(t.Context(), app, incUserA, func(tx *sql.Tx) error {
		for _, statement := range incRegisterSQL {
			if _, err := tx.ExecContext(t.Context(), statement, incUserA); err != nil {
				return err
			}
		}
		return nil
	}))
	logger := slog.New(slog.DiscardHandler)
	deps, err := newTaxpayersDeps(taxpayersWiring{env: config.EnvDev, keys: config.Taxpayers{RFCKeyID: 1}, pool: app, logger: logger})
	require.NoError(t, err)
	records := pausingRecords{
		inner:  regimeRecords{expenses: expinfra.NewStore(), incomes: incomesinfra.NewStore()},
		looked: make(chan struct{}), release: make(chan struct{}),
	}
	deps.Records = records
	queries := tpusecase.NewQueries(deps)
	rates, err := newRateCatalog()
	require.NoError(t, err)
	inc, err := newIncomesModule(incomesWiring{
		tx: deps.Tx, taxpayers: queries, rates: rates,
		clock: clock.NewFixed(incToday), accessClock: clock.NewSystem(), logger: logger,
	})
	require.NoError(t, err)
	exp := expusecase.NewExpenses(expusecase.Deps{
		Tx:      deps.Tx,
		Access:  access.NewPolicy(deps.Tx, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Regimes: queries, Taxpayers: taxpayerProfiles{reader: queries},
		Expenses: expinfra.NewStore(), Categories: expinfra.NewCategoryCatalog(),
		Parameters: newExpenseParameters(rates), Clock: clock.NewFixed(incToday), Logger: logger,
	})
	return &raceEnv{
		app: app, taxpayers: tpusecase.NewTaxpayers(deps), regimes: tpusecase.NewRegimes(deps),
		incomes: inc.incomes, expenses: exp, records: records,
	}
}

// taxpayer is a física holding 612 and 606 since January 2026, so 612 can
// end (it is not the last open regime).
func (e *raceEnv) taxpayer(t *testing.T) string {
	t.Helper()
	created, err := e.taxpayers.Create(t.Context(), incUserA, tpusecase.CreateInput{
		PersonType: "fisica", RFC: incRFCFisica, DisplayName: "Carrera",
		Regimes: []tpusecase.RegimeInput{
			{Code: "612", Start: "2026-01"},
			{Code: "606", Start: "2026-01", Options: map[string]tpdomain.OptionValue{"deduction_mode": tpdomain.StringOption("actual")}},
		},
	})
	require.NoError(t, err)
	return created.Taxpayer.ID
}

func (e *raceEnv) income(tp string, id int) incomes.CreateInput {
	return incomes.CreateInput{ID: incID(id), TaxpayerID: tp, Income: incomes.IncomeInput{
		Regime: "612", Counterparty: "publico_general", Amount: 100_00, IVARate: fiscal.IVAExempt,
	}}
}

func (e *raceEnv) expense(t *testing.T, tp string, id int) expusecase.CreateInput {
	return expusecase.CreateInput{
		UserID: incUserA, TaxpayerID: tp, ID: incID(id), Regime: "612",
		Draft: expdomain.Draft{
			Category: "software", Amount: 500_00, IVARate: fiscal.IVARate16, IVA: 80_00,
			PaymentMethod: expdomain.PaymentCardDebit, PaidOn: incDate(t, raceRecordDay), HasCFDI: true,
		},
	}
}

// raceEndAgainst ends 612 at raceEndPeriod while write runs: the end path
// pauses after its record lookup, write starts and must wait on the
// taxpayer and regime row locks; then the end commits. It returns both results.
func (e *raceEnv) raceEndAgainst(t *testing.T, tp string, write func(context.Context) error) (endErr, writeErr error) {
	t.Helper()
	ended := make(chan error, 1)
	go func() {
		_, err := e.regimes.End(context.WithoutCancel(t.Context()), incUserA, tp,
			tpusecase.EndRegimeInput{Code: "612", End: raceEndPeriod, Confirm: true})
		ended <- err
	}()
	select {
	case <-e.records.looked:
	case err := <-ended:
		t.Fatalf("the end path finished before its record lookup paused: %v", err)
	case <-time.After(raceWaitTimeout):
		t.Fatal("the end path never reached its record lookup")
	}

	written := make(chan error, 1)
	go func() { written <- write(context.WithoutCancel(t.Context())) }()
	e.awaitLockWaiter(t, written)
	close(e.records.release)

	select {
	case endErr = <-ended:
	case <-time.After(raceWaitTimeout):
		t.Fatal("the end path never committed")
	}
	select {
	case writeErr = <-written:
	case <-time.After(raceWaitTimeout):
		t.Fatal("the write never finished")
	}
	return endErr, writeErr
}

// awaitLockWaiter waits until a session of this database waits on a row
// lock. A write that finishes first did not wait for the end path: that is
// the race.
func (e *raceEnv) awaitLockWaiter(t *testing.T, written <-chan error) {
	t.Helper()
	deadline := time.Now().Add(raceWaitTimeout)
	for time.Now().Before(deadline) {
		select {
		case err := <-written:
			close(e.records.release)
			t.Fatalf("the write did not wait for the regime end (err: %v): its window check takes no lock on the regime", err)
		default:
		}
		var waiting int
		require.NoError(t, e.app.QueryRowContext(t.Context(),
			`SELECT count(*) FROM pg_stat_activity
			 WHERE datname = current_database() AND wait_event_type = 'Lock'`).Scan(&waiting))
		if waiting > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	close(e.records.release)
	t.Fatal("the write never waited on a lock")
}

// requireRegimeNotActive asserts the write read the ended window.
func requireRegimeNotActive(t *testing.T, err error) {
	t.Helper()
	fields := requireAppCode(t, err, apperr.CodeBusinessRuleViolation)
	require.Len(t, fields, 1)
	assert.Equal(t, "regime_not_active", fields[0].Code)
}

func TestRegimeEndRace_recordWrites(t *testing.T) {
	writes := []struct {
		name string
		// setup runs before the race; write is the racing write.
		setup func(t *testing.T, e *raceEnv, tp string)
		write func(t *testing.T, e *raceEnv, tp string) func(context.Context) error
	}{
		{
			name: "income create",
			write: func(t *testing.T, e *raceEnv, tp string) func(context.Context) error {
				in := e.income(tp, 1)
				in.Income.Dates.CollectedOn = incDate(t, raceRecordDay)
				return func(ctx context.Context) error { _, err := e.incomes.Create(ctx, incUserA, in); return err }
			},
		},
		{
			name: "income restore",
			setup: func(t *testing.T, e *raceEnv, tp string) {
				in := e.income(tp, 1)
				in.Income.Dates.CollectedOn = incDate(t, raceRecordDay)
				_, err := e.incomes.Create(t.Context(), incUserA, in)
				require.NoError(t, err)
				require.NoError(t, e.incomes.Delete(t.Context(), incUserA, tp, incID(1)))
			},
			write: func(_ *testing.T, e *raceEnv, tp string) func(context.Context) error {
				return func(ctx context.Context) error { _, err := e.incomes.Restore(ctx, incUserA, tp, incID(1)); return err }
			},
		},
		{
			// PR #97 review FU-3: an update moving a date past the end goes
			// through the same CheckRegime; pinned so it stays that way.
			name: "income update",
			setup: func(t *testing.T, e *raceEnv, tp string) {
				in := e.income(tp, 1)
				in.Income.Dates.CollectedOn = incDate(t, raceInWindowDay)
				_, err := e.incomes.Create(t.Context(), incUserA, in)
				require.NoError(t, err)
			},
			write: func(t *testing.T, e *raceEnv, tp string) func(context.Context) error {
				moved := incDate(t, raceRecordDay)
				return func(ctx context.Context) error {
					_, err := e.incomes.Update(ctx, incUserA, tp, incID(1), incomes.Patch{CollectedOn: &moved})
					return err
				}
			},
		},
		{
			name: "expense create",
			write: func(t *testing.T, e *raceEnv, tp string) func(context.Context) error {
				in := e.expense(t, tp, 1)
				return func(ctx context.Context) error { _, err := e.expenses.Create(ctx, in); return err }
			},
		},
		{
			name: "expense restore",
			setup: func(t *testing.T, e *raceEnv, tp string) {
				_, err := e.expenses.Create(t.Context(), e.expense(t, tp, 1))
				require.NoError(t, err)
				require.NoError(t, e.expenses.Delete(t.Context(), incUserA, tp, incID(1)))
			},
			write: func(_ *testing.T, e *raceEnv, tp string) func(context.Context) error {
				return func(ctx context.Context) error { _, err := e.expenses.Restore(ctx, incUserA, tp, incID(1)); return err }
			},
		},
		{
			name: "expense update",
			setup: func(t *testing.T, e *raceEnv, tp string) {
				in := e.expense(t, tp, 1)
				in.Draft.PaidOn = incDate(t, raceInWindowDay)
				_, err := e.expenses.Create(t.Context(), in)
				require.NoError(t, err)
			},
			write: func(t *testing.T, e *raceEnv, tp string) func(context.Context) error {
				moved := incDate(t, raceRecordDay)
				in := expusecase.UpdateInput{
					UserID: incUserA, TaxpayerID: tp, ExpenseID: incID(1), Patch: expusecase.Patch{PaidOn: &moved},
				}
				return func(ctx context.Context) error { _, err := e.expenses.Update(ctx, in); return err }
			},
		},
	}
	for _, w := range writes {
		t.Run(w.name, func(t *testing.T) {
			e := newRaceEnv(t)
			tp := e.taxpayer(t)
			if w.setup != nil {
				w.setup(t, e, tp)
			}

			endErr, writeErr := e.raceEndAgainst(t, tp, w.write(t, e, tp))

			require.NoError(t, endErr, "the end path checked first: it wins")
			requireRegimeNotActive(t, writeErr)
			assert.Zero(t, e.liveRecordsAfterEnd(t, tp), "no live record outside the ended window")
		})
	}
}

// liveRecordsAfterEnd counts the taxpayer's live 612 records after the end
// period, as the user.
func (e *raceEnv) liveRecordsAfterEnd(t *testing.T, tp string) int {
	t.Helper()
	var n int
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, incUserA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), `
			SELECT (SELECT count(*) FROM incomes WHERE taxpayer_id = $1 AND regime_code = '612'
			          AND deleted_at IS NULL AND greatest(invoiced_on, collected_on) >= $2::date)
			     + (SELECT count(*) FROM expenses WHERE taxpayer_id = $1 AND regime_code = '612'
			          AND deleted_at IS NULL AND paid_on >= $2::date)`, tp, raceAfterEnd).Scan(&n)
	}))
	return n
}

// M2-14 G-1 for incomes: once an income holds the regime in October, ending
// it in September is refused (records_after_end_period), as for expenses.
func TestRegimeEnd_refusesAnEndBeforeTheLatestIncome(t *testing.T) {
	e := newRaceEnv(t)
	close(e.records.release) // no pause: a plain end
	tp := e.taxpayer(t)
	in := e.income(tp, 1)
	in.Income.Dates.CollectedOn = incDate(t, raceRecordDay)
	_, err := e.incomes.Create(t.Context(), incUserA, in)
	require.NoError(t, err)

	_, err = e.regimes.End(t.Context(), incUserA, tp, tpusecase.EndRegimeInput{Code: "612", End: raceEndPeriod, Confirm: true})

	fields := requireAppCode(t, err, apperr.CodeBusinessRuleViolation)
	require.Len(t, fields, 1)
	assert.Equal(t, "records_after_end_period", fields[0].Code)
}
