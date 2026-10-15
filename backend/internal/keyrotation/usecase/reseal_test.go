package usecase_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/keyrotation/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
)

const (
	target  = "fake.ciphertext"
	userOne = "00000000-0000-4000-8000-000000000001"
	userTwo = "00000000-0000-4000-8000-000000000002"
)

func rowID(i int) string { return fmt.Sprintf("00000000-0000-7000-8000-%012d", i) }

// harness is one job over one fake column.
type harness struct {
	keys  keys
	store *fakeStore
	units *fakeUnits
	logs  *syncBuffer
	usage fakeUsage
}

// newHarness seeds n rows on the old key, alternating between two users.
func newHarness(t *testing.T, n int) *harness {
	t.Helper()
	h := &harness{keys: newKeys(t), store: newFakeStore(), units: &fakeUnits{}, logs: &syncBuffer{}}
	old := h.keys.keyring(t, h.keys.old)
	for i := range n {
		user := userOne
		if i%2 == 1 {
			user = userTwo
		}
		h.store.seal(t, old, rowID(i), user)
	}
	h.usage = fakeUsage{stores: map[string]*fakeStore{target: h.store}}
	return h
}

func (h *harness) job(t *testing.T, batch, maxBatches int) *usecase.ResealJob {
	t.Helper()
	return h.jobWith(t, h.keys.keyring(t, h.keys.active, h.keys.old), batch, maxBatches)
}

func (h *harness) jobWith(t *testing.T, kr *crypto.Keyring, batch, maxBatches int) *usecase.ResealJob {
	t.Helper()
	job, err := usecase.NewResealJob(usecase.ResealDeps{
		Tx: h.units, System: h.units, Usage: h.usage,
		Columns:   []usecase.Column{{Target: target, Store: h.store, Sealer: sealer{kr}}},
		Logger:    slog.New(slog.NewJSONHandler(h.logs, nil)),
		BatchSize: batch, MaxBatches: maxBatches,
	})
	require.NoError(t, err)
	return job
}

// finished decodes the last "key reseal finished" line.
func (h *harness) finished(t *testing.T) map[string]any {
	t.Helper()
	var last map[string]any
	for line := range strings.Lines(h.logs.String()) {
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		if entry["msg"] == "key reseal finished" {
			last = entry
		}
	}
	require.NotNil(t, last, "no finished line in %s", h.logs)
	return last
}

func columnLog(t *testing.T, entry map[string]any) map[string]any {
	t.Helper()
	columns, ok := entry["columns"].(map[string]any)
	require.True(t, ok, "columns group: %v", entry)
	col, ok := columns[target].(map[string]any)
	require.True(t, ok, "column group: %v", columns)
	return col
}

func TestNewResealJob_validates(t *testing.T) {
	h := newHarness(t, 0)
	kr := h.keys.keyring(t, h.keys.active)
	valid := func() usecase.ResealDeps {
		return usecase.ResealDeps{
			Tx: h.units, System: h.units, Usage: h.usage, Logger: slog.New(slog.DiscardHandler),
			Columns:   []usecase.Column{{Target: target, Store: h.store, Sealer: sealer{kr}}},
			BatchSize: usecase.DefaultBatchSize, MaxBatches: usecase.DefaultMaxBatches,
		}
	}
	tests := []struct {
		name   string
		mutate func(*usecase.ResealDeps)
	}{
		{"nil tx", func(d *usecase.ResealDeps) { d.Tx = nil }},
		{"nil system", func(d *usecase.ResealDeps) { d.System = nil }},
		{"nil usage", func(d *usecase.ResealDeps) { d.Usage = nil }},
		{"nil logger", func(d *usecase.ResealDeps) { d.Logger = nil }},
		{"column without store", func(d *usecase.ResealDeps) { d.Columns[0].Store = nil }},
		{"column without sealer", func(d *usecase.ResealDeps) { d.Columns[0].Sealer = nil }},
		{"column without target", func(d *usecase.ResealDeps) { d.Columns[0].Target = "" }},
		{"duplicate target", func(d *usecase.ResealDeps) { d.Columns = append(d.Columns, d.Columns[0]) }},
		{"zero batch", func(d *usecase.ResealDeps) { d.BatchSize = 0 }},
		{"batch above the database bound", func(d *usecase.ResealDeps) { d.BatchSize = usecase.MaxBatchSize + 1 }},
		{"zero max batches", func(d *usecase.ResealDeps) { d.MaxBatches = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := valid()
			tt.mutate(&deps)

			job, err := usecase.NewResealJob(deps)

			require.ErrorIs(t, err, usecase.ErrInvalidConfig)
			assert.Nil(t, job)
		})
	}
	t.Run("no column is valid: the job only reports readiness", func(t *testing.T) {
		deps := valid()
		deps.Columns = nil
		_, err := usecase.NewResealJob(deps)
		require.NoError(t, err)
	})
}

func TestResealJob_identity(t *testing.T) {
	job := newHarness(t, 0).job(t, 1, 1)

	assert.Equal(t, "key_reseal", job.Name())
	assert.Equal(t, 15*time.Minute, job.Interval())
	assert.Less(t, usecase.RunTimeout, job.Interval())
}

func TestResealJob_movesEveryRowToTheActiveKeyInItsOwnersUnitOfWork(t *testing.T) {
	h := newHarness(t, 5)

	require.NoError(t, h.job(t, 2, usecase.DefaultMaxBatches).Run(t.Context()))

	for id, keyID := range h.store.keyIDs(t) {
		assert.Equal(t, activeKeyID, keyID, id)
	}
	assert.Equal(t, []string{userOne, userTwo, userOne, userTwo, userOne}, h.units.users, "one unit of work per row, as its owner")
	assert.Equal(t, []string{userOne, userTwo, userOne, userTwo, userOne}, h.store.lockedAs)
	assert.Equal(t, []string{"", rowID(1), rowID(3)}, h.store.pendingCalls, "keyset pages of 2")
	// The active key alone opens every row: the old one can be dropped.
	onlyActive := h.keys.keyring(t, h.keys.active)
	for id, r := range h.store.rows {
		plain, err := onlyActive.Decrypt(r.ciphertext, rowAAD(usecase.Ref{ID: id, UserID: r.userID}))
		require.NoError(t, err)
		assert.Equal(t, "plaintext-of-"+id, string(plain))
	}
	col := columnLog(t, h.finished(t))
	assert.EqualValues(t, 5, col["resealed"])
	assert.EqualValues(t, 0, col["rows_on_old_keys"])
	assert.Equal(t, false, col["more"])
}

func TestResealJob_secondRunIsANoOp(t *testing.T) {
	h := newHarness(t, 3)
	job := h.job(t, usecase.DefaultBatchSize, usecase.DefaultMaxBatches)
	require.NoError(t, job.Run(t.Context()))
	before := h.store.keyIDs(t)
	h.store.replaced = nil

	require.NoError(t, job.Run(t.Context()))

	assert.Empty(t, h.store.replaced)
	assert.Equal(t, before, h.store.keyIDs(t))
	assert.EqualValues(t, 0, columnLog(t, h.finished(t))["resealed"])
}

// A run cut by MaxBatches leaves the rest for the next run, which picks up
// where the rows' key ids say: no checkpoint to lose.
func TestResealJob_resumesAcrossRuns(t *testing.T) {
	h := newHarness(t, 5)
	job := h.job(t, 2, 1)

	require.NoError(t, job.Run(t.Context()))
	first := columnLog(t, h.finished(t))
	assert.EqualValues(t, 2, first["resealed"])
	assert.Equal(t, true, first["more"])
	assert.EqualValues(t, 3, first["rows_on_old_keys"])

	require.NoError(t, job.Run(t.Context()))
	require.NoError(t, job.Run(t.Context()))

	last := columnLog(t, h.finished(t))
	assert.EqualValues(t, 0, last["rows_on_old_keys"])
	assert.Equal(t, false, last["more"])
	assert.Len(t, h.store.replaced, 5, "every row re-sealed exactly once")
}

// An interrupted run (shutdown, lost leadership) stops between rows; the
// next run finishes the job.
func TestResealJob_interruptionResumes(t *testing.T) {
	h := newHarness(t, 6)
	ctx, cancel := context.WithCancel(t.Context())
	locks := 0
	h.store.onLock = func(string) {
		locks++
		if locks == 3 {
			cancel()
		}
	}

	err := h.job(t, 2, usecase.DefaultMaxBatches).Run(ctx)

	require.ErrorIs(t, err, context.Canceled)
	assert.NotEmpty(t, h.finished(t), "the counts so far are logged")
	h.store.onLock = nil
	require.NoError(t, h.job(t, 2, usecase.DefaultMaxBatches).Run(t.Context()))
	for id, keyID := range h.store.keyIDs(t) {
		assert.Equal(t, activeKeyID, keyID, id)
	}
	assert.Len(t, h.store.replaced, 6, "no row re-sealed twice")
}

// A row the keyring cannot open (a key it lacks, a ciphertext moved from
// another row) is counted and left alone; the run goes on and ends.
func TestResealJob_countsUnreadableRowsAndGoesOn(t *testing.T) {
	h := newHarness(t, 2)
	h.store.seal(t, h.keys.keyring(t, h.keys.foreign), rowID(7), userOne)
	moved := h.store.rows[rowID(0)]
	h.store.rows[rowID(8)] = fakeRow{userID: moved.userID, ciphertext: moved.ciphertext}

	require.NoError(t, h.job(t, 1, usecase.DefaultMaxBatches).Run(t.Context()))

	ids := h.store.keyIDs(t)
	assert.Equal(t, activeKeyID, ids[rowID(0)])
	assert.Equal(t, activeKeyID, ids[rowID(1)])
	assert.Equal(t, foreignKey, ids[rowID(7)])
	assert.Equal(t, oldKeyID, ids[rowID(8)])
	col := columnLog(t, h.finished(t))
	assert.EqualValues(t, 2, col["resealed"])
	assert.EqualValues(t, 2, col["unreadable"])
	assert.EqualValues(t, 2, col["rows_on_old_keys"])
	assert.Contains(t, h.logs.String(), `"level":"WARN"`, "unreadable rows need an operator")
}

// Another worker (split brain) or the API re-sealed the row between the
// list and the lock: it is skipped, never rewritten.
func TestResealJob_skipsRowsAlreadyOnTheActiveKeyOrGone(t *testing.T) {
	h := newHarness(t, 2)
	active := h.keys.keyring(t, h.keys.active)
	h.store.onLock = func(id string) {
		if id == rowID(0) {
			h.store.seal(t, active, rowID(0), userOne)
		} else {
			delete(h.store.rows, id)
		}
	}

	require.NoError(t, h.job(t, usecase.DefaultBatchSize, usecase.DefaultMaxBatches).Run(t.Context()))

	assert.Empty(t, h.store.replaced)
	col := columnLog(t, h.finished(t))
	assert.EqualValues(t, 0, col["resealed"])
	assert.EqualValues(t, 2, col["skipped"])
}

// Two workers that both believe they lead run the job at once: each row
// ends re-sealed once and opens; the two runs' counts add up.
func TestResealJob_twoConcurrentRunsReSealEachRowOnce(t *testing.T) {
	h := newHarness(t, 40)
	jobs := []*usecase.ResealJob{h.job(t, 3, usecase.DefaultMaxBatches), h.job(t, 5, usecase.DefaultMaxBatches)}

	var wg sync.WaitGroup
	errs := make([]error, len(jobs))
	for i, job := range jobs {
		wg.Go(func() { errs[i] = job.Run(t.Context()) })
	}
	wg.Wait()

	require.NoError(t, errors.Join(errs...))
	assert.Len(t, h.store.replaced, 40)
	for id, keyID := range h.store.keyIDs(t) {
		assert.Equal(t, activeKeyID, keyID, id)
	}
}

// Rolling back is the same job with the keys' roles swapped.
func TestResealJob_rollsBack(t *testing.T) {
	h := newHarness(t, 3)
	require.NoError(t, h.job(t, usecase.DefaultBatchSize, 1).Run(t.Context()))

	back := h.keys.keyring(t, h.keys.old, h.keys.active)
	require.NoError(t, h.jobWith(t, back, usecase.DefaultBatchSize, 1).Run(t.Context()))

	onlyOld := h.keys.keyring(t, h.keys.old)
	for id, r := range h.store.rows {
		_, err := onlyOld.Decrypt(r.ciphertext, rowAAD(usecase.Ref{ID: id, UserID: r.userID}))
		require.NoError(t, err, id)
	}
}

func TestResealJob_storeErrorsStopTheRun(t *testing.T) {
	boom := errors.New("connection reset")
	t.Run("listing", func(t *testing.T) {
		h := newHarness(t, 2)
		h.store.pendingErr = boom
		err := h.job(t, 1, 1).Run(t.Context())
		require.ErrorIs(t, err, boom)
		assert.NotContains(t, err.Error(), rowID(0))
	})
	t.Run("rewrite", func(t *testing.T) {
		h := newHarness(t, 2)
		h.store.replaceErr = boom
		err := h.job(t, 1, 1).Run(t.Context())
		require.ErrorIs(t, err, boom)
	})
	t.Run("usage", func(t *testing.T) {
		h := newHarness(t, 1)
		h.usage.err = boom
		err := h.job(t, 1, 1).Run(t.Context())
		require.ErrorIs(t, err, boom)
	})
}

// The readiness line reports every target key_usage returns, columns the
// job does not hold included (a worker without the RFC key still reports
// how many RFCs are on each key), and nothing but counts.
func TestResealJob_logsCountsOnly(t *testing.T) {
	h := newHarness(t, 2)
	h.usage.extra = []usecase.Usage{{Target: "taxpayers.rfc_encrypted", Version: 1, Rows: 4}, {Target: "taxpayers.rfc_hash", Version: 1, Rows: 4}}

	require.NoError(t, h.job(t, 1, usecase.DefaultMaxBatches).Run(t.Context()))

	entry := h.finished(t)
	usage, ok := entry["key_usage"].(map[string]any)
	require.True(t, ok, "%v", entry)
	assert.Equal(t, map[string]any{"v1": float64(4)}, usage["taxpayers.rfc_encrypted"])
	assert.Equal(t, map[string]any{"v2": float64(2)}, usage[target])
	logs := h.logs.String()
	for id, r := range h.store.rows {
		assert.NotContains(t, logs, id)
		assert.NotContains(t, logs, r.userID)
		assert.NotContains(t, logs, hex.EncodeToString(r.ciphertext))
	}
	assert.NotContains(t, logs, "plaintext-of")
}
