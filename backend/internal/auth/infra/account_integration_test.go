//go:build integration

package infra_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const beaEmail = "bea.lopez@example.mx"

// userOwnedTables are the tables account deletion must empty for the user,
// with the column naming the owner.
var userOwnedTables = map[string]string{
	"users":          "id",
	"user_usage":     "user_id",
	"entitlements":   "user_id",
	"consents":       "user_id",
	"refresh_tokens": "user_id",
}

// accountEnv is two verified accounts (Ana and Bea) and the store under test.
type accountEnv struct {
	*stack
	txm      *pgdb.TxManager
	store    *infra.AccountStore
	ana, bea domain.UserID
}

func newAccountEnv(t *testing.T) accountEnv {
	t.Helper()
	s := newStack(t)
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	parse := func(id string) domain.UserID {
		uid, err := domain.ParseUserID(id)
		require.NoError(t, err)
		return uid
	}
	return accountEnv{
		stack: s, txm: txm, store: infra.NewAccountStore(s.keyring),
		ana: parse(s.signUp(t, anaEmail, firstSample)), bea: parse(s.signUp(t, beaEmail, secondSample)),
	}
}

// as runs work in a read-write unit of work as viewer.
func (e accountEnv) as(t *testing.T, viewer domain.UserID, work tx.Work) error {
	t.Helper()
	return e.txm.ReadWrite(t.Context(), viewer.String(), work)
}

// owned counts the rows of each user-owned table that belong to userID, as
// that user (RLS shows a user their own rows).
func (e accountEnv) owned(t *testing.T, userID domain.UserID) map[string]int {
	t.Helper()
	counts := make(map[string]int, len(userOwnedTables))
	e.asUser(t, userID.String(), func(tx *sql.Tx) error {
		for table, column := range userOwnedTables {
			var n int
			query := "SELECT count(*) FROM " + table + " WHERE " + column + " = $1" // #nosec G202 -- fixed test identifiers
			if err := tx.QueryRowContext(t.Context(), query, userID.String()).Scan(&n); err != nil {
				return err
			}
			counts[table] = n
		}
		return nil
	})
	return counts
}

// inTrialRegistry reports whether email's hash is in trial_registry (read as
// cc_owner: the table is ENABLE, not FORCE, and service-only).
func (e accountEnv) inTrialRegistry(t *testing.T, email string) bool {
	t.Helper()
	hash, err := e.index.Hash(email)
	require.NoError(t, err)
	var n int
	require.NoError(t, e.owner.QueryRowContext(t.Context(), "SELECT count(*) FROM trial_registry WHERE email_hash = $1", hash).Scan(&n))
	return n == 1
}

func TestAccountStore_renameAndFindOwnRow(t *testing.T) {
	e := newAccountEnv(t)
	name, err := domain.NewName("Ana Gómez de la Torre")
	require.NoError(t, err)

	var got usecase.Account
	require.NoError(t, e.as(t, e.ana, func(ctx context.Context) error {
		if err := e.store.Rename(ctx, e.ana, name); err != nil {
			return err
		}
		got, err = e.store.Find(ctx, e.ana)
		return err
	}))

	assert.Equal(t, "Ana Gómez de la Torre", got.Name.String())
	assert.Equal(t, anaEmail, got.Email.String(), "decrypted from the sealed column")
}

func TestAccountStore_passwordHashOfOwnRow(t *testing.T) {
	e := newAccountEnv(t)

	var hash string
	require.NoError(t, e.as(t, e.ana, func(ctx context.Context) error {
		var err error
		hash, err = e.store.PasswordHash(ctx, e.ana)
		return err
	}))

	ok, err := e.passwords.Verify(firstSample, hash)
	require.NoError(t, err)
	assert.True(t, ok)
}

// RLS (defense in depth below the use cases, which only ever pass the
// caller's own id): A cannot read, rename or delete B.
func TestAccountStore_aUserCannotReachAnotherUsersAccount(t *testing.T) {
	e := newAccountEnv(t)
	name, err := domain.NewName("Intrusa")
	require.NoError(t, err)

	beaHash := e.storedHash(t, e.bea)

	err = e.as(t, e.ana, func(ctx context.Context) error {
		_, err := e.store.Find(ctx, e.bea)
		assert.ErrorIs(t, err, usecase.ErrAccountNotFound, "find")
		assert.ErrorIs(t, e.store.Rename(ctx, e.bea, name), usecase.ErrAccountNotFound, "rename")
		_, err = e.store.PasswordHash(ctx, e.bea)
		assert.ErrorIs(t, err, usecase.ErrAccountNotFound, "password hash")
		deleted, err := e.store.Delete(ctx, e.bea, beaHash)
		assert.False(t, deleted, "delete, even with the right hash")
		return err
	})
	require.NoError(t, err)

	bea := e.owned(t, e.bea)
	assert.Equal(t, 1, bea["users"])
	beaAccount, err := e.findAs(t, e.bea)
	require.NoError(t, err)
	assert.Equal(t, "Ana Gómez", beaAccount.Name.String(), "not renamed")
}

func (e accountEnv) findAs(t *testing.T, id domain.UserID) (usecase.Account, error) {
	t.Helper()
	var a usecase.Account
	err := e.as(t, id, func(ctx context.Context) error {
		var err error
		a, err = e.store.Find(ctx, id)
		return err
	})
	return a, err
}

func TestAccountStore_deleteCascadesEverythingButTheTrialRegistry(t *testing.T) {
	e := newAccountEnv(t)
	before := e.owned(t, e.ana)
	for table, n := range before {
		require.Positivef(t, n, "fixture: %s has rows of the user before the delete", table)
	}
	require.True(t, e.inTrialRegistry(t, anaEmail))
	beaBefore := e.owned(t, e.bea)
	hash := e.storedHash(t, e.ana)

	var deleted bool
	require.NoError(t, e.as(t, e.ana, func(ctx context.Context) error {
		var err error
		deleted, err = e.store.Delete(ctx, e.ana, hash)
		return err
	}))

	assert.True(t, deleted)
	for table, n := range e.owned(t, e.ana) {
		assert.Zerof(t, n, "%s still holds rows of the deleted user", table)
	}
	assert.True(t, e.inTrialRegistry(t, anaEmail), "the trial registry survives the account (no second trial)")
	assert.Equal(t, beaBefore, e.owned(t, e.bea), "another user's rows are untouched")
	_, err := e.findAs(t, e.ana)
	assert.ErrorIs(t, err, usecase.ErrAccountNotFound)
}

func TestAccountStore_deleteIsConditionedOnTheStoredHash(t *testing.T) {
	e := newAccountEnv(t)

	var deleted bool
	require.NoError(t, e.as(t, e.ana, func(ctx context.Context) error {
		var err error
		deleted, err = e.store.Delete(ctx, e.ana, e.storedHash(t, e.bea))
		return err
	}))

	assert.False(t, deleted)
	assert.Equal(t, 1, e.owned(t, e.ana)["users"])
}

func (e accountEnv) storedHash(t *testing.T, id domain.UserID) string {
	t.Helper()
	var hash string
	e.asUser(t, id.String(), func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT password_hash FROM users WHERE id = $1", id.String()).Scan(&hash)
	})
	return hash
}

// The re-registration rule end to end over the real definer functions:
// delete, register the same email again, verify, and the new account has
// no trial entitlement.
func TestDeleteAccount_reRegisteringTheSameEmailGetsNoNewTrial(t *testing.T) {
	e := newAccountEnv(t)
	remove := usecase.NewDeleteAccount(usecase.DeleteAccountDeps{Tx: e.txm, Passwords: e.passwords, Accounts: e.store, Attempts: e.limits.Bucket(usecase.DeleteAttemptsPolicy)})
	require.NoError(t, remove.Execute(t.Context(), e.ana, firstSample))

	again, err := domain.ParseUserID(e.signUp(t, anaEmail, secondSample))
	require.NoError(t, err)

	assert.NotEqual(t, e.ana, again)
	owned := e.owned(t, again)
	assert.Equal(t, 1, owned["users"])
	assert.Zero(t, owned["entitlements"], "no second trial")
}

func TestDeleteAccount_wrongPasswordDeletesNothing(t *testing.T) {
	e := newAccountEnv(t)
	remove := usecase.NewDeleteAccount(usecase.DeleteAccountDeps{Tx: e.txm, Passwords: e.passwords, Accounts: e.store, Attempts: e.limits.Bucket(usecase.DeleteAttemptsPolicy)})
	before := e.owned(t, e.ana)

	err := remove.Execute(t.Context(), e.ana, secondSample)

	assert.ErrorIs(t, err, usecase.ErrWrongPassword)
	assert.Equal(t, before, e.owned(t, e.ana))
}

func TestLegalDocumentStore_currentDocuments(t *testing.T) {
	s := newStack(t)
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	// A newer terms version already in effect and one not yet in effect.
	s.ownerExec(t, `INSERT INTO legal_documents (kind, version, url, effective_at) VALUES
  ('terms', 'v2', 'https://control-contable.invalid/legal/terminos-v2', now() - interval '1 minute'),
  ('terms', 'v3', 'https://control-contable.invalid/legal/terminos-v3', now() + interval '1 day')`)

	var docs []usecase.LegalDocument
	require.NoError(t, txm.System(t.Context(), func(ctx context.Context) error {
		docs, err = infra.NewLegalDocumentStore().Current(ctx)
		return err
	}))

	require.Len(t, docs, 3, "one per kind")
	byKind := make(map[domain.DocumentKind]usecase.LegalDocument, len(docs))
	for _, d := range docs {
		byKind[d.Kind] = d
	}
	assert.Equal(t, "v2", byKind[domain.DocumentTerms].Version, "the latest in effect, not the future one")
	assert.Equal(t, "https://control-contable.invalid/legal/terminos-v2", byKind[domain.DocumentTerms].URL)
	assert.Equal(t, "v1", byKind[domain.DocumentPrivacyIntegral].Version)
	assert.Equal(t, "https://control-contable.invalid/legal/aviso-de-privacidad-integral", byKind[domain.DocumentPrivacyIntegral].URL)
	assert.False(t, byKind[domain.DocumentPrivacySimplified].PublishedAt.IsZero())
	kinds := []domain.DocumentKind{docs[0].Kind, docs[1].Kind, docs[2].Kind}
	assert.IsIncreasing(t, kinds, "stable order by kind")
}
