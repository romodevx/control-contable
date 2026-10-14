package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// ProfileAccounts is what the profile operations need of the users table,
// in the unit of work of the signed-in user (RLS shows only that row).
type ProfileAccounts interface {
	// Find reads the account; ErrAccountNotFound when it does not exist.
	Find(ctx context.Context, id domain.UserID) (Account, error)
	// Rename stores name; ErrAccountNotFound when no row was updated.
	Rename(ctx context.Context, id domain.UserID, name domain.Name) error
}

// Profile reads and edits the signed-in user's own account (US-A3).
type Profile struct {
	tx       tx.Manager
	accounts ProfileAccounts
}

// NewProfile returns the use case.
func NewProfile(units tx.Manager, accounts ProfileAccounts) *Profile {
	return &Profile{tx: units, accounts: accounts}
}

// Get returns the account of userID (getMe). A deleted account is
// ErrAccountNotFound: its access tokens outlive it by up to 15 minutes.
func (uc *Profile) Get(ctx context.Context, userID domain.UserID) (Account, error) {
	var account Account
	err := uc.tx.ReadOnly(ctx, userID.String(), func(ctx context.Context) error {
		var err error
		account, err = uc.accounts.Find(ctx, userID)
		return err
	})
	if err != nil {
		return Account{}, fmt.Errorf("auth: read profile: %w", err)
	}
	return account, nil
}

// Rename sets the display name and returns the updated account (updateMe).
// An invalid name is domain.ErrInvalidName, before any database access.
func (uc *Profile) Rename(ctx context.Context, userID domain.UserID, rawName string) (Account, error) {
	name, err := domain.NewName(rawName)
	if err != nil {
		return Account{}, err
	}
	var account Account
	err = uc.tx.ReadWrite(ctx, userID.String(), func(ctx context.Context) error {
		if err := uc.accounts.Rename(ctx, userID, name); err != nil {
			return err
		}
		account, err = uc.accounts.Find(ctx, userID)
		return err
	})
	if err != nil {
		return Account{}, fmt.Errorf("auth: rename profile: %w", err)
	}
	return account, nil
}

// AccountRemover is what account deletion needs of the users table, in the
// unit of work of the signed-in user.
type AccountRemover interface {
	// PasswordHash reads the stored hash; ErrAccountNotFound when the
	// account does not exist.
	PasswordHash(ctx context.Context, id domain.UserID) (string, error)
	// Delete removes the account (its user-owned rows cascade) only while
	// its stored hash is still passwordHash; deleted is false when no row
	// matched.
	Delete(ctx context.Context, id domain.UserID, passwordHash string) (deleted bool, err error)
}

// PasswordChecker verifies a password against a stored argon2id hash.
type PasswordChecker interface {
	Verify(password, encoded string) (bool, error)
}

// DeleteAccountDeps are the ports DeleteAccount needs.
type DeleteAccountDeps struct {
	Tx        tx.Manager
	Passwords PasswordChecker
	Accounts  AccountRemover
	// Attempts admits the password attempts of each account, keyed by its
	// user id: a stolen access token must not become a password oracle.
	Attempts RateLimiter
	// Logger receives the login_failed and account_deleted audit events;
	// nil discards them.
	Logger *slog.Logger
}

// DeleteAccount deletes the signed-in user's account and everything it owns
// after re-checking the current password (US-A4, D4). Irreversible: the
// users row is hard-deleted and every user-owned table cascades (consents
// too, D6); only the trial_registry email hash survives, which has no
// foreign key to users.
type DeleteAccount struct {
	deps DeleteAccountDeps
}

// NewDeleteAccount returns the use case. It panics without deps.Attempts:
// deletion with no attempt limit is a wiring error, never a default.
func NewDeleteAccount(deps DeleteAccountDeps) *DeleteAccount {
	if deps.Attempts == nil {
		panic("auth: NewDeleteAccount needs an attempt RateLimiter")
	}
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &DeleteAccount{deps: deps}
}

// Execute verifies password and deletes the account of userID.
//
// The stored hash is read in a short read-only unit of work, argon2id runs
// outside any transaction (no row lock or pooled connection is held for its
// cost), and the delete is conditioned on the hash it verified: a password
// changed in between deletes nothing and answers ErrWrongPassword.
//
// Every call spends one of the account's attempts, before any database
// access or argon2id run (a success ends the account, so attempts are
// failures plus at most one); none left is a *RateLimitedError.
func (uc *DeleteAccount) Execute(ctx context.Context, userID domain.UserID, password string) error {
	allowed, wait, err := uc.deps.Attempts.Take(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("auth: deletion attempts: %w", err)
	}
	if !allowed {
		uc.audit(ctx, "delete_attempts_exhausted", userID)
		return &RateLimitedError{RetryAfter: wait}
	}
	candidate, err := domain.PasswordForVerification(password)
	if err != nil {
		uc.audit(ctx, "login_failed", userID)
		return ErrWrongPassword
	}
	hash, err := uc.storedHash(ctx, userID)
	if err != nil {
		return err
	}
	ok, err := uc.deps.Passwords.Verify(candidate.Reveal(), hash)
	if err != nil {
		return fmt.Errorf("auth: verify password for deletion: %w", err)
	}
	if !ok {
		uc.audit(ctx, "login_failed", userID)
		return ErrWrongPassword
	}
	return uc.delete(ctx, userID, hash)
}

func (uc *DeleteAccount) storedHash(ctx context.Context, userID domain.UserID) (string, error) {
	var hash string
	err := uc.deps.Tx.ReadOnly(ctx, userID.String(), func(ctx context.Context) error {
		var err error
		hash, err = uc.deps.Accounts.PasswordHash(ctx, userID)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("auth: read password hash: %w", err)
	}
	return hash, nil
}

func (uc *DeleteAccount) delete(ctx context.Context, userID domain.UserID, hash string) error {
	var deleted bool
	err := uc.deps.Tx.ReadWrite(ctx, userID.String(), func(ctx context.Context) error {
		var err error
		deleted, err = uc.deps.Accounts.Delete(ctx, userID, hash)
		return err
	})
	if err != nil {
		return fmt.Errorf("auth: delete account: %w", err)
	}
	if !deleted {
		// The password changed (or the account went) after the check.
		return ErrWrongPassword
	}
	uc.audit(ctx, "account_deleted", userID)
	return nil
}

// audit records an event naming the account by id only, never its email.
func (uc *DeleteAccount) audit(ctx context.Context, event string, userID domain.UserID) {
	uc.deps.Logger.InfoContext(ctx, "auth audit",
		slog.String("event", event), slog.String("user_id", userID.String()), slog.String("operation", "deleteMe"))
}
