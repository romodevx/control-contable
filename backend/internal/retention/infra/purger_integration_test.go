//go:build integration

package infra_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/retention/infra"
	"github.com/romodevx/control-contable/backend/internal/retention/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The adapter calls purge_soft_deleted as cc_worker through the System unit of work
// and maps every target's count, in the function's order.
func TestPurger_ReportsEveryTargetAsTheWorker(t *testing.T) {
	db := pgtest.NewDatabase(t)
	// email_verification_failures is ENABLE (not FORCE) RLS: the owner seeds it directly.
	_, err := db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(), `
INSERT INTO email_verification_failures (email_hash, failures, window_started_at)
VALUES (sha256('closed-window'), 2, now() - interval '25 hours')`)
	require.NoError(t, err)
	units, err := pgdb.NewTxManager(db.Connect(t, pgtest.RoleWorker))
	require.NoError(t, err)

	var counts []usecase.TargetCount
	require.NoError(t, units.System(t.Context(), func(ctx context.Context) error {
		var err error
		counts, err = infra.NewPurger().PurgeBatch(ctx, usecase.DefaultBatchSize)
		return err
	}))

	assert.Equal(t, []usecase.TargetCount{
		{Target: "incomes"}, {Target: "expenses"}, {Target: "personal_deductions"},
		{Target: "refresh_tokens"}, {Target: "email_verifications"},
		{Target: "email_verification_failures", Purged: 1}, {Target: "notification_log"},
	}, counts)
}

func TestPurger_RefusesToRunOutsideAUnitOfWork(t *testing.T) {
	_, err := infra.NewPurger().PurgeBatch(t.Context(), usecase.DefaultBatchSize)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
}

// The whole job against the real database: a full run purges, a second run is a no-op.
func TestPurgeJob_RunIsIdempotentAgainstTheDatabase(t *testing.T) {
	db := pgtest.NewDatabase(t)
	_, err := db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(), `
INSERT INTO email_verifications (email_hash, code_hash, pending, expires_at, created_at)
VALUES (sha256('old'), sha256('1'), '{}', now() - interval '25 hours', now() - interval '26 hours')`)
	require.NoError(t, err)
	units, err := pgdb.NewTxManager(db.Connect(t, pgtest.RoleWorker))
	require.NoError(t, err)
	purger := infra.NewPurger()
	job, err := usecase.NewPurgeJob(usecase.PurgeDeps{
		System: units, Purger: purger, Logger: slog.New(slog.DiscardHandler),
		BatchSize: usecase.DefaultBatchSize, MaxBatches: usecase.DefaultMaxBatches,
	})
	require.NoError(t, err)

	require.NoError(t, job.Run(t.Context()))
	require.NoError(t, job.Run(t.Context()))

	var left int
	require.NoError(t, db.Connect(t, pgtest.RoleOwner).QueryRowContext(t.Context(),
		"SELECT count(*) FROM email_verifications").Scan(&left))
	assert.Zero(t, left)
}
