//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Taxpayers migration (M2-03): taxpayers and taxpayer_regimes, docs/project-specs/
// MODULE-taxpayers.md §2 and docs/architecture/data-model.md.

const (
	taxpayersVersion = 20261013190000

	// One taxpayer per user for the isolation suites; the cap test adds more.
	taxpayerA = "00000000-0000-7000-8000-0000000000a1"
	taxpayerB = "00000000-0000-7000-8000-0000000000b1"

	// taxpayerCap is the domain cap (OD-2): the database only provides the lock and the count.
	taxpayerCap = 10

	sqlstateExclusionViolation = "23P01"
	sqlstateNoDataFound        = "P0002"
)

var (
	taxpayersTable       = pgtest.Table{Name: "taxpayers", OwnerColumn: "user_id"}
	taxpayerRegimesTable = pgtest.Table{Name: "taxpayer_regimes", OwnerColumn: "user_id"}

	errCapReached = errors.New("taxpayer cap reached")
)

// insertTaxpayerSQL writes taxpayer $2 of user $1 with person type $3. The RFC columns
// derive from the taxpayer id like crypto output would: a framed AES-GCM ciphertext
// (key_id 1 || nonce || ct || tag), a 32-byte HMAC blind index and the masked form.
const insertTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked)
VALUES ($2::text::uuid, $1, $3, 'Ana (personal)',
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')),
        CASE $3 WHEN 'fisica' THEN 'GOMA******AB6' ELSE 'GOM******AB6' END)`

// insertRegimeSQL adds regime $3 to taxpayer $2 of user $1 (person type $4) from $5 to $6.
const insertRegimeSQL = `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period, end_period)
VALUES ($1, $2, $3, $4, $5, $6)`

var taxpayerOf = map[string]string{userA: taxpayerA, userB: taxpayerB}

func insertOwnTaxpayer(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertTaxpayerSQL, ownerID, taxpayerOf[ownerID], "fisica")
	return err
}

func insertOwnRegime(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertRegimeSQL, ownerID, taxpayerOf[ownerID], "612", "fisica", "2026-01", nil)
	return err
}

// seedTaxpayer creates taxpayer id of userID (person type personType) under userID's context.
func seedTaxpayer(t *testing.T, app *sql.DB, userID, id, personType string) {
	t.Helper()
	_, err := execAsUser(t, app, userID, insertTaxpayerSQL, userID, id, personType)
	require.NoError(t, err)
}

// addRegime adds a regime to taxpayerA of userA; end "" means open-ended.
func addRegime(t *testing.T, app *sql.DB, taxpayerID, personType, code, start, end string) error {
	t.Helper()
	var endPeriod any
	if end != "" {
		endPeriod = end
	}
	_, err := execAsUser(t, app, userA, insertRegimeSQL, userA, taxpayerID, code, personType, start, endPeriod)
	return err
}

func TestTaxpayerTables_AppRoleIsolation(t *testing.T) {
	t.Run("taxpayers", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		app := db.Connect(t, pgtest.RoleApp)
		seedUsers(t, app, userA, userB)

		pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
			Table: taxpayersTable, UserA: userA, UserB: userB, Insert: insertOwnTaxpayer,
		})
	})
	t.Run("taxpayer_regimes", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		app := db.Connect(t, pgtest.RoleApp)
		seedUsers(t, app, userA, userB)
		seedTaxpayer(t, app, userA, taxpayerA, "fisica")
		seedTaxpayer(t, app, userB, taxpayerB, "fisica")

		pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
			Table: taxpayerRegimesTable, UserA: userA, UserB: userB, Insert: insertOwnRegime,
		})
	})
}

// AssertIsolation's generic UPDATE rewrites user_id, which cc_app may not update at all;
// prove the cross-user writes on the columns it may write match no row.
func TestTaxpayerTables_CrossUserWritesMatchNoRow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	seedTaxpayer(t, app, userB, taxpayerB, "fisica")
	_, err := execAsUser(t, app, userB, insertRegimeSQL, userB, taxpayerB, "612", "fisica", "2026-01", nil)
	require.NoError(t, err)
	writes := []string{
		"UPDATE taxpayers SET display_name = 'Mallory' WHERE user_id = $1",
		"UPDATE taxpayers SET archived_at = now() WHERE user_id = $1",
		"UPDATE taxpayers SET rfc_hash = sha256('x') WHERE user_id = $1",
		"UPDATE taxpayer_regimes SET options = '{\"x\": 1}' WHERE user_id = $1",
		"UPDATE taxpayer_regimes SET end_period = '2026-02' WHERE user_id = $1",
		"DELETE FROM taxpayers WHERE user_id = $1",
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
}

// The composite foreign keys make a cross-user reference impossible even when the
// application forges ids: RLS checks user_id, the FK checks that the taxpayer, its owner
// and its person type belong together, and that the regime exists for that person type.
func TestTaxpayerRegimes_CompositeForeignKeysRefuseForgedReferences(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	seedTaxpayer(t, app, userA, taxpayerA, "fisica")
	seedTaxpayer(t, app, userB, taxpayerB, "fisica")
	cases := []struct {
		name                       string
		userID, taxpayerID, regime string
		personType, wantSQLState   string
	}{
		{"own user id, another user's taxpayer", userA, taxpayerB, "612", "fisica", sqlstateForeignKeyViolation},
		{"another user's id and taxpayer", userB, taxpayerB, "612", "fisica", sqlstateInsufficientPrivilege},
		{"own taxpayer, forged person type", userA, taxpayerA, "601", "moral", sqlstateForeignKeyViolation},
		{"own taxpayer, regime of the other person type", userA, taxpayerA, "601", "fisica", sqlstateForeignKeyViolation},
		{"own taxpayer, unknown regime", userA, taxpayerA, "999", "fisica", sqlstateForeignKeyViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insertRegimeSQL, c.userID, c.taxpayerID, c.regime, c.personType, "2026-01", nil)
			requireSQLState(t, err, c.wantSQLState)
		})
	}
	t.Run("the well-formed reference is accepted", func(t *testing.T) {
		require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-01", ""))
	})
}

// person_type is immutable (MODULE-taxpayers §2.1): cc_app has no UPDATE grant on it,
// and a trigger refuses the change for any role that might ever get one.
func TestTaxpayers_IdentityColumnsAreImmutable(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	seedTaxpayer(t, app, userA, taxpayerA, "fisica")

	for _, query := range []string{
		"UPDATE taxpayers SET person_type = 'moral' WHERE id = $1",
		"UPDATE taxpayers SET user_id = '" + userB + "' WHERE id = $1",
		"UPDATE taxpayers SET id = uuidv7() WHERE id = $1",
		"UPDATE taxpayers SET created_at = now() WHERE id = $1",
	} {
		t.Run("cc_app: "+query, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, query, taxpayerA)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}

	t.Run("trigger refuses even a role with the grant", func(t *testing.T) {
		// The owner bypasses RLS once FORCE is lifted on this throwaway database, and
		// owns every privilege: only the trigger stands in the way.
		owner := db.Connect(t, pgtest.RoleOwner)
		_, err := owner.ExecContext(t.Context(), "ALTER TABLE taxpayers NO FORCE ROW LEVEL SECURITY")
		require.NoError(t, err)
		for _, query := range []string{
			"UPDATE taxpayers SET person_type = 'moral', rfc_masked = 'GOM******AB6' WHERE id = $1",
			"UPDATE taxpayers SET user_id = '" + userB + "' WHERE id = $1",
			"UPDATE taxpayers SET id = uuidv7() WHERE id = $1",
		} {
			_, err := owner.ExecContext(t.Context(), query, taxpayerA)
			requireSQLState(t, err, sqlstateRaiseException)
		}
		res, err := owner.ExecContext(t.Context(), "UPDATE taxpayers SET display_name = 'Ana' WHERE id = $1", taxpayerA)
		require.NoError(t, err)
		affected, err := res.RowsAffected()
		require.NoError(t, err)
		assert.EqualValues(t, 1, affected, "the trigger lets mutable columns change")
	})
}

func TestTaxpayers_RejectInvalidRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	const insert = `
INSERT INTO taxpayers (user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked)
VALUES ($1, $2, $3, $4, $5, $6, $7)`
	framed := append([]byte{0x00, 0x01}, make([]byte, 40)...)
	hash := make([]byte, 32)
	valid := []any{userA, "fisica", "Ana", framed, 1, hash, "GOMA******AB6"}
	with := func(i int, v any) []any {
		args := append([]any(nil), valid...)
		args[i] = v
		return args
	}
	cases := []struct {
		name string
		args []any
		want string
	}{
		{"unknown person type", with(1, "otra"), sqlstateCheckViolation},
		{"blank display name", with(2, "  "), sqlstateCheckViolation},
		{"untrimmed display name", with(2, " Ana"), sqlstateCheckViolation},
		{"display name over 80 characters", with(2, strings.Repeat("a", 81)), sqlstateCheckViolation},
		{"display name not NFC", with(2, "José"), sqlstateCheckViolation},
		{"ciphertext shorter than its frame", with(3, []byte{0x00, 0x01, 0x02}), sqlstateCheckViolation},
		{"ciphertext key id differs from rfc_key_id", with(4, 2), sqlstateCheckViolation},
		{"key id below 1", with(4, 0), sqlstateCheckViolation},
		{"hash not 32 bytes", with(5, make([]byte, 31)), sqlstateCheckViolation},
		{"física mask on a moral", with(1, "moral"), sqlstateCheckViolation},
		{"moral mask on a física", with(6, "GOM******AB6"), sqlstateCheckViolation},
		{"mask that reveals the date", with(6, "GOMA800101AB6"), sqlstateCheckViolation},
		{"null display name", with(2, nil), sqlstateNotNullViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, insert, c.args...)
			requireSQLState(t, err, c.want)
		})
	}
	t.Run("valid física and moral rows", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, insert, valid...)
		require.NoError(t, err)
		moral := with(1, "moral")
		moral[5] = blindIndex("moral")
		moral[6] = "GOM******AB6"
		_, err = execAsUser(t, app, userA, insert, moral...)
		require.NoError(t, err)
	})
	t.Run("same RFC hash twice for one user", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, insert, valid...)
		requireSQLState(t, err, sqlstateUniqueViolation)
	})
}

// The same RFC (same blind index) may be registered by two users: uniqueness is per user.
func TestTaxpayers_RFCHashIsUniquePerUserOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	const insert = `
INSERT INTO taxpayers (user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked)
VALUES ($1, 'fisica', 'Ana', '\x0001'::bytea || sha256('n') || sha256('c'), 1, sha256('same rfc'), 'GOMA******AB6')`
	for _, user := range []string{userA, userB} {
		_, err := execAsUser(t, app, user, insert, user)
		require.NoError(t, err, user)
	}
}

func TestTaxpayerRegimes_RejectInvalidRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	seedTaxpayer(t, app, userA, taxpayerA, "fisica")
	cases := []struct {
		name, start, end string
	}{
		{"start before the catalog floor", "2021-12", ""},
		{"month 13", "2026-13", ""},
		{"month 00", "2026-00", ""},
		{"not a period", "26-01-1", ""},
		{"end before start", "2026-05", "2026-04"},
		{"malformed end", "2026-05", "2026-5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireSQLState(t, addRegime(t, app, taxpayerA, "fisica", "612", c.start, c.end), sqlstateCheckViolation)
		})
	}
	t.Run("options must be an object", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period, options)
VALUES ($1, $2, '612', 'fisica', '2026-01', '[]')`, userA, taxpayerA)
		requireSQLState(t, err, sqlstateCheckViolation)
	})
	t.Run("a one-month window is valid", func(t *testing.T) {
		require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-05", "2026-05"))
	})
}

// One row per (taxpayer, regime) — the PK — and two regimes the catalog marks as
// incompatible (regime_compatibility.allowed = false) may not have overlapping periods
// (MODULE-taxpayers §2.5). Sequential windows and allowed pairs are fine.
func TestTaxpayerRegimes_IncompatibleRegimesNeverOverlap(t *testing.T) {
	type regime struct{ code, start, end string }
	cases := []struct {
		name       string
		personType string
		existing   []regime
		add        regime
		want       string // "" = accepted
	}{
		{"same regime twice", "fisica", []regime{{"612", "2026-01", ""}}, regime{"612", "2026-03", ""}, sqlstateUniqueViolation},
		{"626 x 612 overlapping", "fisica", []regime{{"612", "2026-01", ""}}, regime{"626", "2026-03", ""}, sqlstateExclusionViolation},
		{"626 x 612 overlapping by one month", "fisica", []regime{{"626", "2026-01", "2026-06"}}, regime{"612", "2026-06", ""}, sqlstateExclusionViolation},
		{"612 opened before a closed 626", "fisica", []regime{{"626", "2026-03", "2026-06"}}, regime{"612", "2025-01", ""}, sqlstateExclusionViolation},
		{"626 exits to 612 the next month (113-E)", "fisica", []regime{{"626", "2026-01", "2026-06"}}, regime{"612", "2026-07", ""}, ""},
		{"allowed pair 606 x 625", "fisica", []regime{{"606", "2026-01", ""}}, regime{"625", "2026-01", ""}, ""},
		{"moral 601 x 626", "moral", []regime{{"601", "2026-01", ""}}, regime{"626", "2026-01", ""}, sqlstateExclusionViolation},
		{"moral pair allowed with a CPA warning", "moral", []regime{{"601", "2026-01", ""}}, regime{"603", "2026-01", ""}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := pgtest.NewDatabase(t)
			app := db.Connect(t, pgtest.RoleApp)
			seedUsers(t, app, userA)
			seedTaxpayer(t, app, userA, taxpayerA, c.personType)
			for _, r := range c.existing {
				require.NoError(t, addRegime(t, app, taxpayerA, c.personType, r.code, r.start, r.end))
			}

			err := addRegime(t, app, taxpayerA, c.personType, c.add.code, c.add.start, c.add.end)

			if c.want == "" {
				require.NoError(t, err)
				return
			}
			requireSQLState(t, err, c.want)
		})
	}

	t.Run("reopening a closed regime into an overlap is refused", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		app := db.Connect(t, pgtest.RoleApp)
		seedUsers(t, app, userA)
		seedTaxpayer(t, app, userA, taxpayerA, "fisica")
		require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "626", "2026-01", "2026-06"))
		require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-07", ""))

		_, err := execAsUser(t, app, userA,
			"UPDATE taxpayer_regimes SET end_period = NULL WHERE taxpayer_id = $1 AND regime_code = '626'", taxpayerA)

		requireSQLState(t, err, sqlstateExclusionViolation)
	})

	t.Run("another taxpayer's regimes never clash", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		app := db.Connect(t, pgtest.RoleApp)
		seedUsers(t, app, userA)
		const second = "00000000-0000-7000-8000-0000000000a2"
		seedTaxpayer(t, app, userA, taxpayerA, "fisica")
		seedTaxpayer(t, app, userA, second, "fisica")
		require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-01", ""))
		require.NoError(t, addRegime(t, app, second, "fisica", "626", "2026-01", ""))
	})
}

// countForCreate runs the cap-check step of CreateTaxpayer: lock the user, count.
func countForCreate(ctx context.Context, tx *sql.Tx) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx, "SELECT taxpayer_count_for_create()").Scan(&n)
	return n, err
}

// OD-2 / S-5 (MODULE-taxpayers §2.8, US-T2): at 9 taxpayers, 5 concurrent creates
// admit exactly one. Each create runs what the use case runs: lock + count, the cap
// check, a pause that widens the race window, then the insert.
func TestTaxpayerCountForCreate_ConcurrentCreatesAtNineAdmitExactlyOne(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	for i := range taxpayerCap - 1 {
		seedTaxpayer(t, app, userA, taxpayerID(i), "fisica")
	}
	const creates = 5
	var (
		start   sync.WaitGroup
		done    sync.WaitGroup
		results = make([]error, creates)
	)
	start.Add(1)
	for i := range creates {
		done.Add(1)
		go func() {
			defer done.Done()
			start.Wait()
			results[i] = pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
				n, err := countForCreate(t.Context(), tx)
				if err != nil {
					return err
				}
				if n >= taxpayerCap {
					return errCapReached
				}
				if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.05)"); err != nil {
					return err
				}
				_, err = tx.ExecContext(t.Context(), insertTaxpayerSQL, userA, taxpayerID(100+i), "fisica")
				return err
			})
		}()
	}
	start.Done()
	done.Wait()

	var admitted, refused int
	for _, err := range results {
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, errCapReached):
			refused++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}
	assert.Equal(t, 1, admitted)
	assert.Equal(t, creates-1, refused)
	var total int
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM taxpayers").Scan(&total)
	}))
	assert.Equal(t, taxpayerCap, total)
}

// taxpayerID returns a deterministic uuidv7-shaped id numbered i.
func taxpayerID(i int) string {
	return fmt.Sprintf("00000000-0000-7000-8000-%012d", i)
}

func TestTaxpayerCountForCreate_FailsClosed(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	seedTaxpayer(t, app, userB, taxpayerB, "fisica")

	t.Run("counts archived taxpayers too", func(t *testing.T) {
		_, err := execAsUser(t, app, userB, "UPDATE taxpayers SET archived_at = now() WHERE id = $1", taxpayerB)
		require.NoError(t, err)
		var n int
		require.NoError(t, pgtest.AsUser(t.Context(), app, userB, func(tx *sql.Tx) error {
			n, err = countForCreate(t.Context(), tx)
			return err
		}))
		assert.Equal(t, 1, n)
	})
	t.Run("counts only the context user's taxpayers", func(t *testing.T) {
		var n int
		require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
			var err error
			n, err = countForCreate(t.Context(), tx)
			return err
		}))
		assert.Zero(t, n)
	})
	t.Run("no user context", func(t *testing.T) {
		err := pgtest.WithoutUser(t.Context(), app, func(tx *sql.Tx) error {
			_, err := countForCreate(t.Context(), tx)
			return err
		})
		requireSQLState(t, err, sqlstateNoDataFound)
	})
	t.Run("refuses a snapshot isolation level that would miss concurrent creates", func(t *testing.T) {
		tx, err := app.BeginTx(t.Context(), &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		_, err = tx.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", userA)
		require.NoError(t, err)
		_, err = countForCreate(t.Context(), tx)
		requireSQLState(t, err, sqlstateInvalidTransactionState)
	})
	t.Run("not executable by the other runtime roles", func(t *testing.T) {
		for _, role := range []pgtest.Role{pgtest.RoleWebhook, pgtest.RoleWorker} {
			err := pgtest.AsUser(t.Context(), db.Connect(t, role), userA, func(tx *sql.Tx) error {
				_, err := countForCreate(t.Context(), tx)
				return err
			})
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		}
	})
}

const sqlstateInvalidTransactionState = "25000"

// Account deletion (US-A4, S-7) and taxpayer deletion (§2.7) cascade through both tables.
func TestTaxpayerTables_DeletionCascades(t *testing.T) {
	setup := func(t *testing.T) *sql.DB {
		t.Helper()
		app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
		seedUsers(t, app, userA)
		seedTaxpayer(t, app, userA, taxpayerA, "fisica")
		require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-01", ""))
		return app
	}
	countRows := func(t *testing.T, app *sql.DB) (taxpayers, regimes int) {
		t.Helper()
		require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(),
				"SELECT (SELECT count(*) FROM taxpayers), (SELECT count(*) FROM taxpayer_regimes)").Scan(&taxpayers, &regimes)
		}))
		return taxpayers, regimes
	}
	t.Run("account deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		taxpayers, regimes := countRows(t, app)
		assert.Zero(t, taxpayers)
		assert.Zero(t, regimes)
	})
	t.Run("taxpayer deletion", func(t *testing.T) {
		app := setup(t)
		affected, err := execAsUser(t, app, userA, "DELETE FROM taxpayers WHERE id = $1", taxpayerA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		_, regimes := countRows(t, app)
		assert.Zero(t, regimes)
	})
}

// Least privilege (data-model.md "Database roles"): only cc_app reaches these tables,
// and only on the columns its use cases write.
func TestTaxpayerTables_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	seedTaxpayer(t, app, userA, taxpayerA, "fisica")
	require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-01", ""))
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE taxpayers SET updated_at = now()"},
		{pgtest.RoleApp, "UPDATE taxpayer_regimes SET taxpayer_id = taxpayer_id"},
		{pgtest.RoleApp, "UPDATE taxpayer_regimes SET user_id = user_id"},
		{pgtest.RoleApp, "UPDATE taxpayer_regimes SET regime_code = '606'"},
		{pgtest.RoleApp, "UPDATE taxpayer_regimes SET person_type = 'moral'"},
		{pgtest.RoleApp, "UPDATE taxpayer_regimes SET start_period = '2026-02'"},
		{pgtest.RoleApp, "DELETE FROM taxpayer_regimes"},
		{pgtest.RoleApp, "TRUNCATE taxpayers CASCADE"},
		{pgtest.RoleWebhook, "SELECT 1 FROM taxpayers"},
		{pgtest.RoleWebhook, "SELECT 1 FROM taxpayer_regimes"},
		{pgtest.RoleWorker, "SELECT 1 FROM taxpayers"},
		{pgtest.RoleWorker, "SELECT 1 FROM taxpayer_regimes"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
	t.Run("cc_app writes what the use cases need", func(t *testing.T) {
		for _, query := range []string{
			"UPDATE taxpayers SET display_name = 'Ana (negocio)'",
			"UPDATE taxpayers SET rfc_encrypted = rfc_encrypted, rfc_key_id = rfc_key_id, rfc_hash = rfc_hash, rfc_masked = rfc_masked",
			"UPDATE taxpayers SET archived_at = now()",
			"UPDATE taxpayers SET archived_at = NULL",
			"UPDATE taxpayer_regimes SET options = '{\"iva_fully_taxable\": true}', needs_cpa_review_ack = true",
			"UPDATE taxpayer_regimes SET end_period = '2026-03'",
			"UPDATE taxpayer_regimes SET end_period = NULL",
		} {
			affected, err := execAsUser(t, app, userA, query)
			require.NoError(t, err, query)
			assert.EqualValues(t, 1, affected, query)
		}
	})
}

// updated_at is now() of the writing transaction on both tables (set_updated_at).
func TestTaxpayerTables_UpdatedAtFollowsWrites(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	seedTaxpayer(t, app, userA, taxpayerA, "fisica")
	require.NoError(t, addRegime(t, app, taxpayerA, "fisica", "612", "2026-01", ""))
	writes := map[string]string{
		"taxpayers":        "UPDATE taxpayers SET display_name = 'Ana (negocio)'",
		"taxpayer_regimes": "UPDATE taxpayer_regimes SET options = '{}'",
	}
	for table, write := range writes {
		t.Run(table, func(t *testing.T) {
			var followed bool
			require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(t.Context(), "SELECT pg_sleep(0.01)"); err != nil {
					return err
				}
				if _, err := tx.ExecContext(t.Context(), write); err != nil {
					return err
				}
				return tx.QueryRowContext(t.Context(),
					"SELECT bool_and(updated_at = now() AND updated_at > created_at) FROM "+table).Scan(&followed)
			}))
			assert.True(t, followed)
		})
	}
}

func TestTaxpayersMigration_DownRemovesEverything(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), taxpayersVersion-1)
	require.NoError(t, err)
	for _, obj := range []string{
		"to_regclass('public.taxpayers')", "to_regclass('public.taxpayer_regimes')",
		"to_regprocedure('public.taxpayer_count_for_create()')",
		"to_regprocedure('public.taxpayers_forbid_identity_change()')",
		"to_regprocedure('public.taxpayer_regimes_check_compatible()')",
	} {
		var found sql.NullString
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT "+obj+"::text").Scan(&found))
		assert.False(t, found.Valid, "down left %s", found.String)
	}

	require.NoError(t, migrations.Up(t.Context(), owner))
	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var exists bool
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass('public.taxpayer_regimes') IS NOT NULL").Scan(&exists))
	assert.True(t, exists)
}
