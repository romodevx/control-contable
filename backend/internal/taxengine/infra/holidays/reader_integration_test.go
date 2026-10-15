//go:build integration

package holidays_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/holidays"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// readerUser is this suite's own unit-of-work user: holidays are global, so
// no users row is needed (RLS on them does not read app.current_user_id).
const readerUser = "0199d4a2-0000-7000-8000-00000000e9a1"

func read(t *testing.T, db *pgtest.Database) fiscal.HolidaySet {
	t.Helper()
	txm, err := pgdb.NewTxManager(db.Connect(t, pgtest.RoleApp))
	require.NoError(t, err)
	var set fiscal.HolidaySet
	require.NoError(t, txm.ReadOnly(t.Context(), readerUser, func(ctx context.Context) error {
		set, err = holidays.NewReader().Holidays(ctx)
		return err
	}))
	return set
}

func day(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

// The M9-03 seed, read as cc_app: 2026 and 2027 covered, their days move a
// due date, and the kernel's rule over it gives research §A.4.1's dates.
func TestReader_readsTheSeed(t *testing.T) {
	set := read(t, pgtest.NewDatabase(t))

	assert.Equal(t, []int{2026, 2027}, set.Years())
	for _, raw := range []string{"2026-01-01", "2026-09-16", "2026-11-16", "2027-03-26", "2027-12-25"} {
		assert.True(t, set.IsHoliday(day(t, raw)), raw)
	}
	assert.False(t, set.IsHoliday(day(t, "2026-09-17")))
	_, due, reason, err := fiscal.MonthlyDueDate(2026, 9, false, set)
	require.NoError(t, err)
	assert.Equal(t, "2026-10-19", due.String())
	assert.Equal(t, fiscal.MovedWeekend, reason)
}

// A day stored with moves_due_date = false is not in the set (the data
// says it does not move a payment deadline).
func TestReader_leavesOutDaysThatDoNotMoveDueDates(t *testing.T) {
	db := pgtest.NewDatabase(t)
	_, err := db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(),
		`UPDATE holidays SET moves_due_date = false WHERE day = DATE '2026-09-16'`)
	require.NoError(t, err)

	set := read(t, db)

	assert.False(t, set.IsHoliday(day(t, "2026-09-16")))
	assert.True(t, set.IsHoliday(day(t, "2026-11-16")))
}

// Outside a unit of work there is no connection (pgdb.Conn).
func TestReader_needsAUnitOfWork(t *testing.T) {
	_, err := holidays.NewReader().Holidays(t.Context())

	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
}
