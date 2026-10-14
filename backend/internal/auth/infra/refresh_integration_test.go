//go:build integration

package infra_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// sessionStack is refreshSession and logout over the stack's database as
// cc_app, wired as cmd/api wires them.
type sessionStack struct {
	*stack
	refresh *usecase.RefreshSession
	logout  *usecase.Logout
	store   infra.RefreshTokenStore
	txm     *pgdb.TxManager
}

func newSessionStack(t *testing.T) *sessionStack {
	t.Helper()
	s := &sessionStack{stack: newStack(t), store: infra.NewRefreshTokenStore()}
	var err error
	s.txm, err = pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	access, err := infra.NewAccessTokenSigner(tokenSeed, "k1")
	require.NoError(t, err)
	s.refresh = usecase.NewRefreshSession(usecase.RefreshDeps{
		Tx: s.txm, Clock: clock.NewSystem(), Entropy: rand.Reader, Tokens: s.store,
		Accounts: infra.NewAccountStore(s.keyring), Access: access, Logger: slog.New(slog.DiscardHandler),
	})
	s.logout = usecase.NewLogout(usecase.LogoutDeps{Tx: s.txm, Tokens: s.store})
	return s
}

// session signs a new account up and returns its first session.
func (s *sessionStack) session(t *testing.T, email string) usecase.Session {
	t.Helper()
	require.NoError(t, s.register.Execute(t.Context(), registration(email)))
	session, err := s.verify.Execute(t.Context(), verifying(email, s.inbox.lastCode(t, email), firstSample, nil))
	require.NoError(t, err)
	return session
}

// tokenRow is a refresh_tokens row as the tests inspect it.
type tokenRow struct {
	id, family, rotatedFrom, revokeReason string
	rotated, revoked                      bool
	expiresAt, absoluteExpiresAt          time.Time
}

const rowByHashSQL = `SELECT id::text, family_id::text, coalesce(rotated_from::text, ''), coalesce(revoke_reason, ''),
  rotated_at IS NOT NULL, revoked_at IS NOT NULL, expires_at, absolute_expires_at
FROM refresh_tokens WHERE token_hash = $1`

// row reads token's row as its user (RLS); ok = false when there is none.
func (s *sessionStack) row(t *testing.T, token domain.RefreshToken) (tokenRow, bool) {
	t.Helper()
	var (
		r     tokenRow
		found bool
	)
	s.asUser(t, token.UserID().String(), func(tx *sql.Tx) error {
		err := tx.QueryRowContext(t.Context(), rowByHashSQL, token.Hash().Bytes()).Scan(
			&r.id, &r.family, &r.rotatedFrom, &r.revokeReason, &r.rotated, &r.revoked, &r.expiresAt, &r.absoluteExpiresAt)
		if err == sql.ErrNoRows {
			return nil
		}
		found = err == nil
		return err
	})
	return r, found
}

func (s *sessionStack) mustRow(t *testing.T, token domain.RefreshToken) tokenRow {
	t.Helper()
	r, ok := s.row(t, token)
	require.True(t, ok, "token row exists")
	return r
}

// familyRows reads every row of a family as its user.
func (s *sessionStack) familyRows(t *testing.T, userID, family string) []tokenRow {
	t.Helper()
	var rows []tokenRow
	s.asUser(t, userID, func(tx *sql.Tx) error {
		res, err := tx.QueryContext(t.Context(), strings.Replace(rowByHashSQL, "token_hash = $1", "family_id = $1::uuid", 1), family)
		if err != nil {
			return err
		}
		defer res.Close()
		for res.Next() {
			var r tokenRow
			if err := res.Scan(&r.id, &r.family, &r.rotatedFrom, &r.revokeReason, &r.rotated, &r.revoked, &r.expiresAt, &r.absoluteExpiresAt); err != nil {
				return err
			}
			rows = append(rows, r)
		}
		return res.Err()
	})
	return rows
}

// ownerUpdateTokens runs a fixture UPDATE on refresh_tokens as cc_owner.
// The table is FORCE RLS with no policy for cc_owner, so the fixture lifts
// FORCE for its one statement, in one transaction, and checks it is back.
func (s *sessionStack) ownerUpdateTokens(t *testing.T, set string, token domain.RefreshToken) {
	t.Helper()
	s.ownerForceLifted(t, "refresh_tokens", "UPDATE refresh_tokens SET "+set+" WHERE token_hash = $1", token.Hash().Bytes())
}

func (s *sessionStack) suspend(t *testing.T, userID domain.UserID) {
	t.Helper()
	s.ownerForceLifted(t, "users", "UPDATE users SET status = 'suspended' WHERE id = $1", userID.String())
}

func (s *sessionStack) ownerForceLifted(t *testing.T, table, statement string, arg any) {
	t.Helper()
	ctx := t.Context()
	tx, err := s.owner.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	res, err := tx.ExecContext(ctx, statement, arg)
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, affected, "the fixture changed its row")
	_, err = tx.ExecContext(ctx, "ALTER TABLE "+table+" FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	var forced bool
	require.NoError(t, s.owner.QueryRowContext(ctx,
		"SELECT relforcerowsecurity FROM pg_class WHERE relname = $1", table).Scan(&forced))
	require.True(t, forced, "FORCE is back on")
}

func TestRefresh_rotatesInsideTheFamilyWithItsAbsoluteExpiry(t *testing.T) {
	s := newSessionStack(t)
	parent := s.session(t, anaEmail).Tokens.Refresh

	tokens, err := s.refresh.Execute(t.Context(), parent.Reveal())

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.Access)
	assert.Equal(t, usecase.RefreshSlidingWindow, tokens.RefreshExpiresIn)
	assert.Equal(t, parent.UserID(), tokens.Refresh.UserID())
	old, child := s.mustRow(t, parent), s.mustRow(t, tokens.Refresh)
	assert.True(t, old.rotated, "the old token is marked rotated")
	assert.False(t, old.revoked)
	assert.Equal(t, old.family, child.family, "same family")
	assert.Equal(t, old.id, child.rotatedFrom, "linked to its parent")
	assert.False(t, child.rotated || child.revoked)
	// Security gate M1-02, condition 5.
	assert.True(t, old.absoluteExpiresAt.Equal(child.absoluteExpiresAt), "the family's absolute expiry is inherited")
}

func TestRefresh_lifetimeIsCappedByTheAbsoluteExpiry(t *testing.T) {
	s := newSessionStack(t)
	parent := s.session(t, anaEmail).Tokens.Refresh
	s.ownerUpdateTokens(t, "absolute_expires_at = now() + interval '2 hours', expires_at = now() + interval '2 hours'", parent)

	tokens, err := s.refresh.Execute(t.Context(), parent.Reveal())

	require.NoError(t, err)
	assert.InDelta(t, (2 * time.Hour).Seconds(), tokens.RefreshExpiresIn.Seconds(), 60)
	old, child := s.mustRow(t, parent), s.mustRow(t, tokens.Refresh)
	assert.True(t, child.expiresAt.Equal(old.absoluteExpiresAt), "the successor's expiry stops at the family's limit")
}

// US-A2b: reuse of a rotated token revokes the whole family, the successor
// included, and every later use of any of its tokens is refused.
func TestRefresh_reuseRevokesTheWholeFamily(t *testing.T) {
	s := newSessionStack(t)
	parent := s.session(t, anaEmail).Tokens.Refresh
	tokens, err := s.refresh.Execute(t.Context(), parent.Reveal())
	require.NoError(t, err)

	_, err = s.refresh.Execute(t.Context(), parent.Reveal())

	require.ErrorIs(t, err, usecase.ErrTokenReused)
	rows := s.familyRows(t, parent.UserID().String(), s.mustRow(t, parent).family)
	require.Len(t, rows, 2)
	for _, r := range rows {
		assert.True(t, r.revoked)
		assert.Equal(t, "reuse_detected", r.revokeReason)
	}
	_, err = s.refresh.Execute(t.Context(), tokens.Refresh.Reveal())
	require.ErrorIs(t, err, usecase.ErrTokenRevoked)
}

// An expired token that was never rotated is merely expired: 401 without
// revoking anything (security gate M1-02, condition 2).
func TestRefresh_expiredTokenDoesNotRevokeTheFamily(t *testing.T) {
	s := newSessionStack(t)
	token := s.session(t, anaEmail).Tokens.Refresh
	s.ownerUpdateTokens(t,
		"created_at = now() - interval '31 days', expires_at = now() - interval '1 day', absolute_expires_at = now() + interval '59 days'",
		token)

	_, err := s.refresh.Execute(t.Context(), token.Reveal())

	require.ErrorIs(t, err, usecase.ErrTokenExpired)
	r := s.mustRow(t, token)
	assert.False(t, r.revoked, "no family revoke")
	assert.False(t, r.rotated)
}

// A suspended account cannot keep a session alive with a live token
// (security gate M1-02, condition 3); a deleted one has no token left.
func TestRefresh_suspendedOrDeletedAccountCannotRefresh(t *testing.T) {
	t.Run("suspended", func(t *testing.T) {
		s := newSessionStack(t)
		token := s.session(t, anaEmail).Tokens.Refresh
		s.suspend(t, token.UserID())

		_, err := s.refresh.Execute(t.Context(), token.Reveal())

		require.ErrorIs(t, err, usecase.ErrAccountInactive)
		assert.False(t, s.mustRow(t, token).rotated, "nothing was rotated")
	})
	t.Run("deleted", func(t *testing.T) {
		s := newSessionStack(t)
		token := s.session(t, anaEmail).Tokens.Refresh
		s.asUser(t, token.UserID().String(), func(tx *sql.Tx) error {
			_, err := tx.ExecContext(t.Context(), "DELETE FROM users WHERE id = $1", token.UserID().String())
			return err
		})

		_, err := s.refresh.Execute(t.Context(), token.Reveal())

		require.ErrorIs(t, err, usecase.ErrRefreshRejected)
	})
}

// A token whose prefix names another user puts the lookup under that user,
// who owns no row with this hash: unknown, and the real owner's family is
// untouched.
func TestRefresh_forgedPrefixFindsNothing(t *testing.T) {
	s := newSessionStack(t)
	victim := s.session(t, anaEmail).Tokens.Refresh
	attacker := s.session(t, "jose@example.mx").Tokens.Refresh
	_, secret, _ := strings.Cut(victim.Reveal(), ".")
	forged := attacker.UserID().String() + "." + secret

	_, err := s.refresh.Execute(t.Context(), forged)

	require.ErrorIs(t, err, usecase.ErrTokenUnknown)
	assert.False(t, s.mustRow(t, victim).rotated || s.mustRow(t, victim).revoked)
}

// Concurrent refreshes of one token: the row lock serialises them, exactly
// one wins, every other is reuse, and the family ends up revoked (strict
// policy, usecase.RefreshSession).
func TestRefresh_concurrentDoubleSubmitHasOneWinnerThenRevokesTheFamily(t *testing.T) {
	const racers = 6
	s := newSessionStack(t)
	token := s.session(t, anaEmail).Tokens.Refresh

	var (
		wg       sync.WaitGroup
		start    = make(chan struct{})
		mu       sync.Mutex
		winners  []domain.RefreshToken
		refusals []error
	)
	for range racers {
		wg.Go(func() {
			<-start
			tokens, err := s.refresh.Execute(context.WithoutCancel(t.Context()), token.Reveal())
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				refusals = append(refusals, err)
				return
			}
			winners = append(winners, tokens.Refresh)
		})
	}
	close(start)
	wg.Wait()

	require.Len(t, winners, 1, "exactly one refresh wins")
	require.Len(t, refusals, racers-1)
	// The first loser detects the reuse and revokes the family; the ones
	// queued behind it then find the token revoked.
	reused := 0
	for _, err := range refusals {
		require.ErrorIs(t, err, usecase.ErrRefreshRejected)
		if errors.Is(err, usecase.ErrTokenReused) {
			reused++
		} else {
			assert.ErrorIs(t, err, usecase.ErrTokenRevoked)
		}
	}
	assert.Equal(t, 1, reused, "one reuse detection")
	rows := s.familyRows(t, token.UserID().String(), s.mustRow(t, token).family)
	assert.Len(t, rows, 2, "one rotation only")
	for _, r := range rows {
		assert.True(t, r.revoked, "the winner's token is revoked too")
	}
}

// Reuse detection racing the family's legitimate rotation: the revocation
// waits on the row the rotation locked, and the successor that rotation
// commits must not escape it. A READ COMMITTED UPDATE never sees rows
// committed after its statement began, so a single statement would leave
// the successor live: the thief (or the victim) keeps a session the reuse
// was meant to end.
func TestRefresh_reuseRevokesASuccessorCommittedWhileItWaited(t *testing.T) {
	s := newSessionStack(t)
	stolen := s.session(t, anaEmail).Tokens.Refresh
	tokens, err := s.refresh.Execute(t.Context(), stolen.Reveal())
	require.NoError(t, err)
	current := tokens.Refresh
	successor, err := domain.NewRefreshToken(current.UserID(), rand.Reader)
	require.NoError(t, err)

	locked, release, rotated := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		rotated <- s.txm.ReadWrite(context.WithoutCancel(t.Context()), current.UserID().String(), func(ctx context.Context) error {
			stored, found, err := s.store.Lock(ctx, current.Hash())
			if err != nil || !found {
				return errors.Join(err, errors.New("current token not found"))
			}
			if _, err := s.store.Rotate(ctx, stored, successor, usecase.RefreshSlidingWindow); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	<-locked
	reuse := make(chan error, 1)
	go func() {
		_, err := s.refresh.Execute(context.WithoutCancel(t.Context()), stolen.Reveal())
		reuse <- err
	}()
	s.awaitBlockedLock(t)
	close(release)

	require.NoError(t, <-rotated)
	require.ErrorIs(t, <-reuse, usecase.ErrTokenReused)
	for _, r := range s.familyRows(t, current.UserID().String(), s.mustRow(t, current).family) {
		assert.True(t, r.revoked, "every token of the family is revoked")
	}
	_, err = s.refresh.Execute(t.Context(), successor.Reveal())
	require.ErrorIs(t, err, usecase.ErrTokenRevoked)
}

// awaitBlockedLock waits until some backend waits on a lock.
func (s *sessionStack) awaitBlockedLock(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting int
		err := s.owner.QueryRowContext(t.Context(), "SELECT count(*) FROM pg_locks WHERE NOT granted").Scan(&waiting)
		return err == nil && waiting > 0
	}, 10*time.Second, 10*time.Millisecond, "the revocation waits on the rotation's row lock")
}

// A revocation stamps the statement's clock, not its transaction's start: a
// transaction that began before a newer token of the family was issued can
// still revoke it (refresh_tokens_revoked_after_issue would refuse now()).
func TestRevokeFamily_inATransactionOlderThanTheFamilysNewestToken(t *testing.T) {
	s := newSessionStack(t)
	parent := s.session(t, anaEmail).Tokens.Refresh
	var successor domain.RefreshToken

	err := s.txm.ReadWrite(t.Context(), parent.UserID().String(), func(ctx context.Context) error {
		// Another unit of work issues a successor after this one began.
		tokens, err := s.refresh.Execute(t.Context(), parent.Reveal())
		if err != nil {
			return err
		}
		successor = tokens.Refresh
		return s.store.RevokeFamily(ctx, parent.Hash(), usecase.RevokeLogout)
	})

	require.NoError(t, err)
	assert.True(t, s.mustRow(t, successor).revoked)
}

// US-A2c: logout revokes (never deletes) the family; a refresh afterwards
// fails.
func TestLogout_revokesTheFamilyThenRefreshFails(t *testing.T) {
	s := newSessionStack(t)
	first := s.session(t, anaEmail).Tokens.Refresh
	tokens, err := s.refresh.Execute(t.Context(), first.Reveal())
	require.NoError(t, err)
	current := tokens.Refresh

	err = s.logout.Execute(t.Context(), usecase.LogoutInput{UserID: current.UserID().String(), RefreshToken: current.Reveal()})

	require.NoError(t, err)
	rows := s.familyRows(t, current.UserID().String(), s.mustRow(t, current).family)
	require.Len(t, rows, 2, "revoked, not deleted")
	for _, r := range rows {
		assert.True(t, r.revoked)
		assert.Equal(t, "logout", r.revokeReason)
	}
	_, err = s.refresh.Execute(t.Context(), current.Reveal())
	require.ErrorIs(t, err, usecase.ErrTokenRevoked)

	t.Run("a second logout is a no-op success", func(t *testing.T) {
		err := s.logout.Execute(t.Context(), usecase.LogoutInput{UserID: current.UserID().String(), RefreshToken: current.Reveal()})
		require.NoError(t, err)
		assert.Equal(t, "logout", s.mustRow(t, current).revokeReason)
	})
}

func TestLogout_anotherUsersTokenIsUntouched(t *testing.T) {
	s := newSessionStack(t)
	victim := s.session(t, anaEmail).Tokens.Refresh
	caller := s.session(t, "jose@example.mx").Tokens.Refresh

	err := s.logout.Execute(t.Context(), usecase.LogoutInput{UserID: caller.UserID().String(), RefreshToken: victim.Reveal()})

	require.NoError(t, err)
	assert.False(t, s.mustRow(t, victim).revoked)
	_, err = s.refresh.Execute(t.Context(), victim.Reveal())
	require.NoError(t, err, "the victim's session still works")
}
