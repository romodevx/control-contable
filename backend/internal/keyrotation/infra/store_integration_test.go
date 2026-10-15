//go:build integration

package infra_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/keyrotation/infra"
	"github.com/romodevx/control-contable/backend/internal/keyrotation/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

const (
	oldKeyID    crypto.KeyID = 1
	activeKeyID crypto.KeyID = 2

	lockWaitDeadline = 5 * time.Second
	pollInterval     = 10 * time.Millisecond
)

// Fixture ids of this package only.
var (
	users     = []string{"00000000-0000-4000-8000-00000000c001", "00000000-0000-4000-8000-00000000c002"}
	taxpayers = map[string][]string{ // per user: one live, one archived
		users[0]: {"00000000-0000-7000-8000-00000000c011", "00000000-0000-7000-8000-00000000c012"},
		users[1]: {"00000000-0000-7000-8000-00000000c021", "00000000-0000-7000-8000-00000000c022"},
	}
	devices = map[string]string{
		users[0]: "0199d4a4-3c4d-7f5e-8a61-00000000c0d1",
		users[1]: "0199d4a4-3c4d-7f5e-8a61-00000000c0d2",
	}
)

func randomKey(t *testing.T, id crypto.KeyID) crypto.KeyVersion {
	t.Helper()
	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return crypto.KeyVersion{ID: id, Key: key}
}

func keyring(t *testing.T, active crypto.KeyVersion, retired ...crypto.KeyVersion) *crypto.Keyring {
	t.Helper()
	kr, err := crypto.NewRotatingKeyring(active, retired...)
	require.NoError(t, err)
	return kr
}

// aad binds a test value to its target and row; the stores never see it.
func aad(target string, ref usecase.Ref) []byte {
	return []byte(target + ":" + ref.ID + ":" + ref.UserID)
}

type sealer struct {
	target  string
	keyring *crypto.Keyring
}

func (s sealer) ActiveKeyID() crypto.KeyID { return s.keyring.ActiveKeyID() }

func (s sealer) Reseal(ciphertext []byte, ref usecase.Ref) ([]byte, error) {
	return s.keyring.Reseal(ciphertext, aad(s.target, ref))
}

// world is one database seeded through cc_app, every value sealed with old.
type world struct {
	db       *pgtest.Database
	app      *sql.DB
	worker   *sql.DB
	old, new crypto.KeyVersion
}

func seed(t *testing.T) *world {
	t.Helper()
	db := pgtest.NewDatabase(t)
	w := &world{db: db, app: db.Connect(t, pgtest.RoleApp), worker: db.Connect(t, pgtest.RoleWorker),
		old: randomKey(t, oldKeyID), new: randomKey(t, activeKeyID)}
	old := keyring(t, w.old)
	seal := func(target string, ref usecase.Ref) []byte {
		ct, err := old.Encrypt([]byte("secret-of-"+ref.ID), aad(target, ref))
		require.NoError(t, err)
		return ct
	}
	for _, user := range users {
		require.NoError(t, pgtest.AsUser(t.Context(), w.app, user, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(t.Context(), `
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash)
VALUES ($1::text::uuid, sha256(convert_to($1::text, 'UTF8')), $2, 1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo')`,
				user, seal(infra.TargetUserEmail, usecase.Ref{ID: user, UserID: user})); err != nil {
				return err
			}
			for i, id := range taxpayers[user] {
				if _, err := tx.ExecContext(t.Context(), `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($1::text::uuid, $2, 'fisica', 'Ana', $3, 1, sha256(convert_to($1::text, 'UTF8')), 'GOMA******AB6', true)`,
					id, user, seal(infra.TargetTaxpayerRFC, usecase.Ref{ID: id, UserID: user})); err != nil {
					return err
				}
				if i == 1 { // archived: its RFC is re-sealed too
					if _, err := tx.ExecContext(t.Context(), "UPDATE taxpayers SET archived_at = now() WHERE id = $1", id); err != nil {
						return err
					}
				}
			}
			_, err := tx.ExecContext(t.Context(), `
SELECT claim_push_token($1::uuid, $2::text::uuid, 'android', $3, 1::smallint, sha256(convert_to($2::text, 'UTF8')), '1.0.0')`,
				user, devices[user], seal(infra.TargetDeviceToken, usecase.Ref{ID: devices[user], UserID: user}))
			return err
		}))
	}
	return w
}

// columns are the three real stores, each with a sealer over kr.
func columns(kr *crypto.Keyring, wrap func(usecase.Store) usecase.Store) []usecase.Column {
	if wrap == nil {
		wrap = func(s usecase.Store) usecase.Store { return s }
	}
	return []usecase.Column{
		{Target: infra.TargetUserEmail, Store: wrap(infra.NewUserEmailStore()), Sealer: sealer{infra.TargetUserEmail, kr}},
		{Target: infra.TargetTaxpayerRFC, Store: wrap(infra.NewTaxpayerRFCStore()), Sealer: sealer{infra.TargetTaxpayerRFC, kr}},
		{Target: infra.TargetDeviceToken, Store: wrap(infra.NewDeviceTokenStore()), Sealer: sealer{infra.TargetDeviceToken, kr}},
	}
}

func (w *world) job(t *testing.T, cols []usecase.Column, batch int) *usecase.ResealJob {
	t.Helper()
	units, err := pgdb.NewTxManager(w.worker)
	require.NoError(t, err)
	job, err := usecase.NewResealJob(usecase.ResealDeps{
		Tx: units, System: units, Usage: infra.NewUsageReader(), Columns: cols,
		Logger: slog.New(slog.DiscardHandler), BatchSize: batch, MaxBatches: usecase.DefaultMaxBatches,
	})
	require.NoError(t, err)
	return job
}

// sealedRow is one stored value as the owner of the row sees it (cc_app).
type sealedRow struct {
	target string
	ref    usecase.Ref
	value  []byte
}

func (w *world) rows(t *testing.T) []sealedRow {
	t.Helper()
	var out []sealedRow
	for _, user := range users {
		require.NoError(t, pgtest.AsUser(t.Context(), w.app, user, func(tx *sql.Tx) error {
			var email []byte
			if err := tx.QueryRowContext(t.Context(), "SELECT email_encrypted FROM users WHERE id = $1", user).Scan(&email); err != nil {
				return err
			}
			out = append(out, sealedRow{infra.TargetUserEmail, usecase.Ref{ID: user, UserID: user}, email})
			for _, id := range taxpayers[user] {
				var rfc []byte
				if err := tx.QueryRowContext(t.Context(), "SELECT rfc_encrypted FROM taxpayers WHERE id = $1", id).Scan(&rfc); err != nil {
					return err
				}
				out = append(out, sealedRow{infra.TargetTaxpayerRFC, usecase.Ref{ID: id, UserID: user}, rfc})
			}
			return nil
		}))
		// cc_app cannot read token_ciphertext; cc_worker can, under the owner's context.
		require.NoError(t, pgtest.AsUser(t.Context(), w.worker, user, func(tx *sql.Tx) error {
			var token []byte
			err := tx.QueryRowContext(t.Context(), "SELECT token_ciphertext FROM devices WHERE id = $1", devices[user]).Scan(&token)
			out = append(out, sealedRow{infra.TargetDeviceToken, usecase.Ref{ID: devices[user], UserID: user}, token})
			return err
		}))
	}
	return out
}

// requireAllOpenWith checks every value is on kr's active key and opens with kr alone.
func (w *world) requireAllOpenWith(t *testing.T, kr *crypto.Keyring) {
	t.Helper()
	rows := w.rows(t)
	require.Len(t, rows, 8)
	for _, r := range rows {
		id, err := crypto.KeyIDOf(r.value)
		require.NoError(t, err)
		assert.Equal(t, kr.ActiveKeyID(), id, r.target+" "+r.ref.ID)
		plain, err := kr.Decrypt(r.value, aad(r.target, r.ref))
		require.NoError(t, err, r.target+" "+r.ref.ID)
		assert.Equal(t, "secret-of-"+r.ref.ID, string(plain))
	}
}

func (w *world) usage(t *testing.T) map[string]int64 {
	t.Helper()
	units, err := pgdb.NewTxManager(w.worker)
	require.NoError(t, err)
	out := map[string]int64{}
	require.NoError(t, units.System(t.Context(), func(ctx context.Context) error {
		rows, err := infra.NewUsageReader().KeyUsage(ctx)
		for _, u := range rows {
			out[fmt.Sprintf("%s/v%d", u.Target, u.Version)] = u.Rows
		}
		return err
	}))
	return out
}

func TestResealJob_MovesEveryColumnToTheActiveKey(t *testing.T) {
	w := seed(t)
	rotating := keyring(t, w.new, w.old)

	require.NoError(t, w.job(t, columns(rotating, nil), 1).Run(t.Context()))

	w.requireAllOpenWith(t, keyring(t, w.new))
	assert.Equal(t, map[string]int64{
		"users.email_encrypted/v2": 2, "users.email_hash/v1": 2,
		"taxpayers.rfc_encrypted/v2": 4, "taxpayers.rfc_hash/v1": 4,
		"devices.token_ciphertext/v2": 2, "devices.token_hmac/v1": 2,
	}, w.usage(t))
	before := w.rows(t)
	require.NoError(t, w.job(t, columns(rotating, nil), 1).Run(t.Context()))
	assert.Equal(t, before, w.rows(t), "a second run rewrites nothing")
}

// Rolling back is the same job with the keys swapped: rehearse it on a copy
// (here: this test's own database) before the forward run in production.
func TestResealJob_RollsBack(t *testing.T) {
	w := seed(t)
	require.NoError(t, w.job(t, columns(keyring(t, w.new, w.old), nil), usecase.DefaultBatchSize).Run(t.Context()))

	require.NoError(t, w.job(t, columns(keyring(t, w.old, w.new), nil), usecase.DefaultBatchSize).Run(t.Context()))

	w.requireAllOpenWith(t, keyring(t, w.old))
}

// countingStore counts Replace calls and runs hook after each.
type countingStore struct {
	usecase.Store
	replaced *atomic.Int64
	hook     func()
}

func (s countingStore) Replace(ctx context.Context, ref usecase.Ref, ct []byte, keyID crypto.KeyID) error {
	if err := s.Store.Replace(ctx, ref, ct, keyID); err != nil {
		return err
	}
	s.replaced.Add(1)
	if s.hook != nil {
		s.hook()
	}
	return nil
}

func TestResealJob_ResumesAfterAnInterruption(t *testing.T) {
	w := seed(t)
	rotating := keyring(t, w.new, w.old)
	var replaced atomic.Int64
	ctx, cancel := context.WithCancel(t.Context())
	interrupted := columns(rotating, func(s usecase.Store) usecase.Store {
		return countingStore{Store: s, replaced: &replaced, hook: func() {
			if replaced.Load() == 3 {
				cancel()
			}
		}}
	})

	err := w.job(t, interrupted, 1).Run(ctx)

	require.ErrorIs(t, err, context.Canceled)
	pending := w.usage(t)
	assert.Positive(t, pending["taxpayers.rfc_encrypted/v1"]+pending["devices.token_ciphertext/v1"], "the run stopped part way")

	var resumed atomic.Int64
	require.NoError(t, w.job(t, columns(rotating, func(s usecase.Store) usecase.Store {
		return countingStore{Store: s, replaced: &resumed}
	}), 1).Run(t.Context()))
	w.requireAllOpenWith(t, keyring(t, w.new))
	assert.LessOrEqual(t, replaced.Load()+resumed.Load(), int64(8+1), "at most the interrupted row's commit is redone")
}

// Two workers that both believe they lead: each row ends re-sealed and opens.
func TestResealJob_TwoWorkersAtOnce(t *testing.T) {
	w := seed(t)
	rotating := keyring(t, w.new, w.old)
	var replaced atomic.Int64
	count := func(s usecase.Store) usecase.Store { return countingStore{Store: s, replaced: &replaced} }
	jobs := []*usecase.ResealJob{w.job(t, columns(rotating, count), 1), w.job(t, columns(rotating, count), 2)}

	var wg sync.WaitGroup
	errs := make([]error, len(jobs))
	for i, job := range jobs {
		wg.Go(func() { errs[i] = job.Run(t.Context()) })
	}
	wg.Wait()

	require.NoError(t, errors.Join(errs...))
	w.requireAllOpenWith(t, keyring(t, w.new))
	assert.EqualValues(t, 8, replaced.Load(), "no row re-sealed twice")
}

// The row lock is what makes split brain safe: worker 1 has re-sealed a row
// and not yet committed; worker 2 must wait for it, then find the row on
// the active key and leave worker 1's value alone. Without FOR UPDATE worker
// 2 would read the old value and overwrite worker 1's.
func TestResealJob_WaitsForAConcurrentReSealOfTheSameRow(t *testing.T) {
	w := seed(t)
	rotating := keyring(t, w.new, w.old)
	user, id := users[0], taxpayers[users[0]][0]
	blocker, err := w.worker.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = blocker.Rollback() })
	_, err = blocker.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", user)
	require.NoError(t, err)
	var old []byte
	require.NoError(t, blocker.QueryRowContext(t.Context(), "SELECT rfc_encrypted FROM taxpayers WHERE id = $1 FOR UPDATE", id).Scan(&old))
	first, err := rotating.Reseal(old, aad(infra.TargetTaxpayerRFC, usecase.Ref{ID: id, UserID: user}))
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "UPDATE taxpayers SET rfc_encrypted = $1, rfc_key_id = $2 WHERE id = $3", first, activeKeyID, id)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- w.job(t, columns(rotating, nil)[1:2], usecase.DefaultBatchSize).Run(context.WithoutCancel(t.Context()))
	}()
	awaitLockWaiter(t, w.worker) // same role: pg_stat_activity hides other roles' waits
	require.NoError(t, blocker.Commit())
	require.NoError(t, <-done)

	for _, r := range w.rows(t) {
		if r.ref.ID == id {
			assert.Equal(t, first, r.value, "worker 1's value is kept")
		}
	}
}

// awaitLockWaiter polls until a backend of db's role waits on a lock.
func awaitLockWaiter(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(lockWaitDeadline)
	for time.Now().Before(deadline) {
		var waiting int
		require.NoError(t, pgtest.WithoutUser(t.Context(), db, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(), `
SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'`).Scan(&waiting)
		}))
		if waiting > 0 {
			return
		}
		time.Sleep(pollInterval)
	}
	t.Fatal("the second worker never waited on the row lock")
}

// A forged ref (another user's row under one user's context) finds nothing:
// RLS scopes the store like the API.
func TestStores_SeeOnlyTheContextUsersRows(t *testing.T) {
	w := seed(t)
	units, err := pgdb.NewTxManager(w.worker)
	require.NoError(t, err)
	other := users[1]
	for _, c := range []struct {
		store usecase.Store
		ref   usecase.Ref
	}{
		{infra.NewUserEmailStore(), usecase.Ref{ID: other, UserID: other}},
		{infra.NewTaxpayerRFCStore(), usecase.Ref{ID: taxpayers[other][0], UserID: other}},
		{infra.NewDeviceTokenStore(), usecase.Ref{ID: devices[other], UserID: other}},
	} {
		require.NoError(t, units.ReadWrite(t.Context(), users[0], func(ctx context.Context) error {
			_, found, err := c.store.Lock(ctx, c.ref)
			assert.False(t, found)
			if err != nil {
				return err
			}
			err = c.store.Replace(ctx, c.ref, []byte{0, 2}, activeKeyID)
			assert.ErrorIs(t, err, infra.ErrRowNotUpdated)
			return nil
		}))
	}
}

func TestStores_RefuseToRunOutsideAUnitOfWork(t *testing.T) {
	_, err := infra.NewTaxpayerRFCStore().Pending(t.Context(), activeKeyID, "", 1)
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
	_, err = infra.NewUsageReader().KeyUsage(t.Context())
	require.ErrorIs(t, err, pgdb.ErrNoTransaction)
}
