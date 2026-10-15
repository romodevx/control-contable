package usecase_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/keyrotation/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	oldKeyID    crypto.KeyID = 1
	activeKeyID crypto.KeyID = 2
	foreignKey  crypto.KeyID = 9
)

// randomKey is a fresh AES-256 key: no key material is ever committed.
func randomKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

// keys is one test's key material: the retired key, the active key and a key
// the job never holds.
type keys struct {
	old, active, foreign crypto.KeyVersion
}

func newKeys(t *testing.T) keys {
	t.Helper()
	return keys{
		old:     crypto.KeyVersion{ID: oldKeyID, Key: randomKey(t)},
		active:  crypto.KeyVersion{ID: activeKeyID, Key: randomKey(t)},
		foreign: crypto.KeyVersion{ID: foreignKey, Key: randomKey(t)},
	}
}

func (k keys) keyring(t *testing.T, active crypto.KeyVersion, retired ...crypto.KeyVersion) *crypto.Keyring {
	t.Helper()
	kr, err := crypto.NewRotatingKeyring(active, retired...)
	require.NoError(t, err)
	return kr
}

// rowAAD binds a fake row's ciphertext to its id, as every real column does.
func rowAAD(ref usecase.Ref) []byte { return []byte("fake:" + ref.ID + ":" + ref.UserID) }

// sealer is the fake column's Sealer over a real keyring.
type sealer struct{ keyring *crypto.Keyring }

func (s sealer) ActiveKeyID() crypto.KeyID { return s.keyring.ActiveKeyID() }

func (s sealer) Reseal(ciphertext []byte, ref usecase.Ref) ([]byte, error) {
	return s.keyring.Reseal(ciphertext, rowAAD(ref))
}

// fakeRow is one stored ciphertext.
type fakeRow struct {
	userID     string
	ciphertext []byte
}

// fakeStore is an in-memory sealed column. Every method must run inside a
// unit of work of the right kind; the fake checks it.
type fakeStore struct {
	mu       sync.Mutex
	rows     map[string]fakeRow
	rowLocks map[string]*sync.Mutex

	pendingCalls []string // the after cursor of each call
	replaced     []string
	lockedAs     []string // the user of the unit of work each Lock ran in

	pendingErr error
	replaceErr error
	// onLock runs before Lock reads the row (a concurrent writer, a deletion).
	onLock func(id string)
}

func newFakeStore() *fakeStore {
	return &fakeStore{rows: map[string]fakeRow{}, rowLocks: map[string]*sync.Mutex{}}
}

func (s *fakeStore) seal(t *testing.T, kr *crypto.Keyring, id, userID string) {
	t.Helper()
	ct, err := kr.Encrypt([]byte("plaintext-of-"+id), rowAAD(usecase.Ref{ID: id, UserID: userID}))
	require.NoError(t, err)
	s.rows[id] = fakeRow{userID: userID, ciphertext: ct}
}

func (s *fakeStore) Pending(ctx context.Context, active crypto.KeyID, after string, limit int) ([]usecase.Ref, error) {
	if unitOf(ctx) != systemUnit {
		return nil, errors.New("fake store: Pending outside a System unit of work")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingCalls = append(s.pendingCalls, after)
	if s.pendingErr != nil {
		return nil, s.pendingErr
	}
	ids := make([]string, 0, len(s.rows))
	for id, r := range s.rows {
		keyID, err := crypto.KeyIDOf(r.ciphertext)
		if err != nil || keyID != active {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	var out []usecase.Ref
	for _, id := range ids {
		if id > after && len(out) < limit {
			out = append(out, usecase.Ref{ID: id, UserID: s.rows[id].userID})
		}
	}
	return out, nil
}

// Lock emulates SELECT ... FOR UPDATE: the row stays locked until the unit
// of work in ctx ends.
func (s *fakeStore) Lock(ctx context.Context, ref usecase.Ref) ([]byte, bool, error) {
	if s.onLock != nil {
		s.onLock(ref.ID)
	}
	s.mu.Lock()
	rowLock, ok := s.rowLocks[ref.ID]
	if !ok {
		rowLock = &sync.Mutex{}
		s.rowLocks[ref.ID] = rowLock
	}
	s.mu.Unlock()
	rowLock.Lock()
	unitStateOf(ctx).onEnd(rowLock.Unlock)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lockedAs = append(s.lockedAs, unitOf(ctx))
	r, ok := s.rows[ref.ID]
	if !ok || r.userID != unitOf(ctx) {
		return nil, false, nil // RLS: another user's row is invisible
	}
	return r.ciphertext, true, nil
}

func (s *fakeStore) Replace(ctx context.Context, ref usecase.Ref, ciphertext []byte, keyID crypto.KeyID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.replaceErr != nil {
		return s.replaceErr
	}
	r := s.rows[ref.ID]
	if r.userID != unitOf(ctx) {
		return fmt.Errorf("fake store: Replace of %s outside its owner's unit of work", ref.ID)
	}
	if prefix, err := crypto.KeyIDOf(ciphertext); err != nil || prefix != keyID {
		return errors.New("fake store: key id does not match the ciphertext (framing CHECK)")
	}
	s.rows[ref.ID] = fakeRow{userID: r.userID, ciphertext: ciphertext}
	s.replaced = append(s.replaced, ref.ID)
	return nil
}

func (s *fakeStore) keyIDs(t *testing.T) map[string]crypto.KeyID {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]crypto.KeyID{}
	for id, r := range s.rows {
		keyID, err := crypto.KeyIDOf(r.ciphertext)
		require.NoError(t, err)
		out[id] = keyID
	}
	return out
}

// unitKey marks the unit of work a ctx runs in.
type unitKey struct{}

const systemUnit = "<system>"

// unitState is one unit of work: its user and what runs when it ends (the
// row locks it took).
type unitState struct {
	user  string
	ended []func()
}

func (u *unitState) onEnd(f func()) { u.ended = append(u.ended, f) }

func (u *unitState) end() {
	for _, f := range u.ended {
		f()
	}
}

func unitStateOf(ctx context.Context) *unitState {
	u, _ := ctx.Value(unitKey{}).(*unitState)
	if u == nil {
		return &unitState{}
	}
	return u
}

func unitOf(ctx context.Context) string { return unitStateOf(ctx).user }

func runUnit(ctx context.Context, user string, work tx.Work) error {
	u := &unitState{user: user}
	defer u.end()
	return work(context.WithValue(ctx, unitKey{}, u))
}

// fakeUnits runs work inline and marks ctx with its unit of work; a failed
// work is a rollback, which the fake store cannot undo, so the job must not
// write before it fails.
type fakeUnits struct {
	mu    sync.Mutex
	users []string
}

var (
	_ tx.Manager       = (*fakeUnits)(nil)
	_ tx.SystemManager = (*fakeUnits)(nil)
)

func (u *fakeUnits) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	u.mu.Lock()
	u.users = append(u.users, userID)
	u.mu.Unlock()
	return runUnit(ctx, userID, work)
}

func (u *fakeUnits) ReadOnly(context.Context, string, tx.Work) error {
	return errors.New("fake units: the job never reads only")
}

func (u *fakeUnits) System(ctx context.Context, work tx.Work) error {
	return runUnit(ctx, systemUnit, work)
}

// syncBuffer is a log sink two concurrent runs may share.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// fakeUsage answers key_usage from the stores' current rows, plus extra
// targets the job holds no column for.
type fakeUsage struct {
	stores map[string]*fakeStore
	extra  []usecase.Usage
	err    error
}

func (f fakeUsage) KeyUsage(ctx context.Context) ([]usecase.Usage, error) {
	if unitOf(ctx) != systemUnit {
		return nil, errors.New("fake usage: outside a System unit of work")
	}
	if f.err != nil {
		return nil, f.err
	}
	var out []usecase.Usage
	for target, s := range f.stores {
		counts := map[int]int64{}
		s.mu.Lock()
		for _, r := range s.rows {
			keyID, _ := crypto.KeyIDOf(r.ciphertext)
			counts[int(keyID)]++
		}
		s.mu.Unlock()
		for version, n := range counts {
			out = append(out, usecase.Usage{Target: target, Version: version, Rows: n})
		}
	}
	return append(out, f.extra...), nil
}
