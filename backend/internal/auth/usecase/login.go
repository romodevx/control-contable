package usecase

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// dummySecretBytes is the size of the random phrase behind the dummy hash.
const dummySecretBytes = 32

// LoginDeps are the ports Login needs.
type LoginDeps struct {
	// System runs the credential lookup, a definer function, with no user
	// context (tx.SystemManager is allowed on the sign-in paths only, D-2).
	System tx.SystemManager
	// Tx runs the session writes as the user the password proved.
	Tx          tx.Manager
	Clock       clock.Clock
	Entropy     io.Reader
	Index       EmailIndex
	Passwords   PasswordVerifier
	Credentials CredentialLookup
	Accounts    LoginAccounts
	Refresh     RefreshTokens
	Access      AccessTokens
	// Logger receives the login_failed audit event; nil discards it.
	Logger *slog.Logger
}

// LoginInput is a login request as the contract carries it.
type LoginInput struct {
	Email    string
	Password string
}

// Login signs an account in with email and password (US-A2a).
type Login struct {
	deps     LoginDeps
	sessions sessionIssuer
	// dummyHash is verified when there is no stored hash to verify (an
	// unknown email), so that path costs one argon2id run like the others.
	dummyHash string
}

// NewLogin builds the use case. It hashes a random phrase under the current
// password policy once, for the unknown-email path.
func NewLogin(deps LoginDeps) (*Login, error) {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	secret := make([]byte, dummySecretBytes)
	if _, err := io.ReadFull(deps.Entropy, secret); err != nil {
		return nil, fmt.Errorf("auth: dummy password: %w", err)
	}
	dummyHash, err := deps.Passwords.Hash(hex.EncodeToString(secret))
	if err != nil {
		return nil, fmt.Errorf("auth: dummy password hash: %w", err)
	}
	return &Login{
		deps:      deps,
		sessions:  sessionIssuer{entropy: deps.Entropy, refresh: deps.Refresh, access: deps.Access},
		dummyHash: dummyHash,
	}, nil
}

// Execute checks the credentials and starts a new token family.
//
// Every way of failing (unknown or invalid email, wrong password, an account
// that may not sign in, a stored hash that cannot be read) is
// ErrInvalidCredentials after exactly one argon2id verification under the
// current policy, so neither the answer nor its latency enumerates accounts.
// Only an infrastructure failure is a different error.
func (uc *Login) Execute(ctx context.Context, in LoginInput) (Session, error) {
	password, passwordErr := domain.PasswordForVerification(in.Password)
	email, emailErr := domain.NewEmail(in.Email)
	if passwordErr != nil || emailErr != nil {
		return Session{}, uc.reject(ctx, password, domain.UserID{})
	}
	emailHash, err := hashEmail(uc.deps.Index, email)
	if err != nil {
		return Session{}, err
	}
	creds, found, err := uc.lookup(ctx, emailHash)
	if err != nil {
		return Session{}, err
	}
	if !found {
		return Session{}, uc.reject(ctx, password, domain.UserID{})
	}
	if !uc.verify(ctx, password, creds) {
		uc.audit(ctx, creds.UserID)
		return Session{}, ErrInvalidCredentials
	}
	return uc.signIn(ctx, creds, uc.rehash(ctx, password, creds))
}

func (uc *Login) lookup(ctx context.Context, emailHash []byte) (Credentials, bool, error) {
	var (
		creds Credentials
		found bool
	)
	err := uc.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		creds, found, err = uc.deps.Credentials.Lookup(ctx, emailHash)
		return err
	})
	if err != nil {
		return Credentials{}, false, fmt.Errorf("auth: look up credentials: %w", err)
	}
	return creds, found, nil
}

// reject spends the dummy verification and answers ErrInvalidCredentials.
func (uc *Login) reject(ctx context.Context, password domain.Password, userID domain.UserID) error {
	_, _ = uc.deps.Passwords.Verify(password.Reveal(), uc.dummyHash) // only its cost matters
	uc.audit(ctx, userID)
	return ErrInvalidCredentials
}

// verify reports whether the password opens an account allowed to sign in.
// The status is checked after the verification so a suspended account costs
// the same as any other.
func (uc *Login) verify(ctx context.Context, password domain.Password, creds Credentials) bool {
	ok, err := uc.deps.Passwords.Verify(password.Reveal(), creds.PasswordHash)
	if err != nil {
		uc.deps.Logger.ErrorContext(ctx, "auth: unreadable stored password hash",
			slog.String("user_id", creds.UserID.String()), slog.Any("error", err))
		return false
	}
	return ok && creds.Active
}

// rehash returns the password hashed under the current policy when the
// stored hash was made with other parameters, or "" when it need not (or
// cannot) be replaced: a failed rehash never fails the login.
func (uc *Login) rehash(ctx context.Context, password domain.Password, creds Credentials) string {
	needs, err := uc.deps.Passwords.NeedsRehash(creds.PasswordHash)
	if err != nil || !needs {
		return ""
	}
	hash, err := uc.deps.Passwords.Hash(password.Reveal())
	if err != nil {
		uc.deps.Logger.WarnContext(ctx, "auth: password rehash failed",
			slog.String("user_id", creds.UserID.String()), slog.Any("error", err))
		return ""
	}
	return hash
}

// signIn stores the rehashed password (if any), reads the account and
// starts a token family, in one unit of work as the user.
func (uc *Login) signIn(ctx context.Context, creds Credentials, newHash string) (Session, error) {
	now := uc.deps.Clock.Now()
	var session Session
	err := uc.deps.Tx.ReadWrite(ctx, creds.UserID.String(), func(ctx context.Context) error {
		if newHash != "" {
			if err := uc.deps.Accounts.ReplacePasswordHash(ctx, creds.UserID, creds.PasswordHash, newHash); err != nil {
				return fmt.Errorf("auth: rehash password: %w", err)
			}
		}
		account, err := uc.deps.Accounts.Find(ctx, creds.UserID)
		if err != nil {
			return fmt.Errorf("auth: read account: %w", err)
		}
		tokens, err := uc.sessions.issue(ctx, creds.UserID, now)
		if err != nil {
			return err
		}
		session = Session{Account: account, Tokens: tokens}
		return nil
	})
	if errors.Is(err, ErrAccountNotFound) {
		// Deleted between the lookup and the session.
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

// audit records a failed login, naming the account only when one exists;
// never the email.
func (uc *Login) audit(ctx context.Context, userID domain.UserID) {
	attrs := []any{slog.String("event", "login_failed")}
	if !userID.IsZero() {
		attrs = append(attrs, slog.String("user_id", userID.String()))
	}
	uc.deps.Logger.InfoContext(ctx, "auth audit", attrs...)
}
