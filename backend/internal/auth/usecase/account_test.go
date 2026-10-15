package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// uow runs work inline and records each unit of work's kind and user, so a
// test can see which ran read-only and which read-write.
type uow struct {
	units []string // "ro:<user>" or "rw:<user>"
	err   error
}

func (u *uow) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	u.units = append(u.units, "rw:"+userID)
	if u.err != nil {
		return u.err
	}
	return work(ctx)
}

func (u *uow) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	u.units = append(u.units, "ro:"+userID)
	if u.err != nil {
		return u.err
	}
	return work(ctx)
}

var _ tx.Manager = (*uow)(nil)

// fakeProfiles is the users table of one account, as RLS shows it to the
// unit of work's user.
type fakeProfiles struct {
	account usecase.Account
	hash    string
	exists  bool
	err     error

	renamed []string
	deleted []string // the password hashes Delete was conditioned on
}

func newFakeProfiles(t *testing.T) *fakeProfiles {
	t.Helper()
	id, err := domain.ParseUserID(newUserID)
	require.NoError(t, err)
	email, err := domain.NewEmail(normalEmail)
	require.NoError(t, err)
	return &fakeProfiles{
		account: usecase.Account{ID: id, Email: email, CreatedAt: now},
		hash:    "argon:p1:" + acceptedSample,
		exists:  true,
	}
}

func (f *fakeProfiles) Find(_ context.Context, id domain.UserID) (usecase.Account, error) {
	switch {
	case f.err != nil:
		return usecase.Account{}, f.err
	case !f.exists || id != f.account.ID:
		return usecase.Account{}, usecase.ErrAccountNotFound
	}
	return f.account, nil
}

func (f *fakeProfiles) Rename(_ context.Context, id domain.UserID, name domain.Name) error {
	switch {
	case f.err != nil:
		return f.err
	case !f.exists || id != f.account.ID:
		return usecase.ErrAccountNotFound
	}
	f.renamed = append(f.renamed, name.String())
	f.account.Name = name
	return nil
}

func (f *fakeProfiles) PasswordHash(_ context.Context, id domain.UserID) (string, error) {
	switch {
	case f.err != nil:
		return "", f.err
	case !f.exists || id != f.account.ID:
		return "", usecase.ErrAccountNotFound
	}
	return f.hash, nil
}

func (f *fakeProfiles) Delete(_ context.Context, id domain.UserID, passwordHash string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	f.deleted = append(f.deleted, passwordHash)
	if !f.exists || id != f.account.ID || passwordHash != f.hash {
		return false, nil
	}
	f.exists = false
	return true, nil
}

func TestProfile_get(t *testing.T) {
	t.Run("reads the account read-only as the caller", func(t *testing.T) {
		accounts, units := newFakeProfiles(t), &uow{}
		uc := usecase.NewProfile(units, accounts)

		got, err := uc.Get(t.Context(), accounts.account.ID)

		require.NoError(t, err)
		assert.Equal(t, accounts.account, got)
		assert.Equal(t, []string{"ro:" + newUserID}, units.units)
	})

	t.Run("a deleted account is ErrAccountNotFound", func(t *testing.T) {
		accounts := newFakeProfiles(t)
		accounts.exists = false

		_, err := usecase.NewProfile(&uow{}, accounts).Get(t.Context(), accounts.account.ID)

		assert.ErrorIs(t, err, usecase.ErrAccountNotFound)
	})

	t.Run("an infrastructure failure is wrapped", func(t *testing.T) {
		accounts := newFakeProfiles(t)
		accounts.err = errDB

		_, err := usecase.NewProfile(&uow{}, accounts).Get(t.Context(), accounts.account.ID)

		assert.ErrorIs(t, err, errDB)
		assert.NotErrorIs(t, err, usecase.ErrAccountNotFound)
	})
}

func TestProfile_rename(t *testing.T) {
	t.Run("stores the normalized name and returns the account", func(t *testing.T) {
		accounts, units := newFakeProfiles(t), &uow{}
		uc := usecase.NewProfile(units, accounts)

		got, err := uc.Rename(t.Context(), accounts.account.ID, "  Ana Gómez de la Torre ")

		require.NoError(t, err)
		assert.Equal(t, "Ana Gómez de la Torre", got.Name.String())
		assert.Equal(t, normalEmail, got.Email.String())
		assert.Equal(t, []string{"Ana Gómez de la Torre"}, accounts.renamed)
		assert.Equal(t, []string{"rw:" + newUserID}, units.units, "one read-write unit of work")
	})

	for name, tt := range map[string]struct {
		raw  string
		want error
	}{
		"blank":             {"   ", domain.ErrInvalidName},
		"too long":          {strings.Repeat("a", domain.MaxNameLength+1), domain.ErrInvalidName},
		"control-character": {"Ana\u0007", domain.ErrNameInvisible},
		"bidi-override":     {"Ana\u202eB", domain.ErrNameInvisible},
	} {
		t.Run("a "+name+" name is refused and opens no unit of work", func(t *testing.T) {
			accounts, units := newFakeProfiles(t), &uow{}

			_, err := usecase.NewProfile(units, accounts).Rename(t.Context(), accounts.account.ID, tt.raw)

			assert.ErrorIs(t, err, tt.want)
			assert.Empty(t, units.units)
			assert.Empty(t, accounts.renamed)
		})
	}

	t.Run("a deleted account is ErrAccountNotFound", func(t *testing.T) {
		accounts := newFakeProfiles(t)
		accounts.exists = false

		_, err := usecase.NewProfile(&uow{}, accounts).Rename(t.Context(), accounts.account.ID, "Ana")

		assert.ErrorIs(t, err, usecase.ErrAccountNotFound)
	})
}

func newDeleteAccount(units tx.Manager, accounts *fakeProfiles, passwords *fakePasswords) (*usecase.DeleteAccount, func() string) {
	return newThrottledDeleteAccount(units, accounts, passwords, &fakeLimiter{limit: 100})
}

func newThrottledDeleteAccount(units tx.Manager, accounts *fakeProfiles, passwords *fakePasswords, attempts usecase.RateLimiter) (*usecase.DeleteAccount, func() string) {
	logger, logs := capturedLogs()
	return usecase.NewDeleteAccount(usecase.DeleteAccountDeps{
		Tx: units, Passwords: passwords, Accounts: accounts, Attempts: attempts, Logger: logger,
	}), logs.String
}

func TestDeleteAccount(t *testing.T) {
	t.Run("the right password deletes the account, conditioned on the hash it verified", func(t *testing.T) {
		accounts, units, passwords := newFakeProfiles(t), &uow{}, &fakePasswords{params: "p1"}
		uc, logs := newDeleteAccount(units, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		require.NoError(t, err)
		assert.False(t, accounts.exists)
		assert.Equal(t, []string{accounts.hash}, accounts.deleted)
		assert.Equal(t, []string{accounts.hash}, passwords.verified, "one argon2id verification")
		// The hash is read in a short read-only unit of work; argon2id runs
		// outside any transaction; the delete is its own unit of work.
		assert.Equal(t, []string{"ro:" + newUserID, "rw:" + newUserID}, units.units)
		assert.Contains(t, logs(), `"event":"account_deleted"`)
		assert.Contains(t, logs(), newUserID)
		assert.NotContains(t, logs(), acceptedSample)
		assert.NotContains(t, logs(), normalEmail)
	})

	t.Run("a wrong password is ErrWrongPassword, deletes nothing and is audited as a failed login", func(t *testing.T) {
		accounts, units, passwords := newFakeProfiles(t), &uow{}, &fakePasswords{params: "p1"}
		uc, logs := newDeleteAccount(units, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, "otra frase equivocada")

		assert.ErrorIs(t, err, usecase.ErrWrongPassword)
		assert.True(t, accounts.exists)
		assert.Empty(t, accounts.deleted)
		assert.Equal(t, []string{"ro:" + newUserID}, units.units, "no read-write unit of work")
		assert.Contains(t, logs(), `"event":"login_failed"`)
		assert.NotContains(t, logs(), "otra frase equivocada")
	})

	t.Run("a password changed between the check and the delete deletes nothing", func(t *testing.T) {
		accounts, passwords := newFakeProfiles(t), &fakePasswords{params: "p1"}
		units := &racingUOW{onWrite: func() { accounts.hash = "argon:p1:una frase nueva distinta" }}
		uc, _ := newDeleteAccount(units, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		assert.ErrorIs(t, err, usecase.ErrWrongPassword)
		assert.True(t, accounts.exists)
	})

	t.Run("an unusable password is ErrWrongPassword without touching the database", func(t *testing.T) {
		accounts, units, passwords := newFakeProfiles(t), &uow{}, &fakePasswords{params: "p1"}
		uc, _ := newDeleteAccount(units, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, string([]byte{0xff, 0xfe}))

		assert.ErrorIs(t, err, usecase.ErrWrongPassword)
		assert.Empty(t, units.units)
		assert.Empty(t, passwords.verified)
	})

	t.Run("an account already gone is ErrAccountNotFound", func(t *testing.T) {
		accounts, passwords := newFakeProfiles(t), &fakePasswords{params: "p1"}
		accounts.exists = false
		uc, _ := newDeleteAccount(&uow{}, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		assert.ErrorIs(t, err, usecase.ErrAccountNotFound)
		assert.Empty(t, passwords.verified)
	})

	t.Run("an unreadable stored hash is a server fault, not a wrong password", func(t *testing.T) {
		accounts, passwords := newFakeProfiles(t), &fakePasswords{params: "p1"}
		accounts.hash = "not-a-phc-string"
		uc, _ := newDeleteAccount(&uow{}, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		require.Error(t, err)
		assert.NotErrorIs(t, err, usecase.ErrWrongPassword)
		assert.True(t, accounts.exists)
	})

	t.Run("every attempt spends one of the account's password attempts", func(t *testing.T) {
		accounts, passwords := newFakeProfiles(t), &fakePasswords{params: "p1"}
		attempts := &fakeLimiter{limit: 2}
		uc, _ := newThrottledDeleteAccount(&uow{}, accounts, passwords, attempts)

		for range 2 {
			assert.ErrorIs(t, uc.Execute(t.Context(), accounts.account.ID, "otra frase equivocada"), usecase.ErrWrongPassword)
		}

		assert.Equal(t, []string{newUserID, newUserID}, attempts.taken(), "keyed by the account, never by the request")
	})

	t.Run("an account out of attempts is refused before reading the hash or running argon2id", func(t *testing.T) {
		accounts, units, passwords := newFakeProfiles(t), &uow{}, &fakePasswords{params: "p1"}
		uc, logs := newThrottledDeleteAccount(units, accounts, passwords, &fakeLimiter{limit: 0})

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		var tooMany *usecase.RateLimitedError
		require.ErrorAs(t, err, &tooMany)
		assert.Equal(t, time.Minute, tooMany.RetryAfter)
		assert.NotErrorIs(t, err, usecase.ErrWrongPassword)
		assert.True(t, accounts.exists, "not even the right password deletes")
		assert.Empty(t, units.units)
		assert.Empty(t, passwords.verified)
		assert.Contains(t, logs(), `"event":"delete_attempts_exhausted"`)
	})

	t.Run("a limiter that cannot decide fails closed: nothing read, verified or deleted", func(t *testing.T) {
		accounts, units, passwords := newFakeProfiles(t), &uow{}, &fakePasswords{params: "p1"}
		uc, _ := newThrottledDeleteAccount(units, accounts, passwords, &fakeLimiter{err: errDB})

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		require.ErrorIs(t, err, errDB)
		assert.True(t, accounts.exists)
		assert.Empty(t, units.units)
		assert.Empty(t, passwords.verified)
	})

	t.Run("an infrastructure failure is wrapped and deletes nothing", func(t *testing.T) {
		accounts, passwords := newFakeProfiles(t), &fakePasswords{params: "p1"}
		units := &uow{err: errDB}
		uc, _ := newDeleteAccount(units, accounts, passwords)

		err := uc.Execute(t.Context(), accounts.account.ID, acceptedSample)

		assert.ErrorIs(t, err, errDB)
		assert.True(t, accounts.exists)
	})
}

// racingUOW is a uow whose onWrite runs just before the read-write unit of
// work: a concurrent password change landing after the verification.
type racingUOW struct {
	uow
	onWrite func()
}

func (r *racingUOW) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	r.onWrite()
	return r.uow.ReadWrite(ctx, userID, work)
}

func TestNewDeleteAccount_refusesToRunWithoutAnAttemptLimit(t *testing.T) {
	assert.PanicsWithValue(t, "auth: NewDeleteAccount needs an attempt RateLimiter", func() {
		usecase.NewDeleteAccount(usecase.DeleteAccountDeps{Tx: &uow{}, Passwords: &fakePasswords{}, Accounts: newFakeProfiles(t)})
	})
}

// fakeCatalog serves the current legal documents.
type fakeCatalog struct {
	docs []usecase.LegalDocument
	err  error
}

func (f fakeCatalog) Current(context.Context) ([]usecase.LegalDocument, error) { return f.docs, f.err }

func TestListLegalDocuments(t *testing.T) {
	t.Run("returns the current documents through a system unit of work", func(t *testing.T) {
		docs := []usecase.LegalDocument{
			{Kind: domain.DocumentPrivacyIntegral, Version: "v1", URL: "https://example.mx/a", PublishedAt: now},
			{Kind: domain.DocumentTerms, Version: "v2", URL: "https://example.mx/t", PublishedAt: now},
		}
		system := &fakeSystem{}

		got, err := usecase.NewListLegalDocuments(system, fakeCatalog{docs: docs}).Execute(t.Context())

		require.NoError(t, err)
		assert.Equal(t, docs, got)
		assert.Equal(t, 1, system.calls)
	})

	t.Run("a failure is wrapped", func(t *testing.T) {
		_, err := usecase.NewListLegalDocuments(&fakeSystem{}, fakeCatalog{err: errDB}).Execute(t.Context())

		assert.ErrorIs(t, err, errDB)
	})

	t.Run("an empty catalog is a server fault: sign-up could never succeed", func(t *testing.T) {
		_, err := usecase.NewListLegalDocuments(&fakeSystem{}, fakeCatalog{}).Execute(t.Context())

		assert.True(t, errors.Is(err, usecase.ErrNoLegalDocuments))
	})
}
