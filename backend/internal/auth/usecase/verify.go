package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// VerifyDeps are the ports VerifyEmail needs.
type VerifyDeps struct {
	Tx      tx.Manager
	Clock   clock.Clock
	Entropy io.Reader
	// NewUserID draws the id of the account to create (UUIDv7 in production).
	NewUserID func() (domain.UserID, error)
	Index     EmailIndex
	// Passwords hashes the password the verifier chose (D8).
	Passwords PasswordHasher
	Codes     CodeHasher
	SignUps   SignUpVerifier
	Accounts  Accounts
	Trials    TrialStarter
	Refresh   RefreshTokens
	Access    AccessTokens
	// Logger receives the account_created audit event; nil discards it.
	Logger *slog.Logger
}

// VerifyInput is a verification request as the contract carries it. The
// password and the optional name are chosen here, by whoever holds the
// emailed code, never at register (D8).
type VerifyInput struct {
	Email    string
	Code     string
	Password string
	Name     *string // optional
}

// VerifyEmail confirms the emailed code and creates the account (US-A1b).
type VerifyEmail struct {
	deps     VerifyDeps
	sessions sessionIssuer
}

// NewVerifyEmail builds the use case.
func NewVerifyEmail(deps VerifyDeps) *VerifyEmail {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &VerifyEmail{deps: deps, sessions: sessionIssuer{entropy: deps.Entropy, refresh: deps.Refresh, access: deps.Access}}
}

// Execute checks the code and, on a match, creates the user with the
// password given here, its consents and counters, starts the trial if the
// email never had one, and signs the user in, all in one unit of work run as
// the new user (so RLS admits its rows).
//
// The password policy and the name are checked first and answer their domain
// errors (a client error that spends no try). Every failure about the email
// or the code is ErrVerificationRejected.
func (uc *VerifyEmail) Execute(ctx context.Context, in VerifyInput) (Session, error) {
	password, name, err := parseCredentials(in)
	if err != nil {
		return Session{}, err
	}
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return Session{}, ErrVerificationRejected
	}
	code, err := domain.ParseVerificationCode(in.Code)
	if err != nil {
		return Session{}, ErrVerificationRejected
	}
	emailHash, err := hashEmail(uc.deps.Index, email)
	if err != nil {
		return Session{}, err
	}
	userID, err := uc.deps.NewUserID()
	if err != nil {
		return Session{}, fmt.Errorf("auth: new user id: %w", err)
	}
	now := uc.deps.Clock.Now()
	attempt := verification{
		email: email, emailHash: emailHash, codeHash: uc.deps.Codes.HashCode(emailHash, code),
		password: password, name: name, userID: userID, now: now,
	}

	var (
		session  Session
		verified bool
	)
	err = uc.deps.Tx.ReadWrite(ctx, userID.String(), func(ctx context.Context) error {
		session, verified, err = uc.verify(ctx, attempt)
		return err
	})
	switch {
	case errors.Is(err, ErrAccountExists):
		return Session{}, ErrVerificationRejected
	case err != nil:
		return Session{}, err
	case !verified:
		return Session{}, ErrVerificationRejected
	}
	uc.deps.Logger.InfoContext(ctx, "auth audit", slog.String("event", "account_created"), slog.String("user_id", userID.String()))
	return session, nil
}

// parseCredentials validates what the verifier chooses for the account.
func parseCredentials(in VerifyInput) (domain.Password, domain.Name, error) {
	password, err := domain.NewPassword(in.Password)
	if err != nil {
		return domain.Password{}, domain.Name{}, err
	}
	var name domain.Name
	if in.Name != nil {
		if name, err = domain.NewName(*in.Name); err != nil {
			return domain.Password{}, domain.Name{}, err
		}
	}
	return password, name, nil
}

// verification is one checked attempt.
type verification struct {
	email     domain.Email
	emailHash []byte
	codeHash  []byte
	password  domain.Password
	name      domain.Name
	userID    domain.UserID
	now       time.Time
}

// verify runs inside the unit of work. A failed outcome returns nil so the
// unit of work commits: the attempt counter only sticks once committed. The
// password is hashed only after the code matched, so a wrong guess costs no
// argon2 run.
func (uc *VerifyEmail) verify(ctx context.Context, v verification) (Session, bool, error) {
	outcome, pending, err := uc.deps.SignUps.Consume(ctx, v.emailHash, v.codeHash)
	if err != nil {
		return Session{}, false, fmt.Errorf("auth: consume code: %w", err)
	}
	if outcome != ConsumeVerified {
		return Session{}, false, nil
	}
	passwordHash, err := uc.deps.Passwords.Hash(v.password.Reveal())
	if err != nil {
		return Session{}, false, fmt.Errorf("auth: hash password: %w", err)
	}
	account, err := uc.deps.Accounts.Create(ctx, NewAccount{
		ID:           v.userID,
		Email:        v.email,
		EmailHash:    v.emailHash,
		PasswordHash: passwordHash,
		Name:         v.name,
		Consents:     pending.Documents,
		AcceptedAt:   pending.AcceptedAt,
		VerifiedAt:   v.now,
	})
	if err != nil {
		return Session{}, false, fmt.Errorf("auth: create account: %w", err)
	}
	if err := uc.startTrial(ctx, v); err != nil {
		return Session{}, false, err
	}
	tokens, err := uc.sessions.issue(ctx, v.userID, v.now)
	if err != nil {
		return Session{}, false, err
	}
	return Session{Account: account, Tokens: tokens}, true, nil
}

// startTrial starts the trial only for an email that never had one: an
// account deleted and created again stays read-only until a plan.
func (uc *VerifyEmail) startTrial(ctx context.Context, v verification) error {
	first, err := uc.deps.SignUps.ClaimTrial(ctx, v.emailHash)
	if err != nil {
		return fmt.Errorf("auth: claim trial: %w", err)
	}
	if !first {
		return nil
	}
	if err := uc.deps.Trials.StartTrial(ctx, v.userID, v.now); err != nil {
		return fmt.Errorf("auth: start trial: %w", err)
	}
	return nil
}
