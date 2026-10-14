//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// estimate_views migration (M8-02): the last view of one month's estimate per user,
// docs/project-specs/MODULE-history.md §2.3 (DECIDED) and §2.6 (read path).

const (
	estimateViewsVersion = 20261014180000

	// planViewYears seeds every month of 2022..2100 per user for the read-path plan.
	planViewFirstYear = 2022
	planViewLastYear  = 2100

	sqlstateStringDataRightTruncation = "22001"
)

var estimateViewsTable = pgtest.Table{Name: "estimate_views", OwnerColumn: "user_id"}

// upsertViewSQL is the one statement markHistoryViewed runs (§2.3): the last view wins.
const upsertViewSQL = `
INSERT INTO estimate_views (user_id, taxpayer_id, period, records_fingerprint, figures_fingerprint,
                            engine_version, parameter_set, viewed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id, taxpayer_id, period) DO UPDATE SET
  records_fingerprint = EXCLUDED.records_fingerprint,
  figures_fingerprint = EXCLUDED.figures_fingerprint,
  engine_version      = EXCLUDED.engine_version,
  parameter_set       = EXCLUDED.parameter_set,
  viewed_at           = EXCLUDED.viewed_at`

// fingerprint is a well-formed SHA-256 hex digest built from one repeated digit, so no
// literal in this file looks like a secret.
func fingerprint(digit string) string { return strings.Repeat(digit, 64) }

// viewRow is one upsertViewSQL call.
type viewRow struct {
	userID, taxpayerID, period string
	records, figures           string
	engine, parameters         string
	viewedAt                   string
}

func (r viewRow) args() []any {
	return []any{r.userID, r.taxpayerID, r.period, r.records, r.figures, r.engine, r.parameters, r.viewedAt}
}

// ownView is a well-formed view of 2026-03 of taxpayerOf[userID].
func ownView(userID string) viewRow {
	return viewRow{
		userID: userID, taxpayerID: taxpayerOf[userID], period: "2026-03-01",
		records: fingerprint("a"), figures: fingerprint("b"),
		engine: "1.4.0", parameters: "2026/v1", viewedAt: "2026-10-03T12:00:00Z",
	}
}

func upsertView(t *testing.T, app *sql.DB, asUser string, r viewRow) (int64, error) {
	t.Helper()
	return execAsUser(t, app, asUser, upsertViewSQL, r.args()...)
}

func insertOwnView(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, upsertViewSQL, ownView(ownerID).args()...)
	return err
}

func countViews(t *testing.T, app *sql.DB, userID string) int {
	t.Helper()
	var n int
	queryAsUser(t, app, userID, "SELECT count(*) FROM estimate_views", &n)
	return n
}

// §2.3: ENABLE + FORCE RLS; user A cannot read, insert, update or delete B's rows.
func TestEstimateViews_AppRoleIsolation(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))

	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table: estimateViewsTable, UserA: userA, UserB: userB, Insert: insertOwnView,
	})
}

// AssertIsolation's generic UPDATE rewrites user_id, which cc_app may not update at all;
// prove the writes the upsert can run match no foreign row.
func TestEstimateViews_CrossUserWritesMatchNoRow(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	_, err := upsertView(t, app, userB, ownView(userB))
	require.NoError(t, err)

	affected, err := execAsUser(t, app, userA,
		"UPDATE estimate_views SET viewed_at = now() WHERE taxpayer_id = $1", taxpayerB)
	require.NoError(t, err)
	assert.Zero(t, affected)

	t.Run("an upsert naming B's key is an insert for B, refused by WITH CHECK", func(t *testing.T) {
		_, err := upsertView(t, app, userA, ownView(userB))
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
	t.Run("own rows are writable", func(t *testing.T) {
		affected, err := execAsUser(t, app, userB,
			"UPDATE estimate_views SET viewed_at = now() WHERE taxpayer_id = $1", taxpayerB)
		require.NoError(t, err)
		assert.EqualValues(t, 1, affected)
	})
}

// §2.3: the composite FK (user_id, taxpayer_id) -> taxpayers (user_id, id) makes a view of
// another user's taxpayer impossible even with forged ids (FK checks bypass RLS).
func TestEstimateViews_CompositeForeignKeyRefusesForgedReferences(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	forged := func(userID, taxpayerID string) viewRow {
		r := ownView(userA)
		r.userID, r.taxpayerID = userID, taxpayerID
		return r
	}
	cases := []struct {
		name string
		row  viewRow
		want string
	}{
		{"own user id, another user's taxpayer", forged(userA, taxpayerB), sqlstateForeignKeyViolation},
		{"own user id, unknown taxpayer", forged(userA, "00000000-0000-7000-8000-0000000000ff"), sqlstateForeignKeyViolation},
		{"another user's id and taxpayer", forged(userB, taxpayerB), sqlstateInsufficientPrivilege},
		{"another user's id, own taxpayer", forged(userB, taxpayerA), sqlstateInsufficientPrivilege},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := upsertView(t, app, userA, c.row)
			requireSQLState(t, err, c.want)
		})
	}
	t.Run("the well-formed reference is accepted", func(t *testing.T) {
		_, err := upsertView(t, app, userA, ownView(userA))
		require.NoError(t, err)
	})
}

// §2.3 column checks: first-of-month period, SHA-256 hex fingerprints, bounded versions.
func TestEstimateViews_RejectInvalidRows(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	view := func(edit func(*viewRow)) viewRow {
		r := ownView(userA)
		edit(&r)
		return r
	}
	cases := []struct {
		name string
		row  viewRow
	}{
		{"period not the first of the month", view(func(r *viewRow) { r.period = "2026-03-02" })},
		{"period before the catalog floor", view(func(r *viewRow) { r.period = "2021-12-01" })},
		{"period after FiscalYear.maximum", view(func(r *viewRow) { r.period = "2101-01-01" })},
		{"records fingerprint too short", view(func(r *viewRow) { r.records = strings.Repeat("a", 63) })},
		{"records fingerprint upper case", view(func(r *viewRow) { r.records = fingerprint("A") })},
		{"records fingerprint not hex", view(func(r *viewRow) { r.records = fingerprint("g") })},
		{"figures fingerprint too short", view(func(r *viewRow) { r.figures = strings.Repeat("b", 63) })},
		{"figures fingerprint not hex", view(func(r *viewRow) { r.figures = fingerprint("z") })},
		{"empty engine version", view(func(r *viewRow) { r.engine = "" })},
		{"engine version over 32 chars", view(func(r *viewRow) { r.engine = strings.Repeat("9", 33) })},
		{"empty parameter set", view(func(r *viewRow) { r.parameters = "" })},
		{"parameter set over 32 chars", view(func(r *viewRow) { r.parameters = strings.Repeat("v", 33) })},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := upsertView(t, app, userA, c.row)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}
	t.Run("a fingerprint longer than 64 does not fit", func(t *testing.T) {
		_, err := upsertView(t, app, userA, view(func(r *viewRow) { r.figures = strings.Repeat("b", 65) }))
		requireSQLState(t, err, sqlstateStringDataRightTruncation)
	})
	t.Run("the boundary values are accepted", func(t *testing.T) {
		for _, r := range []viewRow{
			view(func(r *viewRow) { r.period = "2022-01-01" }),
			view(func(r *viewRow) { r.period = "2100-12-01" }),
			view(func(r *viewRow) {
				r.period, r.records, r.figures = "2026-04-01", fingerprint("0"), fingerprint("f")
				r.engine, r.parameters = strings.Repeat("9", 32), strings.Repeat("v", 32)
			}),
		} {
			_, err := upsertView(t, app, userA, r)
			require.NoError(t, err, "%+v", r)
		}
	})
}

// §2.3: write = one INSERT ... ON CONFLICT DO UPDATE; idempotent, and the last view wins.
func TestEstimateViews_UpsertIsIdempotentAndLastViewWins(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	first := ownView(userA)
	for range 2 {
		affected, err := upsertView(t, app, userA, first)
		require.NoError(t, err)
		assert.EqualValues(t, 1, affected)
	}
	assert.Equal(t, 1, countViews(t, app, userA), "a repeated view keeps one row")

	later := first
	later.records, later.figures = fingerprint("c"), fingerprint("d")
	later.engine, later.parameters, later.viewedAt = "1.5.0", "2026/v2", "2026-10-05T09:30:00Z"
	_, err := upsertView(t, app, userA, later)
	require.NoError(t, err)

	var records, figures, engine, parameters, viewedAt string
	queryAsUser(t, app, userA, `
		SELECT records_fingerprint, figures_fingerprint, engine_version, parameter_set,
		       to_char(viewed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM estimate_views`, &records, &figures, &engine, &parameters, &viewedAt)
	assert.Equal(t, []string{later.records, later.figures, later.engine, later.parameters, later.viewedAt},
		[]string{records, figures, engine, parameters, viewedAt})
	assert.Equal(t, 1, countViews(t, app, userA))

	t.Run("another month is another row", func(t *testing.T) {
		other := first
		other.period = "2026-04-01"
		_, err := upsertView(t, app, userA, other)
		require.NoError(t, err)
		assert.Equal(t, 2, countViews(t, app, userA))
	})
}

// §2.3 / US-A4 / MODULE-taxpayers §2.7: account and taxpayer deletion cascade to the views.
func TestEstimateViews_DeletionCascades(t *testing.T) {
	setup := func(t *testing.T) *sql.DB {
		t.Helper()
		app := seedIncomeOwners(t, pgtest.NewDatabase(t))
		for _, user := range []string{userA, userB} {
			_, err := upsertView(t, app, user, ownView(user))
			require.NoError(t, err)
		}
		return app
	}
	t.Run("account deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		assert.Zero(t, countViews(t, app, userA))
		assert.Equal(t, 1, countViews(t, app, userB), "another user's views stay")
	})
	t.Run("taxpayer deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM taxpayers WHERE id = $1", taxpayerA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		assert.Zero(t, countViews(t, app, userA))
		assert.Equal(t, 1, countViews(t, app, userB))
	})
	t.Run("the users-FK guard sees estimate_views directly and through taxpayers", func(t *testing.T) {
		owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
		assert.Contains(t, queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'users'::regclass AND confdeltype = 'c'"),
			"estimate_views")
		assert.Contains(t, queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'taxpayers'::regclass AND confdeltype = 'c'"),
			"estimate_views")
	})
}

// Least privilege: cc_app selects and upserts; the key never changes, rows leave only by
// cascade, and no other runtime role reaches the table.
func TestEstimateViews_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	_, err := upsertView(t, app, userA, ownView(userA))
	require.NoError(t, err)
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE estimate_views SET user_id = user_id"},
		{pgtest.RoleApp, "UPDATE estimate_views SET taxpayer_id = taxpayer_id"},
		{pgtest.RoleApp, "UPDATE estimate_views SET period = period"},
		{pgtest.RoleApp, "UPDATE estimate_views SET created_at = now()"},
		{pgtest.RoleApp, "UPDATE estimate_views SET updated_at = now()"},
		{pgtest.RoleApp, "DELETE FROM estimate_views"},
		{pgtest.RoleApp, "TRUNCATE estimate_views"},
		{pgtest.RoleWebhook, "SELECT 1 FROM estimate_views"},
		{pgtest.RoleWorker, "SELECT 1 FROM estimate_views"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
	t.Run("cc_app may not write the audit columns on insert", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, `
			INSERT INTO estimate_views (user_id, taxpayer_id, period, records_fingerprint, figures_fingerprint,
			                            engine_version, parameter_set, viewed_at, created_at)
			VALUES ($1, $2, '2026-05-01', $3, $3, '1', '2026/v1', now(), now())`, userA, taxpayerA, fingerprint("e"))
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
	t.Run("the policies cover exactly what the grants allow", func(t *testing.T) {
		owner := db.Connect(t, pgtest.RoleOwner)
		assert.Equal(t, []string{"INSERT", "SELECT", "UPDATE"}, queryStrings(t, owner,
			"SELECT cmd FROM pg_policies WHERE tablename = 'estimate_views' AND roles = '{cc_app}' ORDER BY cmd"))
		assert.Equal(t, []string{"UPDATE"}, queryStrings(t, owner,
			"SELECT cmd FROM pg_policies WHERE tablename = 'estimate_views' AND with_check IS NOT NULL AND qual IS NOT NULL"),
			"UPDATE carries USING and WITH CHECK")
	})
}

// updated_at is now() of the writing transaction (set_updated_at), also through the upsert.
func TestEstimateViews_UpdatedAtFollowsWrites(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	_, err := upsertView(t, app, userA, ownView(userA))
	require.NoError(t, err)
	var followed bool
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.01)"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(t.Context(), upsertViewSQL, ownView(userA).args()...); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(),
			"SELECT bool_and(updated_at = now() AND updated_at > created_at) FROM estimate_views").Scan(&followed)
	}))
	assert.True(t, followed)
}

// planViewsSQL writes, for user $1 / taxpayer $2, a view of every month 2022..2100.
var planViewsSQL = fmt.Sprintf(`
INSERT INTO estimate_views (user_id, taxpayer_id, period, records_fingerprint, figures_fingerprint,
                            engine_version, parameter_set, viewed_at)
SELECT $1::uuid, $2::uuid, make_date(y, m, 1), repeat('a', 64), repeat('b', 64), '1.4.0', '2026/v1', now()
FROM generate_series(%d, %d) AS y, generate_series(1, 12) AS m`, planViewFirstYear, planViewLastYear)

// §2.3 / §2.6 step 6: the PK btree serves the fiscal-year read (one range scan), the
// upsert and both cascades; no other index exists.
func TestEstimateViews_PrimaryKeyServesTheYearRead(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	for i := range planUsers {
		user, taxpayer := planUserID(i), planTaxpayerID(i)
		seedUsers(t, app, user)
		seedTaxpayer(t, app, user, taxpayer, "fisica")
		_, err := execAsUser(t, app, user, planViewsSQL, user, taxpayer)
		require.NoError(t, err)
	}
	owner := db.Connect(t, pgtest.RoleOwner)
	_, err := owner.ExecContext(t.Context(), "ANALYZE estimate_views")
	require.NoError(t, err)

	defs := queryStrings(t, owner,
		"SELECT indexname || ' ' || regexp_replace(indexdef, '^.* USING ', '') FROM pg_indexes WHERE tablename = 'estimate_views' ORDER BY 1")
	assert.Equal(t, []string{"estimate_views_pkey btree (user_id, taxpayer_id, period)"}, defs)

	user, taxpayer := planUserID(7), planTaxpayerID(7)
	plan := explain(t, app, user, fmt.Sprintf(`
		SELECT period, records_fingerprint, figures_fingerprint, engine_version, parameter_set, viewed_at
		FROM estimate_views
		WHERE taxpayer_id = '%s' AND period BETWEEN '2026-01-01' AND '2026-12-01'`, taxpayer))
	assert.Contains(t, plan, "estimate_views_pkey")
	assert.NotContains(t, plan, "Seq Scan on estimate_views")
}

func TestEstimateViewsMigration_DownRemovesEverything(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), estimateViewsVersion-1)
	require.NoError(t, err)
	var found sql.NullString
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass('public.estimate_views')::text").Scan(&found))
	assert.False(t, found.Valid, "down left %s", found.String)

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var exists bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT to_regclass('public.estimate_views') IS NOT NULL").Scan(&exists))
	assert.True(t, exists)
}
