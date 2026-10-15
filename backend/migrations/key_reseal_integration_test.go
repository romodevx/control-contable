//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Key re-seal migration (ADR-0012): the worker's cross-user list of the rows still
// sealed with a retired key, the per-key row counts (readiness), cc_worker's per-user
// right to rewrite a sealed column, and the hash version columns of the blind indexes.

const (

	// The reseal fixtures' own ids, so no other test's rows are selected by value.
	resealTaxpayerA2 = "00000000-0000-7000-8000-0000000000a2"
	resealTaxpayerB2 = "00000000-0000-7000-8000-0000000000b2"

	resealActiveKey = 2
	resealMaxLimit  = 1000

	// insertTaxpayerOnKeySQL is insertTaxpayerSQL with the key id as $4: int2send gives
	// the 2-byte big-endian prefix the framing CHECK compares with rfc_key_id.
	insertTaxpayerOnKeySQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, 'fisica', 'Ana (personal)',
        int2send($3::smallint) || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        $3, sha256(convert_to($2::text || '/rfc', 'UTF8')), 'GOMA******AB6', true)`

	// reframeSQL is a ciphertext of $1 bytes sealed with key id $2 (framing only).
	reframeSQL = "int2send($2::smallint) || sha256(convert_to($1::text, 'UTF8')) || sha256(convert_to($1::text, 'UTF8'))"
)

// resealFixture: users A and B (email key 1), A with two taxpayers (key 1, one archived,
// and key 2) and B with one on key 1 and one on key 2; one device each on key 1.
type resealFixture struct {
	owner, app, worker *sql.DB
}

func seedResealFixture(t *testing.T) resealFixture {
	t.Helper()
	db := pgtest.NewDatabase(t)
	f := resealFixture{
		owner: db.Connect(t, pgtest.RoleOwner), app: db.Connect(t, pgtest.RoleApp), worker: db.Connect(t, pgtest.RoleWorker),
	}
	seedUsers(t, f.app, userA, userB)
	for _, row := range []struct {
		user, id string
		key      int
	}{{userA, taxpayerA, 1}, {userA, resealTaxpayerA2, resealActiveKey}, {userB, taxpayerB, 1}, {userB, resealTaxpayerB2, resealActiveKey}} {
		_, err := execAsUser(t, f.app, row.user, insertTaxpayerOnKeySQL, row.user, row.id, row.key)
		require.NoError(t, err)
	}
	// An archived taxpayer still holds its RFC ciphertext: it is re-sealed too.
	_, err := execAsUser(t, f.app, userA, "UPDATE taxpayers SET archived_at = now() WHERE id = $1", taxpayerA)
	require.NoError(t, err)
	for _, user := range []string{userA, userB} {
		require.True(t, claim(t, f.app, user, installationOf[user], "token-of-"+user))
	}
	return f
}

type pendingRow struct{ ID, UserID string }

func pendingRows(t *testing.T, worker *sql.DB, fn string, active int, after any, limit int) []pendingRow {
	t.Helper()
	rows, err := worker.QueryContext(t.Context(),
		"SELECT id::text, user_id::text FROM "+fn+"($1::smallint, $2::uuid, $3)", active, after, limit)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var out []pendingRow
	for rows.Next() {
		var r pendingRow
		require.NoError(t, rows.Scan(&r.ID, &r.UserID))
		out = append(out, r)
	}
	require.NoError(t, rows.Err())
	return out
}

func TestKeyReseal_PendingFunctionsListRowsOnOtherKeysAcrossUsers(t *testing.T) {
	f := seedResealFixture(t)

	assert.Equal(t, []pendingRow{{taxpayerA, userA}, {taxpayerB, userB}},
		pendingRows(t, f.worker, "reseal_pending_taxpayers", resealActiveKey, nil, resealMaxLimit),
		"rows on any key but the active one, archived included, ordered by id")
	assert.Equal(t, []pendingRow{{taxpayerA, userA}},
		pendingRows(t, f.worker, "reseal_pending_taxpayers", resealActiveKey, nil, 1), "bounded by the limit")
	assert.Equal(t, []pendingRow{{taxpayerB, userB}},
		pendingRows(t, f.worker, "reseal_pending_taxpayers", resealActiveKey, taxpayerA, resealMaxLimit), "keyset: after the cursor")
	assert.Equal(t, []pendingRow{{resealTaxpayerA2, userA}, {resealTaxpayerB2, userB}},
		pendingRows(t, f.worker, "reseal_pending_taxpayers", 1, nil, resealMaxLimit), "rollback lists the other key")

	assert.Equal(t, []pendingRow{{userA, userA}, {userB, userB}},
		pendingRows(t, f.worker, "reseal_pending_users", resealActiveKey, nil, resealMaxLimit))
	assert.Empty(t, pendingRows(t, f.worker, "reseal_pending_users", 1, nil, resealMaxLimit))

	assert.Equal(t, []pendingRow{{installationOf[userA], userA}, {installationOf[userB], userB}},
		pendingRows(t, f.worker, "reseal_pending_devices", resealActiveKey, nil, resealMaxLimit))
	assert.Empty(t, pendingRows(t, f.worker, "reseal_pending_devices", 1, nil, resealMaxLimit))
}

func TestKeyReseal_PendingFunctionsRefuseBadArguments(t *testing.T) {
	f := seedResealFixture(t)
	for _, fn := range []string{"reseal_pending_taxpayers", "reseal_pending_users", "reseal_pending_devices"} {
		for _, args := range [][]any{{resealActiveKey, 0}, {resealActiveKey, resealMaxLimit + 1}, {0, 1}, {nil, 1}, {resealActiveKey, nil}} {
			_, err := f.worker.ExecContext(t.Context(), "SELECT * FROM "+fn+"($1::smallint, NULL, $2::integer)", args...)
			requireSQLState(t, err, sqlstateRaiseException)
		}
	}
}

func TestKeyReseal_KeyUsageCountsRowsPerVersion(t *testing.T) {
	f := seedResealFixture(t)
	type usage struct {
		Target  string
		Version int
		Rows    int
	}
	rows, err := f.worker.QueryContext(t.Context(), "SELECT target, version, row_count FROM key_usage()")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var got []usage
	for rows.Next() {
		var u usage
		require.NoError(t, rows.Scan(&u.Target, &u.Version, &u.Rows))
		got = append(got, u)
	}
	require.NoError(t, rows.Err())

	assert.Equal(t, []usage{
		{"devices.token_ciphertext", 1, 2}, {"devices.token_hmac", 1, 2},
		{"taxpayers.rfc_encrypted", 1, 2}, {"taxpayers.rfc_encrypted", 2, 2}, {"taxpayers.rfc_hash", 1, 4},
		{"users.email_encrypted", 1, 2}, {"users.email_hash", 1, 2},
	}, got)
}

func TestKeyReseal_DefinerFunctionsPosture(t *testing.T) {
	f := seedResealFixture(t)
	for _, fn := range []string{
		"reseal_pending_taxpayers(smallint, uuid, integer)", "reseal_pending_users(smallint, uuid, integer)",
		"reseal_pending_devices(smallint, uuid, integer)", "key_usage()",
	} {
		definer, fnOwner, config := definerFunction(t, f.owner, fn)
		assert.True(t, definer, fn)
		assert.Equal(t, workerDefinerRole, fnOwner, fn)
		assert.Equal(t, "search_path=pg_catalog, public, pg_temp", config, fn)
		assert.Equal(t, []string{"cc_worker"}, executors(t, f.owner, fn, fnOwner), fn)
	}
}

// cc_worker rewrites a sealed column only under the row owner's context, and only the
// ciphertext and its key id: never a hash, a name or another user's row.
func TestKeyReseal_WorkerRewritesOnlyOwnSealedColumns(t *testing.T) {
	f := seedResealFixture(t)
	cases := []struct {
		name, ownUpdate, otherUpdate string
		ownArgs, otherArgs           []any
		forbidden                    []string
	}{
		{
			name:      "taxpayers",
			ownUpdate: "UPDATE taxpayers SET rfc_encrypted = " + reframeSQL + ", rfc_key_id = $2 WHERE id = $3",
			ownArgs:   []any{"new-a", resealActiveKey, taxpayerA}, otherArgs: []any{"new-b", resealActiveKey, taxpayerB},
			forbidden: []string{
				"UPDATE taxpayers SET rfc_hash = sha256('x')", "UPDATE taxpayers SET display_name = 'x'",
				"UPDATE taxpayers SET rfc_hash_version = 2",
			},
		},
		{
			name:      "users",
			ownUpdate: "UPDATE users SET email_encrypted = " + reframeSQL + ", email_key_id = $2 WHERE id = $3",
			ownArgs:   []any{"new-a", resealActiveKey, userA}, otherArgs: []any{"new-b", resealActiveKey, userB},
			forbidden: []string{
				"UPDATE users SET email_hash = sha256('x')", "UPDATE users SET name = 'x'",
				"UPDATE users SET email_hash_version = 2",
			},
		},
		{
			name:      "devices",
			ownUpdate: "UPDATE devices SET token_ciphertext = " + reframeSQL + ", token_key_id = $2 WHERE id = $3",
			ownArgs:   []any{"new-a", resealActiveKey, installationOf[userA]}, otherArgs: []any{"new-b", resealActiveKey, installationOf[userB]},
			forbidden: []string{"UPDATE devices SET token_hmac = sha256('x')", "UPDATE devices SET token_hmac_version = 2"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n, err := execAsUser(t, f.worker, userA, c.ownUpdate, c.ownArgs...)
			require.NoError(t, err)
			assert.EqualValues(t, 1, n, "own row")

			n, err = execAsUser(t, f.worker, userA, c.ownUpdate, c.otherArgs...)
			require.NoError(t, err)
			assert.Zero(t, n, "another user's row is invisible")

			for _, stmt := range c.forbidden {
				_, err := execAsUser(t, f.worker, userA, stmt)
				requireSQLState(t, err, sqlstateInsufficientPrivilege)
			}
		})
	}
}

// The re-seal reads the row FOR UPDATE under the owner's context; without a context
// cc_worker still sees no users or taxpayers row.
func TestKeyReseal_WorkerLocksOwnRowsAndSeesNothingWithoutContext(t *testing.T) {
	f := seedResealFixture(t)
	require.NoError(t, pgtest.AsUser(t.Context(), f.worker, userA, func(tx *sql.Tx) error {
		var keyID int
		if err := tx.QueryRowContext(t.Context(),
			"SELECT email_key_id FROM users WHERE id = $1 FOR UPDATE", userA).Scan(&keyID); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(),
			"SELECT rfc_key_id FROM taxpayers WHERE id = $1 FOR UPDATE", taxpayerA).Scan(&keyID)
	}))
	for _, table := range []string{"users", "taxpayers"} {
		require.NoError(t, pgtest.WithoutUser(t.Context(), f.worker, func(tx *sql.Tx) error {
			var n int
			require.NoError(t, tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table).Scan(&n))
			assert.Zero(t, n, table)
			return nil
		}))
	}
}

// The definer sees archived taxpayers only inside a re-seal function (the scan flag the
// functions set and clear): due_reminder_users keeps its archived_at IS NULL policy.
func TestKeyReseal_DefinerSeesArchivedTaxpayersOnlyDuringAScan(t *testing.T) {
	f := seedResealFixture(t)
	count := func(scan bool) int {
		var n int
		require.NoError(t, pgtest.WithoutUser(t.Context(), f.owner, func(tx *sql.Tx) error {
			if scan {
				if _, err := tx.ExecContext(t.Context(), "SELECT set_config('app.key_reseal_scan', 'on', true)"); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(t.Context(), "SET LOCAL ROLE "+workerDefinerRole); err != nil {
				return err
			}
			return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM taxpayers").Scan(&n)
		}))
		return n
	}

	assert.Equal(t, 3, count(false), "live taxpayers only")
	assert.Equal(t, 4, count(true), "every taxpayer during a scan")

	var flag string
	require.NoError(t, pgtest.WithoutUser(t.Context(), f.worker, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), "SELECT * FROM key_usage()"); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(), "SELECT coalesce(current_setting('app.key_reseal_scan', true), '')").Scan(&flag)
	}))
	assert.Empty(t, flag, "the function clears the flag before it returns")
}

func TestKeyReseal_HashVersionColumns(t *testing.T) {
	f := seedResealFixture(t)
	for _, c := range []struct{ constraint, def string }{
		{"users_email_hash_version_range", "CHECK ((email_hash_version >= 1))"},
		{"taxpayers_rfc_hash_version_range", "CHECK ((rfc_hash_version >= 1))"},
		{"devices_token_hmac_version_range", "CHECK ((token_hmac_version >= 1))"},
	} {
		var def string
		require.NoError(t, f.owner.QueryRowContext(t.Context(),
			"SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = $1", c.constraint).Scan(&def))
		assert.Equal(t, c.def, def)
	}
	var version int
	require.NoError(t, pgtest.AsUser(t.Context(), f.app, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT email_hash_version FROM users WHERE id = $1", userA).Scan(&version)
	}))
	assert.Equal(t, 1, version, "every existing and new hash is version 1")
}
