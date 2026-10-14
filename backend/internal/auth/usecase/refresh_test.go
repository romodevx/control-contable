package usecase_test

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// seqReader is a deterministic entropy source whose every read differs, so
// each refresh secret it draws is new.
type seqReader struct{ next byte }

func (s *seqReader) Read(p []byte) (int, error) {
	s.next++
	for i := range p {
		p[i] = s.next
	}
	return len(p), nil
}

// storedToken is one row of the fake refresh_tokens table.
type storedToken struct {
	id      string
	family  string
	userID  domain.UserID
	rotated bool
	revoked bool
	reason  usecase.RevokeReason
	expired bool
	// absoluteLeft is the time left before the family's absolute expiry.
	absoluteLeft time.Duration
}

// fakeRotation is an in-memory refresh_tokens table with the semantics of
// infra.RefreshTokenStore.
type fakeRotation struct {
	mu      sync.Mutex
	rows    map[domain.TokenHash]*storedToken
	lockErr error
	// tamper makes Lock return a hash other than the one asked for, to
	// prove the use case compares it.
	tamper  bool
	nextID  int
	rotates int
	revokes []usecase.RevokeReason
}

func newFakeRotation() *fakeRotation {
	return &fakeRotation{rows: map[domain.TokenHash]*storedToken{}}
}

// add stores token as the first of a new family.
func (f *fakeRotation) add(token domain.RefreshToken, row storedToken) *storedToken {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	row.id = strconv.Itoa(f.nextID)
	if row.family == "" {
		row.family = "family-" + row.id
	}
	row.userID = token.UserID()
	if row.absoluteLeft == 0 {
		row.absoluteLeft = usecase.RefreshAbsoluteLifetime
	}
	f.rows[token.Hash()] = &row
	return &row
}

func (f *fakeRotation) row(token domain.RefreshToken) *storedToken {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rows[token.Hash()]
}

func (f *fakeRotation) Lock(_ context.Context, hash domain.TokenHash) (usecase.StoredRefreshToken, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lockErr != nil {
		return usecase.StoredRefreshToken{}, false, f.lockErr
	}
	row, ok := f.rows[hash]
	if !ok {
		return usecase.StoredRefreshToken{}, false, nil
	}
	stored := hash
	if f.tamper {
		stored[0] ^= 0xff
	}
	return usecase.StoredRefreshToken{
		ID: row.id, Family: row.family, Hash: stored, Rotated: row.rotated, Revoked: row.revoked, Expired: row.expired,
	}, true, nil
}

func (f *fakeRotation) Rotate(_ context.Context, parent usecase.StoredRefreshToken, successor domain.RefreshToken, sliding time.Duration) (time.Duration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rotates++
	for _, row := range f.rows {
		if row.id != parent.ID {
			continue
		}
		if row.rotated || row.revoked {
			return 0, errors.New("fake: parent is not claimable")
		}
		row.rotated = true
		f.nextID++
		f.rows[successor.Hash()] = &storedToken{
			id: strconv.Itoa(f.nextID), family: row.family, userID: successor.UserID(), absoluteLeft: row.absoluteLeft,
		}
		return min(sliding, row.absoluteLeft), nil
	}
	return 0, errors.New("fake: unknown parent")
}

func (f *fakeRotation) RevokeFamily(_ context.Context, hash domain.TokenHash, reason usecase.RevokeReason) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.revokes = append(f.revokes, reason)
	target, ok := f.rows[hash]
	if !ok {
		return nil
	}
	for _, row := range f.rows {
		if row.family == target.family && !row.revoked {
			row.revoked, row.reason = true, reason
		}
	}
	return nil
}

// fakeStatus answers CanSignIn per user.
type fakeStatus struct {
	active map[domain.UserID]bool
	err    error
	asked  []domain.UserID
}

func (f *fakeStatus) CanSignIn(_ context.Context, id domain.UserID) (bool, error) {
	f.asked = append(f.asked, id)
	return f.active[id], f.err
}

type refreshEnv struct {
	uc     *usecase.RefreshSession
	tx     *fakeTx
	store  *fakeRotation
	status *fakeStatus
	userID domain.UserID
	logs   interface{ String() string }
	seed   byte
	// rotations limits the rotations of each family (MODULE-auth §5).
	rotations *fakeLimiter
}

// nextSeed makes each issued fixture token's secret different.
func (e *refreshEnv) nextSeed() byte {
	e.seed++
	return e.seed
}

func newRefreshEnv(t *testing.T) *refreshEnv {
	t.Helper()
	userID, err := fixedUserID()
	require.NoError(t, err)
	logger, logs := capturedLogs()
	e := &refreshEnv{
		tx:        &fakeTx{},
		store:     newFakeRotation(),
		status:    &fakeStatus{active: map[domain.UserID]bool{userID: true}},
		userID:    userID,
		logs:      logs,
		rotations: &fakeLimiter{limit: 100},
	}
	e.uc = usecase.NewRefreshSession(usecase.RefreshDeps{
		Tx: e.tx, Clock: fixedClock, Entropy: &seqReader{next: 100}, Tokens: e.store,
		Accounts: e.status, Access: fakeAccess{}, Rotations: e.rotations, Logger: logger,
	})
	return e
}

// issued stores a new live token of the env's user and returns it.
func (e *refreshEnv) issued(t *testing.T, row storedToken) domain.RefreshToken {
	t.Helper()
	token, err := domain.NewRefreshToken(e.userID, &seqReader{next: e.nextSeed()})
	require.NoError(t, err)
	e.store.add(token, row)
	return token
}

func TestRefresh_liveTokenRotatesInsideItsFamily(t *testing.T) {
	e := newRefreshEnv(t)
	parent := e.issued(t, storedToken{})

	tokens, err := e.uc.Execute(t.Context(), parent.Reveal())

	require.NoError(t, err)
	assert.Equal(t, newUserID, e.tx.userID, "the unit of work runs as the token's user (RLS)")
	require.NoError(t, e.tx.workErr, "committed")
	assert.True(t, e.store.row(parent).rotated, "the old token is marked rotated")
	successor := e.store.row(tokens.Refresh)
	require.NotNil(t, successor, "the new token is stored")
	assert.Equal(t, e.store.row(parent).family, successor.family, "same family")
	assert.False(t, successor.rotated || successor.revoked)
	assert.NotEqual(t, parent.Reveal(), tokens.Refresh.Reveal())
	assert.Equal(t, e.userID, tokens.Refresh.UserID())
	assert.Equal(t, "access:"+newUserID+"@2026-10-13T18:30:00Z", tokens.Access)
	assert.Equal(t, accessTTL, tokens.AccessExpiresIn)
	assert.Equal(t, usecase.RefreshSlidingWindow, tokens.RefreshExpiresIn)
	assert.Equal(t, []domain.UserID{e.userID}, e.status.asked, "the account status is re-checked")
}

// The successor inherits the family's absolute expiry: near the 90-day
// limit the client is told the real, shorter lifetime.
func TestRefresh_lifetimeIsCappedByTheFamilysAbsoluteExpiry(t *testing.T) {
	e := newRefreshEnv(t)
	parent := e.issued(t, storedToken{absoluteLeft: 2 * time.Hour})

	tokens, err := e.uc.Execute(t.Context(), parent.Reveal())

	require.NoError(t, err)
	assert.Equal(t, 2*time.Hour, tokens.RefreshExpiresIn)
}

// US-A2b: every refusal is ErrRefreshRejected (one generic 401). Only reuse
// of a rotated token revokes the family; an expired, revoked or unknown
// token, or an account that may not sign in, changes nothing.
func TestRefresh_refusals(t *testing.T) {
	tests := []struct {
		name        string
		row         *storedToken // nil: the token is not stored
		inactive    bool
		wantErr     error
		wantRevokes []usecase.RevokeReason
	}{
		{name: "unknown token", wantErr: usecase.ErrTokenUnknown},
		{name: "expired token", row: &storedToken{expired: true}, wantErr: usecase.ErrTokenExpired},
		{name: "revoked token", row: &storedToken{revoked: true, reason: usecase.RevokeLogout}, wantErr: usecase.ErrTokenRevoked},
		{
			name: "rotated token (reuse)", row: &storedToken{rotated: true},
			wantErr: usecase.ErrTokenReused, wantRevokes: []usecase.RevokeReason{usecase.RevokeReuseDetected},
		},
		{
			name: "rotated and expired token is still reuse", row: &storedToken{rotated: true, expired: true},
			wantErr: usecase.ErrTokenReused, wantRevokes: []usecase.RevokeReason{usecase.RevokeReuseDetected},
		},
		{name: "suspended or deleted account", row: &storedToken{}, inactive: true, wantErr: usecase.ErrAccountInactive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newRefreshEnv(t)
			token, err := domain.NewRefreshToken(e.userID, &seqReader{next: 7})
			require.NoError(t, err)
			if tt.row != nil {
				e.store.add(token, *tt.row)
			}
			if tt.inactive {
				e.status.active[e.userID] = false
			}

			_, err = e.uc.Execute(t.Context(), token.Reveal())

			require.ErrorIs(t, err, usecase.ErrRefreshRejected)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantRevokes, e.store.revokes)
			assert.Zero(t, e.store.rotates, "nothing is rotated")
			require.NoError(t, e.tx.workErr, "a refusal commits, so a reuse revocation is kept")
			assert.NotContains(t, e.logs.String(), token.Reveal())
		})
	}
}

// Only the token's own row is consulted before the claim: an account that
// may not sign in is checked after it (security gate M1-02, condition 1).
func TestRefresh_statusIsNotReadForATokenThatIsNotLive(t *testing.T) {
	e := newRefreshEnv(t)
	token := e.issued(t, storedToken{expired: true})

	_, err := e.uc.Execute(t.Context(), token.Reveal())

	require.ErrorIs(t, err, usecase.ErrRefreshRejected)
	assert.Empty(t, e.status.asked)
}

func TestRefresh_malformedTokenNeverReachesTheDatabase(t *testing.T) {
	for _, text := range []string{"", "garbage", newUserID + ".short", "00000000-0000-0000-0000-000000000000.x"} {
		e := newRefreshEnv(t)

		_, err := e.uc.Execute(t.Context(), text)

		require.ErrorIs(t, err, usecase.ErrRefreshRejected, text)
		assert.Zero(t, e.tx.calls, text)
	}
}

// The stored hash is compared with the presented one in constant time; a
// row whose hash does not match is an unknown token, not a session.
func TestRefresh_storedHashMustMatch(t *testing.T) {
	e := newRefreshEnv(t)
	token := e.issued(t, storedToken{})
	e.store.tamper = true

	_, err := e.uc.Execute(t.Context(), token.Reveal())

	require.ErrorIs(t, err, usecase.ErrTokenUnknown)
	assert.Zero(t, e.store.rotates)
}

// Reuse of a rotated token kills the whole family, including the successor
// the legitimate client (or the thief) already holds.
func TestRefresh_reuseRevokesTheSuccessorToo(t *testing.T) {
	e := newRefreshEnv(t)
	parent := e.issued(t, storedToken{})
	tokens, err := e.uc.Execute(t.Context(), parent.Reveal())
	require.NoError(t, err)

	_, err = e.uc.Execute(t.Context(), parent.Reveal())
	require.ErrorIs(t, err, usecase.ErrTokenReused)

	successor := e.store.row(tokens.Refresh)
	assert.True(t, successor.revoked)
	assert.Equal(t, usecase.RevokeReuseDetected, successor.reason)
	_, err = e.uc.Execute(t.Context(), tokens.Refresh.Reveal())
	require.ErrorIs(t, err, usecase.ErrTokenRevoked)
	assert.Contains(t, e.logs.String(), `"event":"token_reuse"`)
	assert.Contains(t, e.logs.String(), `"user_id":"`+newUserID+`"`)
}

func TestRefresh_infrastructureFailuresAreNotRefusals(t *testing.T) {
	tests := []struct {
		name  string
		setup func(e *refreshEnv)
	}{
		{"lock fails", func(e *refreshEnv) { e.store.lockErr = errDB }},
		{"status read fails", func(e *refreshEnv) { e.status.err = errDB }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newRefreshEnv(t)
			token := e.issued(t, storedToken{})
			tt.setup(e)

			_, err := e.uc.Execute(t.Context(), token.Reveal())

			require.ErrorIs(t, err, errDB)
			assert.NotErrorIs(t, err, usecase.ErrRefreshRejected, "a 500, not a 401")
			assert.Zero(t, e.store.rotates)
		})
	}
}
