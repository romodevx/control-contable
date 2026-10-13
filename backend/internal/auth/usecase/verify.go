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
	Codes     CodeHasher
	SignUps   SignUpVerifier
	Accounts  Accounts
	Trials    TrialStarter
	Refresh   RefreshTokens
	Access    AccessTokens
	// Logger receives the account_created audit event; nil discards it.
	Logger *slog.Logger
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

// Execute checks the code and, on a match, creates the user, its consents and
// counters, starts the trial if the email never had one, and signs the user in,
// all in one unit of work run as the new user (so RLS admits its rows). Every
// failure the caller could learn from is ErrVerificationRejected.
func (uc *VerifyEmail) Execute(ctx context.Context, rawEmail, rawCode string) (Session, error) {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return Session{}, ErrVerificationRejected
	}
	code, err := domain.ParseVerificationCode(rawCode)
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
	attempt := verification{email: email, emailHash: emailHash, codeHash: uc.deps.Codes.HashCode(emailHash, code), userID: userID, now: now}

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

// verification is one checked attempt.
type verification struct {
	email     domain.Email
	emailHash []byte
	codeHash  []byte
	userID    domain.UserID
	now       time.Time
}

// verify runs inside the unit of work. A failed outcome returns nil so the
// unit of work commits: the attempt counter only sticks once committed.
func (uc *VerifyEmail) verify(ctx context.Context, v verification) (Session, bool, error) {
	outcome, pending, err := uc.deps.SignUps.Consume(ctx, v.emailHash, v.codeHash)
	if err != nil {
		return Session{}, false, fmt.Errorf("auth: consume code: %w", err)
	}
	if outcome != ConsumeVerified {
		return Session{}, false, nil
	}
	account, err := uc.deps.Accounts.Create(ctx, NewAccount{
		ID:           v.userID,
		Email:        v.email,
		EmailHash:    v.emailHash,
		PasswordHash: pending.PasswordHash,
		Name:         pending.Name,
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
