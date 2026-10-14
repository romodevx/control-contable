//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Incomes migration (M3-02): the incomes table and regimes.income_basis,
// docs/project-specs/MODULE-incomes.md §2 and §5 (S-1, S-2, S-9) and
// docs/architecture/data-model.md.

const (
	incomesVersion = 20261013200000

	incomeA = "00000000-0000-7000-8000-00000000c0a1"
	incomeB = "00000000-0000-7000-8000-00000000c0b1"

	// incomeMoralTaxpayer is a persona moral of userA (accrual and cash regimes).
	incomeMoralTaxpayer = "00000000-0000-7000-8000-0000000000a2"
)

var incomesTable = pgtest.Table{Name: "incomes", OwnerColumn: "user_id"}

// insertIncomeSQL writes income $1 of user $2, taxpayer $3, regime $4. Every column the
// app may write is a parameter; booked_on is not: the database derives it.
const insertIncomeSQL = `
INSERT INTO incomes (id, user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
                     isr_withheld_cents, iva_withheld_cents, invoiced_on, collected_on, platform_activity,
                     cfdi_uuid, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

// incomeRow is one insertIncomeSQL call; validIncome is US-I1's 612 income.
type incomeRow struct {
	id, userID, taxpayerID, regime, counterparty string
	amount                                       int64
	ivaRate                                      string
	iva, isrWithheld, ivaWithheld                int64
	invoicedOn, collectedOn                      any
	platformActivity, cfdiUUID, description      any
}

func validIncome(userID, taxpayerID, id string) incomeRow {
	return incomeRow{
		id: id, userID: userID, taxpayerID: taxpayerID, regime: "612", counterparty: "persona_moral",
		amount: 1000000, ivaRate: "rate_16", iva: 160000, isrWithheld: 100000, ivaWithheld: 106667,
		invoicedOn: "2026-09-05", collectedOn: "2026-09-10",
		cfdiUUID: "6F1C2A3B-4D5E-4F60-8A7B-9C0D1E2F3A4B", description: "Consultoría septiembre",
	}
}

func (r incomeRow) args() []any {
	return []any{
		r.id, r.userID, r.taxpayerID, r.regime, r.counterparty, r.amount, r.ivaRate, r.iva, r.isrWithheld,
		r.ivaWithheld, r.invoicedOn, r.collectedOn, r.platformActivity, r.cfdiUUID, r.description,
	}
}

func insertIncome(t *testing.T, app *sql.DB, asUser string, r incomeRow) error {
	t.Helper()
	_, err := execAsUser(t, app, asUser, insertIncomeSQL, r.args()...)
	return err
}

var incomeOf = map[string]string{userA: incomeA, userB: incomeB}

func insertOwnIncome(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertIncomeSQL, validIncome(ownerID, taxpayerOf[ownerID], incomeOf[ownerID]).args()...)
	return err
}

// seedIncomeOwners creates users A and B, each with one física taxpayer holding 612
// since 2022-01, and returns cc_app's connection.
func seedIncomeOwners(t *testing.T, db *pgtest.Database) *sql.DB {
	t.Helper()
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	for _, owner := range []string{userA, userB} {
		seedTaxpayer(t, app, owner, taxpayerOf[owner], "fisica")
		_, err := execAsUser(t, app, owner, insertRegimeSQL, owner, taxpayerOf[owner], "612", "fisica", "2022-01", nil)
		require.NoError(t, err)
	}
	return app
}

func queryAsUser(t *testing.T, app *sql.DB, userID, query string, dest ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), app, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), query).Scan(dest...)
	}))
}

// S-1: ENABLE + FORCE RLS, one owner policy per operation the app runs.
func TestIncomes_AppRoleIsolation(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))

	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table: incomesTable, UserA: userA, UserB: userB, Insert: insertOwnIncome,
	})
}

// AssertIsolation's generic UPDATE rewrites user_id, which cc_app may not update at all;
// prove the writes the use cases run (edit, soft delete, restore) match no foreign row.
func TestIncomes_CrossUserWritesMatchNoRow(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertIncome(t, app, userB, validIncome(userB, taxpayerB, incomeB)))
	writes := []string{
		"UPDATE incomes SET amount_cents = 1, iva_cents = 0, iva_rate = 'exempt', isr_withheld_cents = 0, iva_withheld_cents = 0 WHERE id = $1",
		"UPDATE incomes SET deleted_at = now() WHERE id = $1",
		"UPDATE incomes SET deleted_at = NULL WHERE id = $1",
		"UPDATE incomes SET description = 'Mallory' WHERE id = $1",
	}
	for _, query := range writes {
		t.Run(query, func(t *testing.T) {
			affected, err := execAsUser(t, app, userA, query, incomeB)
			require.NoError(t, err)
			assert.Zero(t, affected)
		})
	}
	t.Run("own rows are writable", func(t *testing.T) {
		for _, query := range writes {
			affected, err := execAsUser(t, app, userB, query, incomeB)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// S-2: the composite FK (user_id, taxpayer_id, regime_code) -> taxpayer_regimes makes a
// cross-user or unassigned-regime income impossible even with forged ids (FK checks
// bypass RLS, so the FK, not the policy, proves the triple belongs together).
func TestIncomes_CompositeForeignKeyRefusesForgedReferences(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	forged := func(userID, taxpayerID, regime string) incomeRow {
		r := validIncome(userID, taxpayerID, "00000000-0000-7000-8000-00000000f001")
		r.regime = regime
		return r
	}
	cases := []struct {
		name string
		row  incomeRow
		want string
	}{
		{"own user id, another user's taxpayer", forged(userA, taxpayerB, "612"), sqlstateForeignKeyViolation},
		{"another user's id and taxpayer", forged(userB, taxpayerB, "612"), sqlstateInsufficientPrivilege},
		{"own taxpayer, regime it does not hold", forged(userA, taxpayerA, "606"), sqlstateForeignKeyViolation},
		{"own taxpayer, unknown regime", forged(userA, taxpayerA, "999"), sqlstateForeignKeyViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertIncome(t, app, userA, c.row), c.want)
		})
	}
	t.Run("the well-formed reference is accepted", func(t *testing.T) {
		require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	})
	t.Run("moving an income to a regime the taxpayer does not hold is refused", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, "UPDATE incomes SET regime_code = '606' WHERE id = $1", incomeA)
		requireSQLState(t, err, sqlstateForeignKeyViolation)
	})
}

// S-2 CHECKs and the column formats of MODULE-incomes §2.1.
func TestIncomes_RejectInvalidRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	_, err := execAsUser(t, app, userA, insertRegimeSQL, userA, taxpayerA, "625", "fisica", "2022-01", nil)
	require.NoError(t, err)
	with := func(edit func(*incomeRow)) incomeRow {
		r := validIncome(userA, taxpayerA, incomeA)
		edit(&r)
		return r
	}
	cases := []struct {
		name string
		row  incomeRow
		want string
	}{
		{"zero amount", with(func(r *incomeRow) { r.amount = 0 }), sqlstateCheckViolation},
		{"amount above NonNegativeCents.maximum", with(func(r *incomeRow) { r.amount = 100000000000000 }), sqlstateCheckViolation},
		{"negative IVA", with(func(r *incomeRow) { r.iva, r.ivaWithheld = -1, 0 }), sqlstateCheckViolation},
		{"negative ISR withheld", with(func(r *incomeRow) { r.isrWithheld = -1 }), sqlstateCheckViolation},
		{"negative IVA withheld", with(func(r *incomeRow) { r.ivaWithheld = -1 }), sqlstateCheckViolation},
		{"ISR withheld above the amount", with(func(r *incomeRow) { r.isrWithheld = r.amount + 1 }), sqlstateCheckViolation},
		{"IVA withheld above the IVA", with(func(r *incomeRow) { r.ivaWithheld = r.iva + 1 }), sqlstateCheckViolation},
		{"unknown IVA rate", with(func(r *incomeRow) { r.ivaRate = "rate_11" }), sqlstateCheckViolation},
		{"exempt with IVA", with(func(r *incomeRow) { r.ivaRate, r.ivaWithheld = "exempt", 0 }), sqlstateCheckViolation},
		{"rate 0 with IVA", with(func(r *incomeRow) { r.ivaRate, r.ivaWithheld = "rate_0", 0 }), sqlstateCheckViolation},
		{"unknown counterparty", with(func(r *incomeRow) { r.counterparty = "gobierno" }), sqlstateCheckViolation},
		{"withholding from a persona física", with(func(r *incomeRow) { r.counterparty, r.ivaWithheld = "persona_fisica", 0 }), sqlstateCheckViolation},
		{"withholding from público en general", with(func(r *incomeRow) { r.counterparty, r.isrWithheld = "publico_general", 0 }), sqlstateCheckViolation},
		{"platform activity outside 625", with(func(r *incomeRow) { r.platformActivity = "lodging" }), sqlstateCheckViolation},
		{"unknown platform activity", with(func(r *incomeRow) {
			r.regime, r.counterparty, r.platformActivity = "625", "plataforma", "mining"
		}), sqlstateCheckViolation},
		{"lower-case CFDI UUID", with(func(r *incomeRow) { r.cfdiUUID = "6f1c2a3b-4d5e-4f60-8a7b-9c0d1e2f3a4b" }), sqlstateCheckViolation},
		{"malformed CFDI UUID", with(func(r *incomeRow) { r.cfdiUUID = "6F1C2A3B4D5E4F608A7B9C0D1E2F3A4B" }), sqlstateCheckViolation},
		{"empty description", with(func(r *incomeRow) { r.description = "" }), sqlstateCheckViolation},
		{"untrimmed description", with(func(r *incomeRow) { r.description = " Renta" }), sqlstateCheckViolation},
		{"description over 200 characters", with(func(r *incomeRow) { r.description = strings.Repeat("a", 201) }), sqlstateCheckViolation},
		{"description not NFC", with(func(r *incomeRow) { r.description = "Jose\u0301" }), sqlstateCheckViolation},
		{"invoiced before 2022", with(func(r *incomeRow) { r.invoicedOn = "2021-12-31" }), sqlstateCheckViolation},
		{"collected before 2022", with(func(r *incomeRow) { r.collectedOn = "2021-12-31" }), sqlstateCheckViolation},
		{"cash regime without its collection date", with(func(r *incomeRow) { r.collectedOn = nil }), sqlstateNotNullViolation},
		{"no counterparty", with(func(r *incomeRow) { r.counterparty = "" }), sqlstateCheckViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, insertIncome(t, app, userA, c.row), c.want)
		})
	}
	t.Run("the boundary values are accepted", func(t *testing.T) {
		accepted := []incomeRow{
			with(func(r *incomeRow) {
				r.id, r.amount, r.isrWithheld = "00000000-0000-7000-8000-00000000b001", 99999999999999, 99999999999999
			}),
			with(func(r *incomeRow) {
				r.id, r.description, r.cfdiUUID = "00000000-0000-7000-8000-00000000b002", strings.Repeat("ñ", 200), nil
			}),
			with(func(r *incomeRow) {
				r.id, r.ivaRate, r.iva, r.ivaWithheld = "00000000-0000-7000-8000-00000000b003", "exempt", 0, 0
				r.counterparty, r.isrWithheld, r.description = "publico_general", 0, nil
				r.invoicedOn, r.collectedOn = nil, "2022-01-01"
			}),
			with(func(r *incomeRow) {
				r.id, r.regime, r.counterparty = "00000000-0000-7000-8000-00000000b004", "625", "plataforma"
				r.platformActivity = "transport_delivery"
			}),
		}
		for _, r := range accepted {
			require.NoError(t, insertIncome(t, app, userA, r), r.id)
		}
	})

	// S-2 lists CHECK (collected_on IS NOT NULL OR invoiced_on IS NOT NULL). Through the
	// app the derived booked_on NOT NULL fires first, so prove the CHECK on its own: the
	// owner, with FORCE lifted on this throwaway database and the trigger off, writes
	// booked_on itself.
	t.Run("an income with no date at all", func(t *testing.T) {
		owner := db.Connect(t, pgtest.RoleOwner)
		for _, ddl := range []string{
			"ALTER TABLE incomes NO FORCE ROW LEVEL SECURITY",
			"ALTER TABLE incomes DISABLE TRIGGER incomes_derive_booked_on",
		} {
			_, err := owner.ExecContext(t.Context(), ddl)
			require.NoError(t, err)
		}
		_, err := owner.ExecContext(t.Context(), `
INSERT INTO incomes (user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents, booked_on)
VALUES ($1, $2, '612', 'persona_fisica', 100, 'exempt', 0, '2026-09-01')`, userA, taxpayerA)
		requireSQLState(t, err, sqlstateCheckViolation)
	})
}

// MODULE-incomes §2.2: the regime's income basis is catalog data, and booked_on (the ISR
// date that decides the list period) is derived from it on every write: invoiced_on for
// accrual regimes, collected_on for cash ones. The app cannot write it.
func TestIncomes_BookedOnFollowsTheRegimeBasis(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	seedTaxpayer(t, app, userA, incomeMoralTaxpayer, "moral")
	for _, regime := range []struct{ code, start, end string }{{"601", "2022-01", "2022-12"}, {"626", "2023-01", ""}} {
		var end any
		if regime.end != "" {
			end = regime.end
		}
		_, err := execAsUser(t, app, userA, insertRegimeSQL, userA, incomeMoralTaxpayer, regime.code, "moral", regime.start, end)
		require.NoError(t, err)
	}
	bookedOn := func(t *testing.T, id string) string {
		t.Helper()
		var got string
		require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(), "SELECT booked_on::text FROM incomes WHERE id = $1", id).Scan(&got)
		}))
		return got
	}

	t.Run("cash regime books on the collection date", func(t *testing.T) {
		require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
		assert.Equal(t, "2026-09-10", bookedOn(t, incomeA))

		_, err := execAsUser(t, app, userA, "UPDATE incomes SET collected_on = '2026-08-31' WHERE id = $1", incomeA)
		require.NoError(t, err)
		assert.Equal(t, "2026-08-31", bookedOn(t, incomeA), "an edit re-derives it")
	})
	const accrualIncome = "00000000-0000-7000-8000-00000000c0a2"
	t.Run("accrual regime books on the invoice date, pending collection included", func(t *testing.T) {
		r := validIncome(userA, incomeMoralTaxpayer, accrualIncome)
		r.regime, r.isrWithheld, r.invoicedOn, r.collectedOn = "601", 0, "2022-03-15", nil
		require.NoError(t, insertIncome(t, app, userA, r))
		assert.Equal(t, "2022-03-15", bookedOn(t, accrualIncome))

		_, err := execAsUser(t, app, userA, "UPDATE incomes SET collected_on = '2022-05-02' WHERE id = $1", accrualIncome)
		require.NoError(t, err)
		assert.Equal(t, "2022-03-15", bookedOn(t, accrualIncome), "collecting it does not move the ISR date")
	})
	t.Run("moving to a cash regime re-derives it", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, "UPDATE incomes SET regime_code = '626' WHERE id = $1", accrualIncome)
		require.NoError(t, err)
		assert.Equal(t, "2022-05-02", bookedOn(t, accrualIncome))
	})
	t.Run("cc_app cannot write it", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, "UPDATE incomes SET booked_on = '2022-01-01' WHERE id = $1", incomeA)
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
	t.Run("a role that may write it is overridden", func(t *testing.T) {
		owner := db.Connect(t, pgtest.RoleOwner)
		_, err := owner.ExecContext(t.Context(), "ALTER TABLE incomes NO FORCE ROW LEVEL SECURITY")
		require.NoError(t, err)
		_, err = owner.ExecContext(t.Context(), "UPDATE incomes SET booked_on = '2022-01-01' WHERE id = $1", incomeA)
		require.NoError(t, err)
		assert.Equal(t, "2026-08-31", bookedOn(t, incomeA))
	})
}

// §2.2: accrual = PM Título II (601, 620, 622, 623, 624); every other regime is cash.
func TestRegimes_IncomeBasisCoversTheCatalog(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	accrual := queryStrings(t, owner,
		"SELECT code || '/' || person_type FROM regimes WHERE income_basis = 'accrual' ORDER BY 1")
	assert.Equal(t, []string{"601/moral", "620/moral", "622/moral", "623/moral", "624/moral"}, accrual)
	assert.Empty(t, queryStrings(t, owner,
		"SELECT code || '/' || person_type FROM regimes WHERE income_basis IS DISTINCT FROM 'accrual' AND income_basis IS DISTINCT FROM 'cash'"))

	_, err := owner.ExecContext(t.Context(), "UPDATE regimes SET income_basis = 'mixed' WHERE code = '612'")
	requireSQLState(t, err, sqlstateCheckViolation)
	_, err = owner.ExecContext(t.Context(), "UPDATE regimes SET income_basis = NULL WHERE code = '612'")
	requireSQLState(t, err, sqlstateNotNullViolation)
}

// OQ-9: a duplicate CFDI UUID is a warning computed by the use case, never a constraint
// (partial collections of one PPD invoice share it).
func TestIncomes_DuplicateCFDIUUIDIsAllowed(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	second := validIncome(userA, taxpayerA, "00000000-0000-7000-8000-00000000c0a3")
	require.NoError(t, insertIncome(t, app, userA, second))
	require.NoError(t, insertIncome(t, app, userB, validIncome(userB, taxpayerB, incomeB)))
}

// §2.3: the client id is the primary key, so a retry with the same id collides, even
// across users (RLS hides the row; the PK does not).
func TestIncomes_ClientIDIsTheIdempotencyKey(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	requireSQLState(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)), sqlstateUniqueViolation)
	requireSQLState(t, insertIncome(t, app, userB, validIncome(userB, taxpayerB, incomeA)), sqlstateUniqueViolation)
}

// S-9 / US-A4 and MODULE-taxpayers §2.7: account and taxpayer deletion cascade to incomes.
func TestIncomes_DeletionCascades(t *testing.T) {
	setup := func(t *testing.T) *sql.DB {
		t.Helper()
		app := seedIncomeOwners(t, pgtest.NewDatabase(t))
		require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
		return app
	}
	countIncomes := func(t *testing.T, app *sql.DB) int {
		t.Helper()
		var n int
		queryAsUser(t, app, userA, "SELECT count(*) FROM incomes", &n)
		return n
	}
	t.Run("account deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		assert.Zero(t, countIncomes(t, app))
	})
	t.Run("taxpayer deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM taxpayers WHERE id = $1", taxpayerA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		assert.Zero(t, countIncomes(t, app))
	})
	t.Run("the users-FK guard sees incomes directly and through its chain", func(t *testing.T) {
		owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
		assert.Contains(t, queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'users'::regclass AND confdeltype = 'c'"),
			"incomes")
		assert.Contains(t, queryStrings(t, owner,
			"SELECT conrelid::regclass::text FROM pg_constraint WHERE contype = 'f' AND confrelid = 'taxpayer_regimes'::regclass AND confdeltype = 'c'"),
			"incomes")
	})
}

// Least privilege: only cc_app reaches incomes, and only on the columns its use cases
// write. Delete is soft (an UPDATE of deleted_at); rows leave by cascade or by the
// purge definer function (M9-08), never by a DELETE from the API.
func TestIncomes_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE incomes SET id = uuidv7()"},
		{pgtest.RoleApp, "UPDATE incomes SET user_id = user_id"},
		{pgtest.RoleApp, "UPDATE incomes SET taxpayer_id = taxpayer_id"},
		{pgtest.RoleApp, "UPDATE incomes SET booked_on = booked_on"},
		{pgtest.RoleApp, "UPDATE incomes SET created_at = now()"},
		{pgtest.RoleApp, "UPDATE incomes SET updated_at = now()"},
		{pgtest.RoleApp, "INSERT INTO incomes (user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents, collected_on, deleted_at) " +
			"VALUES ('" + userA + "', '" + taxpayerA + "', '612', 'persona_fisica', 1, 'exempt', 0, '2026-09-01', now())"},
		{pgtest.RoleApp, "DELETE FROM incomes"},
		{pgtest.RoleApp, "TRUNCATE incomes"},
		{pgtest.RoleWebhook, "SELECT 1 FROM incomes"},
		{pgtest.RoleWorker, "SELECT 1 FROM incomes"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
	t.Run("cc_app writes what the use cases need", func(t *testing.T) {
		for _, query := range []string{
			`UPDATE incomes SET regime_code = '612', counterparty_type = 'persona_moral', amount_cents = 2000000,
			   iva_rate = 'rate_16', iva_cents = 320000, isr_withheld_cents = 200000, iva_withheld_cents = 213333,
			   invoiced_on = NULL, collected_on = '2026-09-11', platform_activity = NULL,
			   cfdi_uuid = NULL, description = NULL`,
			"UPDATE incomes SET deleted_at = now()",
			"UPDATE incomes SET deleted_at = NULL",
		} {
			affected, err := execAsUser(t, app, userA, query)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// updated_at is now() of the writing transaction (set_updated_at).
func TestIncomes_UpdatedAtFollowsWrites(t *testing.T) {
	app := seedIncomeOwners(t, pgtest.NewDatabase(t))
	require.NoError(t, insertIncome(t, app, userA, validIncome(userA, taxpayerA, incomeA)))
	var followed bool
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.01)"); err != nil {
			return err
		}
		if _, err := tx.ExecContext(t.Context(), "UPDATE incomes SET deleted_at = now()"); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(),
			"SELECT bool_and(updated_at = now() AND updated_at > created_at) FROM incomes").Scan(&followed)
	}))
	assert.True(t, followed)
}

// The derivation trigger is an invoker function with a pinned search_path that cc_app
// can neither switch off nor replace.
func TestIncomesTrigger_CannotBeBypassedByTheAppRole(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	owner := db.Connect(t, pgtest.RoleOwner)
	var definer bool
	var config sql.NullString
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT prosecdef, array_to_string(proconfig, ',') FROM pg_proc WHERE oid = 'incomes_derive_booked_on()'::regprocedure").
		Scan(&definer, &config))
	assert.False(t, definer, "SECURITY INVOKER")
	assert.Equal(t, "search_path=pg_catalog, pg_temp", config.String)

	for _, stmt := range []string{
		"ALTER TABLE incomes DISABLE TRIGGER incomes_derive_booked_on",
		"SET LOCAL session_replication_role = replica",
		"UPDATE regimes SET income_basis = 'accrual'",
		"CREATE OR REPLACE FUNCTION public.incomes_derive_booked_on() RETURNS trigger LANGUAGE plpgsql AS 'BEGIN RETURN NEW; END'",
	} {
		t.Run("cc_app: "+stmt, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, stmt)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// The hot paths (MODULE-incomes §2.9, S-12, OQ-9) each have their partial index.
func TestIncomes_IndexesServeTheHotQueries(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	defs := queryStrings(t, owner,
		"SELECT regexp_replace(indexdef, '^.* USING ', '') FROM pg_indexes WHERE tablename = 'incomes' ORDER BY 1")
	assert.Subset(t, defs, []string{
		"btree (deleted_at) WHERE (deleted_at IS NOT NULL)",
		"btree (id)",
		"btree (user_id, taxpayer_id, booked_on, id) WHERE (deleted_at IS NULL)",
		"btree (user_id, taxpayer_id, cfdi_uuid) WHERE ((cfdi_uuid IS NOT NULL) AND (deleted_at IS NULL))",
		"btree (user_id, taxpayer_id, collected_on, id) WHERE ((collected_on IS NOT NULL) AND (deleted_at IS NULL))",
		"btree (user_id, taxpayer_id, regime_code)",
	})
}

func TestIncomesMigration_DownRemovesEverything(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), incomesVersion-1)
	require.NoError(t, err)
	for _, obj := range []string{
		"to_regclass('public.incomes')",
		"to_regprocedure('public.incomes_derive_booked_on()')",
	} {
		var found sql.NullString
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT "+obj+"::text").Scan(&found))
		assert.False(t, found.Valid, "down left %s", found.String)
	}
	var basis bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'regimes' AND column_name = 'income_basis')").
		Scan(&basis))
	assert.False(t, basis, "down left regimes.income_basis")

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var exists bool
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass('public.incomes') IS NOT NULL").Scan(&exists))
	assert.True(t, exists)
}
