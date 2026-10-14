//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Tax inputs migrations (M5-20): the catalog's new input kinds and taxpayer_tax_inputs,
// docs/project-specs/MODULE-tax-engine.md §2.8, §6 (S-1), §9 item 6 and OQ-T1, and
// docs/api/openapi.yaml TaxInputDef / TaxInputValue.

const (
	taxInputKindsVersion = 20261014140000
	taxInputsVersion     = 20261014141000

	// taxInputsMoralTaxpayer is a persona moral of userA holding 601 (rate and boolean kinds).
	taxInputsMoralTaxpayer = "00000000-0000-7000-8000-0000000000a3"

	// planTaxInputYears seeds every fiscal year the table admits (2022..2100) per user.
	planTaxInputFirstYear = 2022
	planTaxInputLastYear  = 2100
)

var taxInputsTable = pgtest.Table{Name: "taxpayer_tax_inputs", OwnerColumn: "user_id"}

// insertTaxInputSQL writes one value. Every column the app may write is a parameter;
// kind is not: the database derives it from regimes.required_inputs.
const insertTaxInputSQL = `
INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month,
                                 value_cents, value_rate, value_bool)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

// taxInputRow is one insertTaxInputSQL call.
type taxInputRow struct {
	userID, taxpayerID, regime string
	fiscalYear                 int
	key                        string
	month, cents, rate, flag   any
}

func (r taxInputRow) args() []any {
	return []any{r.userID, r.taxpayerID, r.regime, r.fiscalYear, r.key, r.month, r.cents, r.rate, r.flag}
}

// monthlyPayment is a 612 prior provisional payment (money_monthly) of taxpayerOf[userID].
func monthlyPayment(userID string, month int) taxInputRow {
	return taxInputRow{
		userID: userID, taxpayerID: taxpayerOf[userID], regime: "612", fiscalYear: 2026,
		key: "prior_provisional_payments", month: month, cents: int64(200000),
	}
}

// annualMoney is a 612 money input (ptu_paid) of taxpayerOf[userID].
func annualMoney(userID string) taxInputRow {
	return taxInputRow{
		userID: userID, taxpayerID: taxpayerOf[userID], regime: "612", fiscalYear: 2026,
		key: "ptu_paid", cents: int64(0),
	}
}

func insertTaxInput(t *testing.T, app *sql.DB, asUser string, r taxInputRow) error {
	t.Helper()
	_, err := execAsUser(t, app, asUser, insertTaxInputSQL, r.args()...)
	return err
}

func insertOwnTaxInput(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertTaxInputSQL, monthlyPayment(ownerID, 1).args()...)
	return err
}

// seedTaxInputOwners creates users A and B, each with one física taxpayer holding 612
// since 2022-01, plus a moral taxpayer of A holding 601, and returns cc_app's connection.
func seedTaxInputOwners(t *testing.T, db *pgtest.Database) *sql.DB {
	t.Helper()
	app := seedIncomeOwners(t, db)
	seedTaxpayer(t, app, userA, taxInputsMoralTaxpayer, "moral")
	_, err := execAsUser(t, app, userA, insertRegimeSQL, userA, taxInputsMoralTaxpayer, "601", "moral", "2022-01", nil)
	require.NoError(t, err)
	return app
}

// catalogKind returns the kind of input key in the catalog row (code, personType).
func catalogKind(t *testing.T, db *sql.DB, code, personType, key string) string {
	t.Helper()
	var kind sql.NullString
	require.NoError(t, db.QueryRowContext(t.Context(), `
		SELECT d->>'kind' FROM regimes, jsonb_array_elements(required_inputs) d
		WHERE code = $1 AND person_type = $2 AND d->>'key' = $3`, code, personType, key).Scan(&kind))
	return kind.String
}

// §9 item 6 / OQ-T1: prior provisional payments are captured per month, and 601 gains
// first_fiscal_year. 622 keeps the 601 definitions (§2.8), so its key has the same kind.
func TestRegimes_TaxInputKindsFollowOQT1(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	for _, regime := range []struct{ code, personType string }{
		{"612", "fisica"}, {"626", "moral"}, {"601", "moral"}, {"622", "moral"},
	} {
		assert.Equal(t, "money_monthly", catalogKind(t, owner, regime.code, regime.personType, "prior_provisional_payments"),
			regime.code+"/"+regime.personType)
	}
	assert.Empty(t, queryStrings(t, owner, `
		SELECT code || '/' || person_type FROM regimes, jsonb_array_elements(required_inputs) d
		WHERE d->>'kind' = 'money_monthly' AND d->>'key' <> 'prior_provisional_payments'`),
		"only the prior provisional payments are monthly")

	var firstYear string
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT d::text FROM regimes, jsonb_array_elements(required_inputs) WITH ORDINALITY e(d, n)
		WHERE code = '601' AND person_type = 'moral' AND d->>'key' = 'first_fiscal_year' AND n = 2`).Scan(&firstYear))
	assert.JSONEq(t, `{"key": "first_fiscal_year", "label": "Primer ejercicio fiscal", "kind": "boolean",
		"required": false}`, firstYear, "second, after coeficiente_utilidad (contract example regime601)")
	assert.Equal(t, []string{"601/moral"}, queryStrings(t, owner, `
		SELECT code || '/' || person_type FROM regimes WHERE required_inputs @> '[{"key": "first_fiscal_year"}]'`))

	t.Run("the shape check admits every contract kind and nothing else", func(t *testing.T) {
		for _, kind := range []string{"money", "money_monthly", "rate", "boolean"} {
			var ok bool
			require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT regime_required_inputs_valid($1::jsonb)",
				`[{"key": "x", "label": "X", "kind": "`+kind+`", "required": true}]`).Scan(&ok))
			assert.True(t, ok, kind)
		}
		requireSQLState(t, insertRegime(t, owner, `[]`, `[{"key": "x", "label": "X", "kind": "integer", "required": true}]`),
			sqlstateCheckViolation)
	})
}

// S-1: ENABLE + FORCE RLS, one owner policy per operation the app runs.
func TestTaxInputs_AppRoleIsolation(t *testing.T) {
	app := seedTaxInputOwners(t, pgtest.NewDatabase(t))

	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table: taxInputsTable, UserA: userA, UserB: userB, Insert: insertOwnTaxInput,
	})
}

// AssertIsolation's generic UPDATE rewrites user_id, which cc_app may not update at all;
// prove the writes PUT runs (replace: delete then insert, or an upsert of a value) match
// no foreign row.
func TestTaxInputs_CrossUserWritesMatchNoRow(t *testing.T) {
	app := seedTaxInputOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertTaxInput(t, app, userB, monthlyPayment(userB, 1)))
	writes := []string{
		"UPDATE taxpayer_tax_inputs SET value_cents = 1 WHERE taxpayer_id = $1",
		"DELETE FROM taxpayer_tax_inputs WHERE taxpayer_id = $1 AND regime_code = '612' AND fiscal_year = 2026",
	}
	for _, query := range writes {
		t.Run(query, func(t *testing.T) {
			affected, err := execAsUser(t, app, userA, query, taxpayerB)
			require.NoError(t, err)
			assert.Zero(t, affected)
		})
	}
	t.Run("own rows are writable", func(t *testing.T) {
		for _, query := range writes {
			affected, err := execAsUser(t, app, userB, query, taxpayerB)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// S-1: the composite FK (user_id, taxpayer_id, regime_code) -> taxpayer_regimes makes a
// value for another user's taxpayer, or for a regime the taxpayer never held, impossible
// even with forged ids. The kind trigger runs first and answers what RLS or the FK would.
func TestTaxInputs_CompositeForeignKeyRefusesForgedReferences(t *testing.T) {
	app := seedTaxInputOwners(t, pgtest.NewDatabase(t))
	forged := func(userID, taxpayerID, regime string) taxInputRow {
		r := annualMoney(userA)
		r.userID, r.taxpayerID, r.regime = userID, taxpayerID, regime
		return r
	}
	cases := []struct {
		name string
		row  taxInputRow
		want string
	}{
		{"own user id, another user's taxpayer", forged(userA, taxpayerB, "612"), sqlstateForeignKeyViolation},
		{"another user's id and taxpayer", forged(userB, taxpayerB, "612"), sqlstateInsufficientPrivilege},
		{"another user's id, own taxpayer", forged(userB, taxpayerA, "612"), sqlstateInsufficientPrivilege},
		{"own taxpayer, regime it does not hold", forged(userA, taxpayerA, "606"), sqlstateForeignKeyViolation},
		{"own taxpayer, unknown regime", forged(userA, taxpayerA, "999"), sqlstateForeignKeyViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertTaxInput(t, app, userA, c.row), c.want)
		})
	}
	t.Run("the well-formed reference is accepted", func(t *testing.T) {
		require.NoError(t, insertTaxInput(t, app, userA, annualMoney(userA)))
	})
}

// OpenAPI TaxInputValue: exactly the value field of the key's kind is set, month only
// for money_monthly; money >= 0 up to NonNegativeCents.maximum; a rate is a DecimalRate
// in [0, 1]; one value per key (and month); the key belongs to the regime's catalog row.
func TestTaxInputs_RejectInvalidRows(t *testing.T) {
	app := seedTaxInputOwners(t, pgtest.NewDatabase(t))
	monthly := func(edit func(*taxInputRow)) taxInputRow {
		r := monthlyPayment(userA, 3)
		edit(&r)
		return r
	}
	money := func(edit func(*taxInputRow)) taxInputRow {
		r := annualMoney(userA)
		edit(&r)
		return r
	}
	moral := func(key string, edit func(*taxInputRow)) taxInputRow {
		r := taxInputRow{userID: userA, taxpayerID: taxInputsMoralTaxpayer, regime: "601", fiscalYear: 2026, key: key}
		edit(&r)
		return r
	}
	cases := []struct {
		name string
		row  taxInputRow
	}{
		{"money with a month", money(func(r *taxInputRow) { r.month = 1 })},
		{"money without a value", money(func(r *taxInputRow) { r.cents = nil })},
		{"money with a rate too", money(func(r *taxInputRow) { r.rate = "0.5" })},
		{"money with a boolean too", money(func(r *taxInputRow) { r.flag = true })},
		{"money as a rate", money(func(r *taxInputRow) { r.cents, r.rate = nil, "0.5" })},
		{"negative money", money(func(r *taxInputRow) { r.cents = int64(-1) })},
		{"money above NonNegativeCents.maximum", money(func(r *taxInputRow) { r.cents = int64(100000000000000) })},
		{"monthly without a month", monthly(func(r *taxInputRow) { r.month = nil })},
		{"monthly month 0", monthly(func(r *taxInputRow) { r.month = 0 })},
		{"monthly month 13", monthly(func(r *taxInputRow) { r.month = 13 })},
		{"monthly without a value", monthly(func(r *taxInputRow) { r.cents = nil })},
		{"monthly with a boolean too", monthly(func(r *taxInputRow) { r.flag = false })},
		{"rate without a value", moral("coeficiente_utilidad", func(*taxInputRow) {})},
		{"rate above 1", moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate = "1.000001" })},
		{"negative rate", moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate = "-0.000001" })},
		{"rate NaN", moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate = "NaN" })},
		{"rate with 7 decimals", moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate = "0.1234567" })},
		{"rate as money", moral("coeficiente_utilidad", func(r *taxInputRow) { r.cents = int64(15) })},
		{"rate with a month", moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate, r.month = "0.15", 1 })},
		{"boolean without a value", moral("first_fiscal_year", func(*taxInputRow) {})},
		{"boolean as money", moral("first_fiscal_year", func(r *taxInputRow) { r.cents = int64(0) })},
		{"boolean with a rate too", moral("first_fiscal_year", func(r *taxInputRow) { r.flag, r.rate = true, "0" })},
		{"key outside the catalog", money(func(r *taxInputRow) { r.key = "isr_favor" })},
		{"key of another regime", money(func(r *taxInputRow) { r.key = "coeficiente_utilidad" })},
		{"fiscal year before the catalog floor", money(func(r *taxInputRow) { r.fiscalYear = 2021 })},
		{"fiscal year after FiscalYear.maximum", money(func(r *taxInputRow) { r.fiscalYear = 2101 })},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertTaxInput(t, app, userA, c.row), sqlstateCheckViolation)
		})
	}

	t.Run("one value per key, month and fiscal year", func(t *testing.T) {
		require.NoError(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)))
		requireSQLState(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)), sqlstateUniqueViolation)
		require.NoError(t, insertTaxInput(t, app, userA, annualMoney(userA)))
		// An annual key (month NULL) is unique too: NULLS NOT DISTINCT.
		requireSQLState(t, insertTaxInput(t, app, userA, annualMoney(userA)), sqlstateUniqueViolation)
	})
	t.Run("the boundary values are accepted", func(t *testing.T) {
		accepted := []taxInputRow{
			monthly(func(r *taxInputRow) { r.month = 12 }),
			monthly(func(r *taxInputRow) { r.month, r.fiscalYear = 1, 2027 }),
			money(func(r *taxInputRow) { r.cents, r.fiscalYear = int64(99999999999999), 2022 }),
			money(func(r *taxInputRow) { r.key, r.fiscalYear = "prior_year_losses", 2100 }),
			moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate = "0.1500" }),
			moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate, r.fiscalYear = "1.000000", 2027 }),
			moral("coeficiente_utilidad", func(r *taxInputRow) { r.rate, r.fiscalYear = "0", 2028 }),
			moral("first_fiscal_year", func(r *taxInputRow) { r.flag = false }),
			moral("prior_provisional_payments", func(r *taxInputRow) { r.month, r.cents = 2, int64(1) }),
		}
		for _, r := range accepted {
			require.NoError(t, insertTaxInput(t, app, userA, r), "%+v", r)
		}
	})
	t.Run("a rate keeps the scale it was given", func(t *testing.T) {
		var rate string
		queryAsUser(t, app, userA, `SELECT value_rate::text FROM taxpayer_tax_inputs
			WHERE input_key = 'coeficiente_utilidad' AND fiscal_year = 2026`, &rate)
		assert.Equal(t, "0.1500", rate)
	})
	t.Run("a value of the wrong kind cannot be written by update either", func(t *testing.T) {
		_, err := execAsUser(t, app, userA,
			"UPDATE taxpayer_tax_inputs SET value_cents = NULL, value_rate = 0.5 WHERE input_key = 'ptu_paid'")
		requireSQLState(t, err, sqlstateCheckViolation)
	})
}

// kind is derived from regimes.required_inputs on every write; cc_app cannot write it.
func TestTaxInputs_KindFollowsTheCatalog(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedTaxInputOwners(t, db)
	require.NoError(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)))
	require.NoError(t, insertTaxInput(t, app, userA, annualMoney(userA)))
	require.NoError(t, insertTaxInput(t, app, userA, taxInputRow{
		userID: userA, taxpayerID: taxInputsMoralTaxpayer, regime: "601", fiscalYear: 2026,
		key: "first_fiscal_year", flag: true,
	}))

	assert.Equal(t, []string{"first_fiscal_year=boolean", "prior_provisional_payments=money_monthly", "ptu_paid=money"},
		func() []string {
			var out []string
			require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
				rows, err := tx.QueryContext(t.Context(),
					"SELECT input_key || '=' || kind FROM taxpayer_tax_inputs ORDER BY input_key")
				if err != nil {
					return err
				}
				defer func() { _ = rows.Close() }()
				for rows.Next() {
					var s string
					if err := rows.Scan(&s); err != nil {
						return err
					}
					out = append(out, s)
				}
				return rows.Err()
			}))
			return out
		}())

	for _, stmt := range []string{
		"UPDATE taxpayer_tax_inputs SET kind = 'money'",
		"INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, kind, value_cents) " +
			"VALUES ('" + userA + "', '" + taxpayerA + "', '612', 2027, 'ptu_paid', 'money', 0)",
	} {
		t.Run("cc_app: "+stmt, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, stmt)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}

	// A writer outside cc_app's policies (the owner running a future catalog migration)
	// that cannot see the regime row is refused, never handed a guessed kind.
	owner := db.Connect(t, pgtest.RoleOwner)
	_, err := owner.ExecContext(t.Context(), "ALTER TABLE taxpayer_tax_inputs NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	_, err = owner.ExecContext(t.Context(), "UPDATE taxpayer_tax_inputs SET kind = 'rate' WHERE input_key = 'ptu_paid'")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
}

// S-1 / US-A4 / §7: account and taxpayer deletion cascade to the inputs, flag OFF or ON.
func TestTaxInputs_DeletionCascades(t *testing.T) {
	setup := func(t *testing.T) *sql.DB {
		t.Helper()
		app := seedTaxInputOwners(t, pgtest.NewDatabase(t))
		require.NoError(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)))
		require.NoError(t, insertTaxInput(t, app, userB, monthlyPayment(userB, 1)))
		return app
	}
	count := func(t *testing.T, app *sql.DB, userID string) int {
		t.Helper()
		var n int
		queryAsUser(t, app, userID, "SELECT count(*) FROM taxpayer_tax_inputs", &n)
		return n
	}
	t.Run("account deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		assert.Zero(t, count(t, app, userA))
		assert.Equal(t, 1, count(t, app, userB), "another user's inputs stay")
	})
	t.Run("taxpayer deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM taxpayers WHERE id = $1", taxpayerA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		assert.Zero(t, count(t, app, userA))
	})
	t.Run("the users-FK guard sees taxpayer_tax_inputs directly and through its chain", func(t *testing.T) {
		owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
		assert.Contains(t, queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'users'::regclass AND confdeltype = 'c'"),
			"taxpayer_tax_inputs")
		assert.Contains(t, queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'taxpayer_regimes'::regclass AND confdeltype = 'c'"),
			"taxpayer_tax_inputs")
	})
}

// Least privilege: only cc_app reaches the inputs, on the columns PUT writes. The worker
// computes no estimate yet (M9 adds its own grant and policy when it does).
func TestTaxInputs_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedTaxInputOwners(t, db)
	require.NoError(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)))
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET user_id = user_id"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET taxpayer_id = taxpayer_id"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET regime_code = regime_code"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET fiscal_year = fiscal_year"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET input_key = input_key"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET month = month"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET created_at = now()"},
		{pgtest.RoleApp, "UPDATE taxpayer_tax_inputs SET updated_at = now()"},
		{pgtest.RoleApp, "TRUNCATE taxpayer_tax_inputs"},
		{pgtest.RoleApp, "ALTER TABLE taxpayer_tax_inputs DISABLE TRIGGER taxpayer_tax_inputs_derive_kind"},
		{pgtest.RoleApp, "CREATE OR REPLACE FUNCTION public.taxpayer_tax_inputs_derive_kind() RETURNS trigger LANGUAGE plpgsql AS 'BEGIN RETURN NEW; END'"},
		{pgtest.RoleApp, "UPDATE regimes SET required_inputs = '[]'"},
		{pgtest.RoleWebhook, "SELECT 1 FROM taxpayer_tax_inputs"},
		{pgtest.RoleWorker, "SELECT 1 FROM taxpayer_tax_inputs"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
	t.Run("cc_app writes what PUT needs", func(t *testing.T) {
		for _, query := range []string{
			"UPDATE taxpayer_tax_inputs SET value_cents = 250000, value_rate = NULL, value_bool = NULL",
			`INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month, value_cents)
			 VALUES ('` + userA + `', '` + taxpayerA + `', '612', 2026, 'prior_provisional_payments', 1, 300000)
			 ON CONFLICT (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month)
			 DO UPDATE SET value_cents = EXCLUDED.value_cents`,
			"DELETE FROM taxpayer_tax_inputs WHERE regime_code = '612' AND fiscal_year = 2026",
		} {
			affected, err := execAsUser(t, app, userA, query)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// updated_at is now() of the writing transaction (set_updated_at).
func TestTaxInputs_UpdatedAtFollowsWrites(t *testing.T) {
	app := seedTaxInputOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)))
	var followed bool
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.01)"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(t.Context(), "UPDATE taxpayer_tax_inputs SET value_cents = 1"); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(),
			"SELECT bool_and(updated_at = now() AND updated_at > created_at) FROM taxpayer_tax_inputs").Scan(&followed)
	}))
	assert.True(t, followed)
}

// The derivation trigger is an invoker function with a pinned search_path.
func TestTaxInputsTrigger_IsInvokerWithPinnedSearchPath(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	var definer bool
	var config sql.NullString
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT prosecdef, array_to_string(proconfig, ',') FROM pg_proc WHERE oid = 'taxpayer_tax_inputs_derive_kind()'::regprocedure").
		Scan(&definer, &config))
	assert.False(t, definer, "SECURITY INVOKER")
	assert.Equal(t, "search_path=pg_catalog, pg_temp", config.String)
}

// planTaxInputsSQL writes, for user $1 / taxpayer $2 (612), every fiscal year 2022..2100:
// twelve monthly prior payments and three annual money inputs.
var planTaxInputsSQL = fmt.Sprintf(`
INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month, value_cents)
SELECT $1::uuid, $2::uuid, '612', y, 'prior_provisional_payments', m, 100000 * m
FROM generate_series(%[1]d, %[2]d) AS y, generate_series(1, 12) AS m
UNION ALL
SELECT $1::uuid, $2::uuid, '612', y, k, NULL, 0
FROM generate_series(%[1]d, %[2]d) AS y,
     unnest(ARRAY['ptu_paid', 'prior_year_losses', 'ytd_before_app_income']) AS k`,
	planTaxInputFirstYear, planTaxInputLastYear)

// The estimate read path (TaxInputReader: one taxpayer + regime + fiscal year) and the
// cascades are served by the unique key's btree, under RLS as cc_app.
func TestTaxInputs_KeyIndexServesTheEstimateReadPath(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	for i := range planUsers {
		user, taxpayer := planUserID(i), planTaxpayerID(i)
		seedUsers(t, app, user)
		seedTaxpayer(t, app, user, taxpayer, "fisica")
		_, err := execAsUser(t, app, user, insertRegimeSQL, user, taxpayer, "612", "fisica", "2022-01", nil)
		require.NoError(t, err)
		_, err = execAsUser(t, app, user, planTaxInputsSQL, user, taxpayer)
		require.NoError(t, err)
	}
	owner := db.Connect(t, pgtest.RoleOwner)
	_, err := owner.ExecContext(t.Context(), "ANALYZE taxpayer_tax_inputs")
	require.NoError(t, err)

	defs := queryStrings(t, owner,
		"SELECT regexp_replace(indexdef, '^.* USING ', '') FROM pg_indexes WHERE tablename = 'taxpayer_tax_inputs' ORDER BY 1")
	assert.Equal(t, []string{
		"btree (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month) NULLS NOT DISTINCT",
	}, defs, "one btree serves RLS, the read path, the upsert and both cascades")

	user, taxpayer := planUserID(7), planTaxpayerID(7)
	plan := explain(t, app, user, fmt.Sprintf(`
		SELECT input_key, month, kind, value_cents, value_rate, value_bool FROM taxpayer_tax_inputs
		WHERE taxpayer_id = '%s' AND regime_code = '612' AND fiscal_year = 2026
		ORDER BY input_key, month`, taxpayer))
	assert.Contains(t, plan, "taxpayer_tax_inputs_key")
	assert.NotContains(t, plan, "Seq Scan on taxpayer_tax_inputs")
}

func TestTaxInputsMigrations_DownRemovesEverything(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), taxInputKindsVersion-1)
	require.NoError(t, err)
	for _, obj := range []string{
		"to_regclass('public.taxpayer_tax_inputs')",
		"to_regprocedure('public.taxpayer_tax_inputs_derive_kind()')",
	} {
		var found sql.NullString
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT "+obj+"::text").Scan(&found))
		assert.False(t, found.Valid, "down left %s", found.String)
	}
	assert.Equal(t, "money", catalogKind(t, owner, "612", "fisica", "prior_provisional_payments"))
	assert.Empty(t, queryStrings(t, owner,
		`SELECT code FROM regimes WHERE required_inputs @> '[{"key": "first_fiscal_year"}]'`))
	assert.Equal(t, []string{"coeficiente_utilidad", "prior_provisional_payments", "ptu_paid", "prior_year_losses",
		"interest_withholding"}, jsonKeys(t, owner, "required_inputs")["601/moral"], "the M2-02 order is restored")
	requireSQLState(t, insertRegime(t, owner, `[]`, `[{"key": "x", "label": "X", "kind": "boolean", "required": true}]`),
		sqlstateCheckViolation)

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	assert.Equal(t, "money_monthly", catalogKind(t, owner, "612", "fisica", "prior_provisional_payments"))
	var exists bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT to_regclass('public.taxpayer_tax_inputs') IS NOT NULL").Scan(&exists))
	assert.True(t, exists)
}
