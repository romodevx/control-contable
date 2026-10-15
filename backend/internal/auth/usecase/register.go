package usecase

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// RegisterInput is a registration request as the contract carries it: no
// password and no name, both are set at verification (D8).
type RegisterInput struct {
	Email             string
	AcceptedDocuments []domain.DocumentVersion
}

// RegisterDeps are the ports Register needs.
type RegisterDeps struct {
	System  tx.SystemManager
	Clock   clock.Clock
	Entropy io.Reader // crypto/rand.Reader in production
	Index   EmailIndex
	Codes   CodeHasher
	Legal   LegalDocuments
	SignUps SignUpStarter
	// Outbox queues the code or the notice in the sign-up's own unit of
	// work; the worker delivers it (no provider call in the request).
	Outbox EmailOutbox
	// Notices throttles the "someone tried to sign up" notice per email (D3).
	Notices Throttle
	// Emails limits the sign-up attempts of each email (MODULE-auth §5),
	// keyed by its blind index; resendVerification spends from the same
	// budget (D7). Required.
	Emails RateLimiter
	Logger *slog.Logger
}

// Register starts a sign-up (US-A1a). Its only observable results are a
// validation error or success: success looks the same for a new email, a
// pending sign-up and a verified account (no enumeration). No branch hashes a
// password (register takes none, D8) or talks to an email provider (the email
// is queued in the same unit of work, EmailOutbox), so none is slower.
type Register struct {
	deps RegisterDeps
}

// NewRegister builds the use case. It panics without deps.Emails: sign-up
// with no per-email limit is a wiring error, never a default.
func NewRegister(deps RegisterDeps) *Register {
	if deps.Emails == nil {
		panic("auth: NewRegister needs an Emails RateLimiter")
	}
	return &Register{deps: deps}
}

// Execute validates the input, then starts (or, for a live pending sign-up,
// resends) the sign-up, or notifies the owner of a verified address. The
// domain validation errors (email, documents) are the only errors a client
// should see; anything else is a server fault.
func (uc *Register) Execute(ctx context.Context, in RegisterInput) error {
	req, err := parseRegistration(in)
	if err != nil {
		return err
	}
	emailHash, err := hashEmail(uc.deps.Index, req.email)
	if err != nil {
		return err
	}
	allowed, wait, err := uc.deps.Emails.Take(ctx, hex.EncodeToString(emailHash))
	if err != nil {
		return fmt.Errorf("auth: sign-up budget: %w", err)
	}
	if !allowed {
		return &RateLimitedError{RetryAfter: wait}
	}
	issued, err := issueCode(uc.deps.Entropy, uc.deps.Codes, emailHash)
	if err != nil {
		return err
	}
	now := uc.deps.Clock.Now()
	pending := PendingSignUp{Documents: req.acceptance.Documents(), AcceptedAt: now}
	return uc.deps.System.System(ctx, func(ctx context.Context) error {
		email, err := uc.start(ctx, req, emailHash, issued, pending)
		if err != nil || email == nil {
			return err
		}
		return enqueue(ctx, uc.deps.Outbox, *email, now)
	})
}

// start runs inside the unit of work and returns the email to queue, if
// any. The document check comes first, so a stale version answers 400
// whatever the email's state.
func (uc *Register) start(ctx context.Context, req registration, emailHash []byte, issued issuedCode, pending PendingSignUp) (*OutboundEmail, error) {
	current, err := uc.deps.Legal.CurrentVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth: current legal documents: %w", err)
	}
	if err := req.acceptance.CheckCurrent(current); err != nil {
		return nil, err
	}
	exists, err := uc.deps.SignUps.AccountExists(ctx, emailHash)
	if err != nil {
		return nil, fmt.Errorf("auth: look up account: %w", err)
	}
	if exists {
		if allowed, _ := uc.deps.Notices.Allow(hex.EncodeToString(emailHash)); !allowed {
			return nil, nil
		}
		return &OutboundEmail{Kind: EmailSignUpAttemptNotice, To: req.email}, nil
	}
	outcome, err := uc.deps.SignUps.Start(ctx, emailHash, issued.hash, pending, pending.AcceptedAt.Add(CodeTTL))
	if err != nil {
		return nil, fmt.Errorf("auth: start sign-up: %w", err)
	}
	if outcome == CodeUnchanged {
		return nil, nil
	}
	uc.deps.Logger.InfoContext(ctx, "auth audit", slog.String("event", "register_started"))
	return &OutboundEmail{Kind: EmailVerificationCode, To: req.email, Code: issued.code}, nil
}

// registration is a validated RegisterInput.
type registration struct {
	email      domain.Email
	acceptance domain.Acceptance
}

func parseRegistration(in RegisterInput) (registration, error) {
	var (
		r   registration
		err error
	)
	if r.email, err = domain.NewEmail(in.Email); err != nil {
		return registration{}, err
	}
	if r.acceptance, err = domain.NewAcceptance(in.AcceptedDocuments); err != nil {
		return registration{}, err
	}
	return r, nil
}
