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
	// Failures limits the failed logins of each email from each client
	// network (MODULE-auth §5), keyed by the email's blind index (whether or
	// not an account holds it) and LoginInput.Client: a stranger spending it
	// locks out only their own network (ADR-0005). Required.
	Failures ResettableLimiter
	// Ceiling limits the failed logins of each email from all networks
	// together, bounding distributed guessing. Required.
	Ceiling ResettableLimiter
	// Logger receives the login_failed audit event; nil discards it.
	Logger *slog.Logger
}

// LoginInput is a login request as the contract carries it.
type LoginInput struct {
	Email    string
	Password string
	// Client is the caller's network (IPv4 /24, IPv6 /64) as the HTTP edge
	// resolved it; "" when unknown (those share one budget).
	Client string
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
	if deps.Failures == nil || deps.Ceiling == nil {
		return nil, errors.New("auth: NewLogin needs Failures and Ceiling limiters: login without failure limits is a wiring error")
	}
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
//
// Before the lookup, each attempt spends one event of the email's failure
// budget for its client network, then one of the email's ceiling (both keyed
// by its blind index, for known and unknown addresses alike); none left is a
// *RateLimitedError, with no lookup and no argon2id run. A success refills
// both, so only failures count (MODULE-auth §5, ADR-0005).
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
	index := hex.EncodeToString(emailHash)
	budgets := []loginBudget{
		{uc.deps.Failures, index + " " + in.Client},
		{uc.deps.Ceiling, index},
	}
	for _, b := range budgets {
		if err := uc.spendAttempt(ctx, b); err != nil {
			return Session{}, err
		}
	}
	session, err := uc.authenticate(ctx, password, emailHash)
	if err != nil {
		return Session{}, err
	}
	for _, b := range budgets {
		if err := b.limiter.Reset(ctx, b.key); err != nil {
			// The session exists: a budget left partly spent only makes the
			// next failures count from where they were.
			uc.deps.Logger.WarnContext(ctx, "auth: login failure budget not refilled", slog.Any("error", err))
		}
	}
	return session, nil
}

// loginBudget is one failure budget of a login: a limiter and its key.
type loginBudget struct {
	limiter ResettableLimiter
	key     string
}

// spendAttempt takes one event of a failure budget.
func (uc *Login) spendAttempt(ctx context.Context, b loginBudget) error {
	allowed, wait, err := b.limiter.Take(ctx, b.key)
	if err != nil {
		return fmt.Errorf("auth: login failure budget: %w", err)
	}
	if !allowed {
		uc.deps.Logger.InfoContext(ctx, "auth audit", slog.String("event", "login_rate_limited"))
		return &RateLimitedError{RetryAfter: wait}
	}
	return nil
}

// authenticate checks the password of the account holding emailHash and
// starts its session.
func (uc *Login) authenticate(ctx context.Context, password domain.Password, emailHash []byte) (Session, error) {
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
