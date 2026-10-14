//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Expenses migration (M4-03): docs/project-specs/MODULE-expenses.md §2.1, §5 (SE-1, SE-2)
// and docs/architecture/data-model.md (expenses).

const (
	expensesVersion = 20261013211000

	expenseA = "00000000-0000-7000-8000-0000000000e1"

	// Plan sizes for the index proofs: ~50k rows across many users, as the PLAN asks.
	planUsers           = 25
	planExpensesPerUser = 2000
)

var expensesTable = pgtest.Table{Name: "expenses", OwnerColumn: "user_id"}

// expenseColumns are the columns of a test expense insert, as SQL literals.
type expenseColumns map[string]string

// validExpense is a well-formed 612 software expense of userID's taxpayer.
func validExpense(userID, taxpayerID string) expenseColumns {
	return expenseColumns{
		"user_id": quote(userID), "taxpayer_id": quote(taxpayerID), "regime_code": "'612'",
		"category_code": "'software'", "amount_cents": "50000", "iva_rate": "'rate_16'", "iva_cents": "8000",
		"payment_method": "'card_credit'", "has_cfdi": "true", "deductible": "true", "paid_on": "'2026-09-12'",
	}
}

func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func insertExpenseSQL(cols expenseColumns) string {
	names, values := splitColumns(cols)
	return fmt.Sprintf("INSERT INTO expenses (%s) VALUES (%s)", names, values)
}

// withOverrides returns base with overrides applied; an empty value removes the column.
func withOverrides(base, overrides expenseColumns) expenseColumns {
	out := maps.Clone(base)
	for k, v := range overrides {
		if v == "" {
			delete(out, k)
			continue
		}
		out[k] = v
	}
	return out
}

func insertOwnExpense(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertExpenseSQL(validExpense(ownerID, taxpayerOf[ownerID])))
	return err
}

// seedExpenseOwners gives userA and userB a física taxpayer holding 612 since 2026-01.
func seedExpenseOwners(t *testing.T, app *sql.DB) {
	t.Helper()
	seedUsers(t, app, userA, userB)
	for _, user := range []string{userA, userB} {
		seedTaxpayer(t, app, user, taxpayerOf[user], "fisica")
		_, err := execAsUser(t, app, user, insertRegimeSQL, user, taxpayerOf[user], "612", "fisica", "2026-01", nil)
		require.NoError(t, err)
	}
}

func seedExpense(t *testing.T, app *sql.DB, userID string, overrides expenseColumns) {
	t.Helper()
	_, err := execAsUser(t, app, userID, insertExpenseSQL(withOverrides(validExpense(userID, taxpayerOf[userID]), overrides)))
	require.NoError(t, err)
}

func TestExpenses_AppRoleIsolation(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)

	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table: expensesTable, UserA: userA, UserB: userB, Insert: insertOwnExpense,
	})
}

// AssertIsolation's generic UPDATE rewrites user_id, which cc_app may not update at all;
// prove the cross-user writes on the columns it may write match no row.
func TestExpenses_CrossUserWritesMatchNoRow(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	seedExpense(t, app, userB, nil)
	writes := []string{
		"UPDATE expenses SET amount_cents = 1 WHERE user_id = $1",
		"UPDATE expenses SET category_code = 'otros', deductible = false WHERE user_id = $1",
		"UPDATE expenses SET deleted_at = now() WHERE user_id = $1",
		"UPDATE expenses SET deleted_at = NULL WHERE user_id = $1",
	}
	for _, query := range writes {
		t.Run(query, func(t *testing.T) {
			affected, err := execAsUser(t, app, userA, query, userB)
			require.NoError(t, err)
			assert.Zero(t, affected)
		})
	}
	t.Run("own rows are writable", func(t *testing.T) {
		for _, query := range writes {
			affected, err := execAsUser(t, app, userB, query, userB)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
	t.Run("an update cannot hand the row to another taxpayer", func(t *testing.T) {
		_, err := execAsUser(t, app, userB, "UPDATE expenses SET taxpayer_id = $1", taxpayerA)
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
}

// The composite FK (user_id, taxpayer_id, regime_code) -> taxpayer_regimes proves the
// expense belongs to a regime its own taxpayer holds, even with forged ids (FK checks
// bypass RLS); the category must exist in the catalog (SE-1).
func TestExpenses_CompositeForeignKeysRefuseForgedReferences(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	base := validExpense(userA, taxpayerA)
	cases := []struct {
		name      string
		overrides expenseColumns
		want      string
	}{
		{"own user id, another user's taxpayer", expenseColumns{"taxpayer_id": quote(taxpayerB)}, sqlstateForeignKeyViolation},
		{"another user's id and taxpayer", expenseColumns{"user_id": quote(userB), "taxpayer_id": quote(taxpayerB)}, sqlstateInsufficientPrivilege},
		{"own taxpayer, regime it does not hold", expenseColumns{"regime_code": "'606'"}, sqlstateForeignKeyViolation},
		{"own taxpayer, unknown regime", expenseColumns{"regime_code": "'999'"}, sqlstateForeignKeyViolation},
		{"unknown category", expenseColumns{"category_code": "'viajes-luna'"}, sqlstateForeignKeyViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(base, c.overrides)))
			requireSQLState(t, err, c.want)
		})
	}
	t.Run("the well-formed reference is accepted", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, insertExpenseSQL(base))
		require.NoError(t, err)
	})
	t.Run("an update to a regime the taxpayer does not hold is refused", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, "UPDATE expenses SET regime_code = '606'")
		requireSQLState(t, err, sqlstateForeignKeyViolation)
	})
}

func TestExpenses_RejectInvalidRows(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	base := validExpense(userA, taxpayerA)
	cases := []struct {
		name      string
		overrides expenseColumns
		want      string
	}{
		{"zero amount", expenseColumns{"amount_cents": "0", "iva_cents": "0"}, sqlstateCheckViolation},
		{"negative amount", expenseColumns{"amount_cents": "-1"}, sqlstateCheckViolation},
		{"amount over the contract maximum", expenseColumns{"amount_cents": "100000000000000"}, sqlstateCheckViolation},
		{"negative iva", expenseColumns{"iva_cents": "-1"}, sqlstateCheckViolation},
		{"iva over the contract maximum", expenseColumns{"iva_cents": "100000000000000"}, sqlstateCheckViolation},
		{"iva on a zero-rated expense", expenseColumns{"iva_rate": "'rate_0'"}, sqlstateCheckViolation},
		{"iva on an exempt expense", expenseColumns{"iva_rate": "'exempt'"}, sqlstateCheckViolation},
		{"unknown iva rate", expenseColumns{"iva_rate": "'rate_15'"}, sqlstateCheckViolation},
		{"unknown payment method", expenseColumns{"payment_method": "'card'"}, sqlstateCheckViolation},
		{"negative isr withheld", expenseColumns{"isr_withheld_cents": "-1"}, sqlstateCheckViolation},
		{"isr withheld above the amount", expenseColumns{"isr_withheld_cents": "50001"}, sqlstateCheckViolation},
		{"negative iva withheld", expenseColumns{"iva_withheld_cents": "-1"}, sqlstateCheckViolation},
		{"iva withheld above the iva", expenseColumns{"iva_withheld_cents": "8001"}, sqlstateCheckViolation},
		{"cfdi uuid without a cfdi", expenseColumns{"has_cfdi": "false", "cfdi_uuid": "'6F9619FF-8B86-D011-B42D-00C04FC964FF'"}, sqlstateCheckViolation},
		{"lower-case cfdi uuid", expenseColumns{"cfdi_uuid": "'6f9619ff-8b86-d011-b42d-00c04fc964ff'"}, sqlstateCheckViolation},
		{"malformed cfdi uuid", expenseColumns{"cfdi_uuid": "'6F9619FF8B86D011B42D00C04FC964FF'"}, sqlstateCheckViolation},
		{"paid before 2022", expenseColumns{"paid_on": "'2021-12-31'"}, sqlstateCheckViolation},
		{"empty description", expenseColumns{"description": "''"}, sqlstateCheckViolation},
		{"untrimmed description", expenseColumns{"description": "' Renta'"}, sqlstateCheckViolation},
		{"description over 200 characters", expenseColumns{"description": "repeat('a', 201)"}, sqlstateCheckViolation},
		{"description not NFC", expenseColumns{"description": "U&'Cafe\\0301'"}, sqlstateCheckViolation},
		{"no deductible verdict", expenseColumns{"deductible": "NULL"}, sqlstateNotNullViolation},
		{"no paid_on", expenseColumns{"paid_on": ""}, sqlstateNotNullViolation},
		{"no payment method", expenseColumns{"payment_method": ""}, sqlstateNotNullViolation},
		{"no has_cfdi", expenseColumns{"has_cfdi": ""}, sqlstateNotNullViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(base, c.overrides)))
			requireSQLState(t, err, c.want)
		})
	}
	accepted := []struct {
		name      string
		overrides expenseColumns
	}{
		{"minimum amount", expenseColumns{"amount_cents": "1", "iva_cents": "0"}},
		{"contract maximum amount", expenseColumns{"amount_cents": "99999999999999"}},
		{"exempt with zero iva", expenseColumns{"iva_rate": "'exempt'", "iva_cents": "0"}},
		{"border rate", expenseColumns{"iva_rate": "'rate_8'", "iva_cents": "4000"}},
		{"every withholding at its bound", expenseColumns{"isr_withheld_cents": "50000", "iva_withheld_cents": "8000"}},
		{"cfdi uuid in upper case", expenseColumns{"cfdi_uuid": "'6F9619FF-8B86-D011-B42D-00C04FC964FF'"}},
		{"no cfdi, user override", expenseColumns{"has_cfdi": "false", "deductible": "true", "deductible_overridden": "true"}},
		{"cash payment", expenseColumns{"payment_method": "'cash'", "deductible": "false"}},
		{"NFC description of 200 characters", expenseColumns{"description": "repeat('é', 200)"}},
		{"paid on the catalog floor", expenseColumns{"paid_on": "'2022-01-01'"}},
	}
	for _, c := range accepted {
		t.Run("accepted: "+c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(base, c.overrides)))
			require.NoError(t, err)
		})
	}
	for _, method := range []string{"cash", "transfer", "card_debit", "card_credit", "check", "e_wallet", "other"} {
		t.Run("accepted: payment method "+method, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(base, expenseColumns{"payment_method": quote(method)})))
			require.NoError(t, err)
		})
	}
}

// Idempotency (MODULE-incomes §2.3, shared): the client-generated id is the key; the
// server default is uuidv7. A repeated id is a primary-key conflict, never a second row.
// cfdi_uuid is NOT unique: a duplicate is a warning (duplicate_cfdi_uuid), not an error.
func TestExpenses_ClientIDIsTheKeyAndCFDIUUIDIsNotUnique(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	const cfdi = "'6F9619FF-8B86-D011-B42D-00C04FC964FF'"

	seedExpense(t, app, userA, expenseColumns{"id": quote(expenseA), "cfdi_uuid": cfdi})
	_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(validExpense(userA, taxpayerA),
		expenseColumns{"id": quote(expenseA)})))
	requireSQLState(t, err, sqlstateUniqueViolation)

	seedExpense(t, app, userA, expenseColumns{"cfdi_uuid": cfdi})
	var rows int
	var version int
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), `
			SELECT count(*), max(uuid_extract_version(id)) FILTER (WHERE id <> $1) FROM expenses WHERE cfdi_uuid = `+cfdi,
			expenseA).Scan(&rows, &version)
	}))
	assert.Equal(t, 2, rows, "the same folio fiscal twice is allowed")
	assert.Equal(t, 7, version, "server-generated ids are uuidv7")
}

// Least privilege: cc_app reads, inserts, edits the mutable columns, soft-deletes and
// restores (deleted_at); it never hard-deletes (rows leave by cascade or by the purge
// job, M9-08) and never moves a row to another user, taxpayer or id.
func TestExpenses_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	seedExpense(t, app, userA, nil)

	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE expenses SET id = uuidv7()"},
		{pgtest.RoleApp, "UPDATE expenses SET user_id = user_id"},
		{pgtest.RoleApp, "UPDATE expenses SET taxpayer_id = taxpayer_id"},
		{pgtest.RoleApp, "UPDATE expenses SET created_at = now()"},
		{pgtest.RoleApp, "UPDATE expenses SET updated_at = now()"},
		{pgtest.RoleApp, "DELETE FROM expenses"},
		{pgtest.RoleApp, "TRUNCATE expenses"},
		{pgtest.RoleWebhook, "SELECT 1 FROM expenses"},
		{pgtest.RoleWorker, "SELECT 1 FROM expenses"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
	t.Run("cc_app may not insert a deleted row", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, insertExpenseSQL(withOverrides(validExpense(userA, taxpayerA),
			expenseColumns{"deleted_at": "now()"})))
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
	t.Run("cc_app writes what the use cases need", func(t *testing.T) {
		for _, query := range []string{
			`UPDATE expenses SET regime_code = '612', category_code = 'papeleria', description = 'Hojas',
			        amount_cents = 100000, iva_rate = 'rate_16', iva_cents = 16000, payment_method = 'cash',
			        has_cfdi = true, cfdi_uuid = '6F9619FF-8B86-D011-B42D-00C04FC964FF', deductible = false,
			        deductible_overridden = false, isr_withheld_cents = 0, iva_withheld_cents = 0, paid_on = '2026-09-13'`,
			"UPDATE expenses SET deductible = true, deductible_overridden = true",
			"UPDATE expenses SET deleted_at = now()",
			"UPDATE expenses SET deleted_at = NULL",
		} {
			affected, err := execAsUser(t, app, userA, query)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// Account deletion (US-A4) and taxpayer deletion cascade to expenses; a category in use
// cannot be removed from the catalog.
func TestExpenses_DeletionCascades(t *testing.T) {
	setup := func(t *testing.T) *sql.DB {
		t.Helper()
		app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
		seedExpenseOwners(t, app)
		seedExpense(t, app, userA, nil)
		seedExpense(t, app, userA, nil)
		_, err := execAsUser(t, app, userA, "UPDATE expenses SET deleted_at = now() WHERE id IN (SELECT id FROM expenses LIMIT 1)")
		require.NoError(t, err)
		return app
	}
	countExpenses := func(t *testing.T, app *sql.DB) int {
		t.Helper()
		var n int
		require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM expenses").Scan(&n)
		}))
		return n
	}
	for _, c := range []struct{ name, query, id string }{
		{"account deletion", "DELETE FROM users WHERE id = $1", userA},
		{"taxpayer deletion", "DELETE FROM taxpayers WHERE id = $1", taxpayerA},
	} {
		t.Run(c.name, func(t *testing.T) {
			app := setup(t)
			require.Equal(t, 2, countExpenses(t, app))
			affected, err := execAsUser(t, app, userA, c.query, c.id)
			require.NoError(t, err)
			require.EqualValues(t, 1, affected)
			assert.Zero(t, countExpenses(t, app), "live and soft-deleted rows both go")
		})
	}
	t.Run("a category in use stays in the catalog", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		app := db.Connect(t, pgtest.RoleApp)
		seedExpenseOwners(t, app)
		seedExpense(t, app, userA, nil)
		owner := db.Connect(t, pgtest.RoleOwner)
		_, err := owner.ExecContext(t.Context(), "DELETE FROM category_regime_rules WHERE category_code = 'software'")
		require.NoError(t, err)
		_, err = owner.ExecContext(t.Context(), "DELETE FROM expense_categories WHERE code = 'software'")
		requireSQLState(t, err, sqlstateForeignKeyViolation)
	})
}

func TestExpenses_ForeignKeyActions(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	got := queryStrings(t, owner, `
		SELECT confrelid::regclass::text || ':' || confdeltype::text
		FROM pg_constraint WHERE contype = 'f' AND conrelid = 'expenses'::regclass ORDER BY 1`)
	assert.Equal(t, []string{"expense_categories:a", "taxpayer_regimes:c", "users:c"}, got)
}

func TestExpenses_UpdatedAtFollowsWrites(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seedExpenseOwners(t, app)
	seedExpense(t, app, userA, nil)
	var followed bool
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.01)"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(t.Context(), "UPDATE expenses SET description = 'Licencia'"); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(),
			"SELECT bool_and(updated_at = now() AND updated_at > created_at) FROM expenses").Scan(&followed)
	}))
	assert.True(t, followed)
}

// planUserID and planTaxpayerID number the users and taxpayers of the plan dataset.
func planUserID(i int) string     { return fmt.Sprintf("00000000-0000-4000-8000-0000000100%02d", i) }
func planTaxpayerID(i int) string { return fmt.Sprintf("00000000-0000-7000-8000-0000000200%02d", i) }

// seedPlanExpensesSQL writes planExpensesPerUser expenses for user $1 / taxpayer $2,
// deterministic: paid over 2025-01-01..2026-09-30, a third with a folio fiscal, a tenth
// without CFDI, rotating categories and payment methods.
var seedPlanExpensesSQL = fmt.Sprintf(`
INSERT INTO expenses (user_id, taxpayer_id, regime_code, category_code, amount_cents, iva_rate, iva_cents,
                      payment_method, has_cfdi, cfdi_uuid, deductible, paid_on)
SELECT $1::uuid, $2::uuid, '612',
       (ARRAY['software', 'papeleria', 'renta_local', 'combustible', 'restaurantes'])[1 + g %% 5],
       10000 + (g * 37) %% 500000, 'rate_16', round((10000 + (g * 37) %% 500000) * 0.16)::bigint,
       (ARRAY['cash', 'transfer', 'card_debit', 'card_credit'])[1 + g %% 4],
       g %% 10 <> 0,
       CASE WHEN g %% 3 = 0 AND g %% 10 <> 0 THEN upper(md5($1::uuid::text || g)::uuid::text) END,
       g %% 10 <> 0,
       date '2025-01-01' + (g * 7) %% 638
FROM generate_series(1, %d) AS g`, planExpensesPerUser)

// The hot queries of M4 (list page, month totals, the M5 estimate aggregate, the
// duplicate folio lookup) are served by the partial indexes at ~50k rows, under RLS as
// cc_app (data-model.md "Each critical query gets EXPLAIN ANALYZE in its module PR").
func TestExpenses_IndexesServeTheHotQueriesAt50kRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	for i := range planUsers {
		user, taxpayer := planUserID(i), planTaxpayerID(i)
		seedUsers(t, app, user)
		seedTaxpayer(t, app, user, taxpayer, "fisica")
		_, err := execAsUser(t, app, user, insertRegimeSQL, user, taxpayer, "612", "fisica", "2022-01", nil)
		require.NoError(t, err)
		_, err = execAsUser(t, app, user, seedPlanExpensesSQL, user, taxpayer)
		require.NoError(t, err)
		// One in twenty is soft-deleted (Undo window, purge after 30 days).
		_, err = execAsUser(t, app, user, "UPDATE expenses SET deleted_at = now() WHERE amount_cents % 20 = 0")
		require.NoError(t, err)
	}
	owner := db.Connect(t, pgtest.RoleOwner)
	_, err := owner.ExecContext(t.Context(), "ANALYZE expenses")
	require.NoError(t, err)
	var total int
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT reltuples::int FROM pg_class WHERE oid = 'expenses'::regclass").Scan(&total))
	require.GreaterOrEqual(t, total, planUsers*planExpensesPerUser*9/10, "~50k rows")

	user, taxpayer := planUserID(7), planTaxpayerID(7)
	queries := []struct {
		name, index, query string
	}{
		{"list, first page", "expenses_live_paid_on_idx", fmt.Sprintf(`
			SELECT id, regime_code, category_code, amount_cents, iva_cents, payment_method, has_cfdi, deductible, paid_on
			FROM expenses
			WHERE taxpayer_id = '%s' AND deleted_at IS NULL AND paid_on >= '2026-09-01' AND paid_on < '2026-10-01'
			ORDER BY paid_on DESC, id DESC LIMIT 51`, taxpayer)},
		{"list, keyset page", "expenses_live_paid_on_idx", fmt.Sprintf(`
			SELECT id, amount_cents, paid_on FROM expenses
			WHERE taxpayer_id = '%s' AND deleted_at IS NULL AND paid_on >= '2026-09-01' AND paid_on < '2026-10-01'
			  AND (paid_on, id) < ('2026-09-15', 'ffffffff-ffff-7fff-bfff-ffffffffffff')
			ORDER BY paid_on DESC, id DESC LIMIT 51`, taxpayer)},
		{"month totals by regime", "expenses_live_paid_on_idx", fmt.Sprintf(`
			SELECT count(*), sum(amount_cents), sum(iva_cents) FROM expenses
			WHERE taxpayer_id = '%s' AND deleted_at IS NULL AND regime_code = '612'
			  AND paid_on >= '2026-09-01' AND paid_on < '2026-10-01'`, taxpayer)},
		{"estimate aggregate, year to date", "expenses_live_paid_on_idx", fmt.Sprintf(`
			SELECT date_trunc('month', paid_on) AS month, regime_code, deductible, has_cfdi,
			       sum(amount_cents), sum(iva_cents), sum(isr_withheld_cents), sum(iva_withheld_cents)
			FROM expenses
			WHERE taxpayer_id = '%s' AND deleted_at IS NULL AND paid_on >= '2026-01-01' AND paid_on < '2026-10-01'
			GROUP BY 1, 2, 3, 4`, taxpayer)},
		{"duplicate folio lookup", "expenses_live_cfdi_uuid_idx", fmt.Sprintf(`
			SELECT EXISTS (SELECT 1 FROM expenses
			               WHERE taxpayer_id = '%s' AND cfdi_uuid = upper(md5('%s' || 3)::uuid::text)
			                 AND deleted_at IS NULL AND id <> '%s')`, taxpayer, user, expenseA)},
	}
	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			plan := explain(t, app, user, q.query)
			assert.Contains(t, plan, q.index)
			assert.NotContains(t, plan, "Seq Scan on expenses")
		})
	}
}

func TestExpensesMigration_DownRemovesEverything(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), expensesVersion-1)
	require.NoError(t, err)
	var found sql.NullString
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass('public.expenses')::text").Scan(&found))
	assert.False(t, found.Valid, "down left %s", found.String)
	var categories bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT to_regclass('public.expense_categories') IS NOT NULL").Scan(&categories))
	assert.True(t, categories, "the catalog belongs to the previous version")

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var exists bool
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass('public.expenses') IS NOT NULL").Scan(&exists))
	assert.True(t, exists)
}
