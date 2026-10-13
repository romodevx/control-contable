//go:build integration

package accessgorm_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The AccessPolicy against the real identity migration, as cc_app under RLS
// (P0-16). Pro rows are written as cc_webhook and comp rows as cc_owner, the
// only roles their policies let write them.

const (
	userA = "00000000-0000-4000-8000-00000000000a"
	userB = "00000000-0000-4000-8000-00000000000b"

	// racers is how many goroutines fight over one slot.
	racers = 24
)

// registerSQL is sign-up as cc_app does it under the new user's context: the user,
// its zeroed counters and its trial (starts now(), ends 7 days later).
var registerSQL = []string{
	`WITH v(id) AS (SELECT $1::uuid)
	 INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
	 SELECT id, sha256(convert_to(id::text, 'UTF8')),
	        '\x0001'::bytea || sha256(convert_to(id::text, 'UTF8')) || sha256(convert_to(id::text, 'UTF8')),
	        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana'
	 FROM v`,
	"INSERT INTO user_usage (user_id) VALUES ($1)",
	"INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '168 hours')",
}

const (
	insertProSQL = "INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at, rc_product_id, rc_original_transaction_id) " +
		"VALUES ($1, 'pro', 'revenuecat', now() - interval '1 day', now() + interval '30 days', 'cc_pro_monthly', 'GPA.0001')"
	insertCompSQL = "INSERT INTO entitlements (user_id, kind, source, starts_at) VALUES ($1, 'comp', 'admin', now() - interval '1 day')"
	// setUsageSQL is a test shortcut to an arbitrary counter (as cc_app, own row).
	setUsageSQL = "UPDATE user_usage SET incomes_created = $2, expenses_created = $3 WHERE user_id = $1"
)

type env struct {
	database *pgtest.Database
	app      *sql.DB
	txm      *pgdb.TxManager
	store    accessgorm.Store
}

func newEnv(t *testing.T) env {
	t.Helper()
	database := pgtest.NewDatabase(t)
	app := database.Connect(t, pgtest.RoleApp)
	txm, err := pgdb.NewTxManager(app)
	require.NoError(t, err)
	return env{database: database, app: app, txm: txm, store: accessgorm.NewStore()}
}

func (e env) policy(clk clock.Clock) *access.Policy {
	return access.NewPolicy(e.txm, e.store, e.store, clk)
}

// systemClock is the production clock: the trial rows start at the database's now().
func systemClock() clock.Clock { return clock.NewSystem() }

func execAs(t *testing.T, db *sql.DB, userID string, statements []string, args ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), db, userID, func(tx *sql.Tx) error {
		for _, stmt := range statements {
			if _, err := tx.ExecContext(t.Context(), stmt, append([]any{userID}, args...)...); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (e env) register(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		execAs(t, e.app, id, registerSQL)
	}
}

func (e env) setUsage(t *testing.T, userID string, incomes, expenses int) {
	t.Helper()
	execAs(t, e.app, userID, []string{setUsageSQL}, incomes, expenses)
}

func (e env) usage(t *testing.T, userID string) access.Usage {
	t.Helper()
	var usage access.Usage
	require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
		var err error
		usage, err = e.store.Usage(ctx, userID)
		return err
	}))
	return usage
}

func (e env) trialEnd(t *testing.T, userID string) time.Time {
	t.Helper()
	var ents []access.Entitlement
	require.NoError(t, e.txm.ReadOnly(t.Context(), userID, func(ctx context.Context) error {
		var err error
		ents, err = e.store.Entitlements(ctx, userID)
		return err
	}))
	for _, ent := range ents {
		if ent.Kind == access.EntitlementTrial {
			return ent.EndsAt
		}
	}
	t.Fatalf("no trial for %s", userID)
	return time.Time{}
}

func assertEntitlementRequired(t *testing.T, err error, target error) {
	t.Helper()
	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %v", err)
	assert.Equal(t, apperr.CodeEntitlementRequired, appErr.Code)
	assert.ErrorIs(t, err, target)
}

func TestStore_ReadsEntitlementsOfEveryKind(t *testing.T) {
	e := newEnv(t)
	e.register(t, userA)
	execAs(t, e.database.Connect(t, pgtest.RoleWebhook), userA, []string{insertProSQL})
	execAs(t, e.database.Connect(t, pgtest.RoleOwner), userA, []string{insertCompSQL})

	var ents []access.Entitlement
	require.NoError(t, e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		var err error
		ents, err = e.store.Entitlements(ctx, userA)
		return err
	}))

	byKind := map[access.EntitlementKind]access.Entitlement{}
	for _, ent := range ents {
		byKind[ent.Kind] = ent
	}
	require.Len(t, byKind, 3)
	trial := byKind[access.EntitlementTrial]
	assert.Equal(t, access.TrialDuration, trial.EndsAt.Sub(trial.StartsAt))
	assert.False(t, byKind[access.EntitlementPro].EndsAt.IsZero())
	assert.True(t, byKind[access.EntitlementComp].EndsAt.IsZero(), "comp is open-ended")
}

func TestReserveRecord_TrialCapFifthVersusSixth(t *testing.T) {
	e := newEnv(t)
	e.register(t, userA)
	policy := e.policy(systemClock())

	for i := range access.TrialRecordCap {
		require.NoError(t, policy.ReserveRecord(t.Context(), userA, access.RecordIncome), "income %d", i+1)
	}
	err := policy.ReserveRecord(t.Context(), userA, access.RecordIncome)

	assertEntitlementRequired(t, err, access.ErrTrialCapReached)
	assert.Equal(t, access.Usage{Incomes: access.TrialRecordCap}, e.usage(t, userA))
	require.NoError(t, policy.ReserveRecord(t.Context(), userA, access.RecordExpense), "caps are per kind")
}

func TestReserveRecord_ConcurrentRacersForTheLastSlot(t *testing.T) {
	tests := []struct {
		name     string
		start    int
		wantWins int
	}{
		{"one slot left: exactly one wins", access.TrialRecordCap - 1, 1},
		{"fresh trial: exactly the cap wins", 0, access.TrialRecordCap},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.register(t, userA)
			e.setUsage(t, userA, tt.start, 0)
			policy := e.policy(systemClock())

			wins, refusals, others := race(t, racers, func(ctx context.Context) error {
				return policy.ReserveRecord(ctx, userA, access.RecordIncome)
			})

			assert.Equal(t, tt.wantWins, wins)
			assert.Equal(t, racers-tt.wantWins, refusals)
			assert.Empty(t, others)
			assert.Equal(t, access.TrialRecordCap, e.usage(t, userA).Incomes, "never overshoots")
		})
	}
}

// race starts n goroutines on a common signal and sorts their outcomes into
// successes, trial-cap refusals and anything else.
func race(t *testing.T, n int, reserve func(context.Context) error) (int, int, []error) {
	t.Helper()
	start := make(chan struct{})
	results := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-start
			results[i] = reserve(t.Context())
		})
	}
	close(start)
	wg.Wait()
	wins, refusals := 0, 0
	var others []error
	for _, err := range results {
		switch {
		case err == nil:
			wins++
		case errors.Is(err, access.ErrTrialCapReached):
			refusals++
		default:
			others = append(others, err)
		}
	}
	return wins, refusals, others
}

func TestReserveRecord_RollbackReleasesTheSlot(t *testing.T) {
	e := newEnv(t)
	e.register(t, userA)
	e.setUsage(t, userA, access.TrialRecordCap-1, 0)
	policy := e.policy(systemClock())
	insertFailed := errors.New("insert income failed")

	err := e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
		if err := policy.ReserveRecord(ctx, userA, access.RecordIncome); err != nil {
			return err
		}
		return insertFailed
	})

	require.ErrorIs(t, err, insertFailed)
	assert.Equal(t, access.TrialRecordCap-1, e.usage(t, userA).Incomes, "slot given back")
	require.NoError(t, policy.ReserveRecord(t.Context(), userA, access.RecordIncome), "the 5th slot is still free")
}

func TestReserveRecord_TrialClockBoundary(t *testing.T) {
	e := newEnv(t)
	e.register(t, userA)
	end := e.trialEnd(t, userA)

	require.NoError(t, e.policy(clock.NewFixed(end.Add(-time.Microsecond))).
		ReserveRecord(t.Context(), userA, access.RecordIncome), "last instant of the trial")
	err := e.policy(clock.NewFixed(end)).ReserveRecord(t.Context(), userA, access.RecordIncome)

	assertEntitlementRequired(t, err, access.ErrReadOnly)
	assert.Equal(t, 1, e.usage(t, userA).Incomes, "a refused reservation writes nothing")
}

func TestReserveRecord_PaidAccessIsUnlimited(t *testing.T) {
	tests := []struct {
		name string
		role pgtest.Role
		sql  string
	}{
		{"pro", pgtest.RoleWebhook, insertProSQL},
		{"comp", pgtest.RoleOwner, insertCompSQL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.register(t, userA)
			execAs(t, e.database.Connect(t, tt.role), userA, []string{tt.sql})
			afterTrial := clock.NewFixed(e.trialEnd(t, userA).Add(24 * time.Hour))
			policy := e.policy(afterTrial)

			for i := range access.TrialRecordCap + 2 {
				require.NoError(t, policy.ReserveRecord(t.Context(), userA, access.RecordExpense), "expense %d", i+1)
			}
			assert.Equal(t, access.TrialRecordCap+2, e.usage(t, userA).Expenses, "still counted")
		})
	}
}

func TestStore_RLSHidesAnotherUsersRows(t *testing.T) {
	e := newEnv(t)
	e.register(t, userA, userB)
	e.setUsage(t, userB, 1, 1)

	err := e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
		ents, err := e.store.Entitlements(ctx, userB)
		require.NoError(t, err)
		assert.Empty(t, ents, "B's entitlements are invisible to A")
		_, err = e.store.Reserve(ctx, userB, access.RecordIncome, access.Unlimited)
		assert.ErrorIs(t, err, access.ErrUsageMissing, "B's counter is invisible to A")
		_, err = e.store.Usage(ctx, userB)
		assert.ErrorIs(t, err, access.ErrUsageMissing)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, access.Usage{Incomes: 1, Expenses: 1}, e.usage(t, userB), "B untouched")
}

func TestReserveRecord_MissingUsageRowIsNotAPaywall(t *testing.T) {
	e := newEnv(t)
	execAs(t, e.app, userA, []string{registerSQL[0], registerSQL[2]}) // no user_usage row

	err := e.policy(systemClock()).ReserveRecord(t.Context(), userA, access.RecordIncome)

	require.ErrorIs(t, err, access.ErrUsageMissing)
	_, typed := errors.AsType[*apperr.Error](err)
	assert.False(t, typed, "a data bug is a 500, never a 402")
}

func TestStore_NeedsAUnitOfWork(t *testing.T) {
	store := accessgorm.NewStore()

	_, err := store.Entitlements(t.Context(), userA)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	_, err = store.Reserve(t.Context(), userA, access.RecordIncome, access.TrialRecordCap)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	_, err = store.Usage(t.Context(), userA)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
}

func TestSummary_ReflectsTheDatabase(t *testing.T) {
	e := newEnv(t)
	e.register(t, userA)
	e.setUsage(t, userA, 2, 5)

	summary, err := e.policy(systemClock()).Summary(t.Context(), userA)

	require.NoError(t, err)
	assert.Equal(t, access.LevelTrial, summary.Status.Level)
	assert.True(t, e.trialEnd(t, userA).Equal(summary.Status.TrialEndsAt))
	incomes, limited := summary.Remaining(access.RecordIncome)
	assert.True(t, limited)
	assert.Equal(t, 3, incomes)
	expenses, _ := summary.Remaining(access.RecordExpense)
	assert.Zero(t, expenses)
}
