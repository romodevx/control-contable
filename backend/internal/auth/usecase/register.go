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
	Email   EmailSender
	// Notices throttles the "someone tried to sign up" notice per email (D3).
	Notices Throttle
	Logger  *slog.Logger
}

// Register starts a sign-up (US-A1a). Its only observable results are a
// validation error or success: success looks the same for a new email, a
// pending sign-up and a verified account (no enumeration). No branch hashes a
// password (register takes none, D8), so none is slower for it.
type Register struct {
	deps   RegisterDeps
	mailer mailer
}

// NewRegister builds the use case.
func NewRegister(deps RegisterDeps) *Register {
	return &Register{deps: deps, mailer: mailer{sender: deps.Email, logger: deps.Logger}}
}

// registerOutcome is what the unit of work decided to send.
type registerOutcome int

const (
	sendNothing registerOutcome = iota
	sendCode
	sendNotice
)

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
	issued, err := issueCode(uc.deps.Entropy, uc.deps.Codes, emailHash)
	if err != nil {
		return err
	}
	now := uc.deps.Clock.Now()
	pending := PendingSignUp{Documents: req.acceptance.Documents(), AcceptedAt: now}

	var outcome registerOutcome
	err = uc.deps.System.System(ctx, func(ctx context.Context) error {
		outcome, err = uc.start(ctx, req.acceptance, emailHash, issued.hash, pending)
		return err
	})
	if err != nil {
		return err
	}
	switch outcome {
	case sendCode:
		uc.mailer.sendCode(ctx, req.email, issued.code)
	case sendNotice:
		if allowed, _ := uc.deps.Notices.Allow(hex.EncodeToString(emailHash)); allowed {
			uc.mailer.sendNotice(ctx, req.email)
		}
	case sendNothing:
	}
	return nil
}

// start runs inside the unit of work. The document check comes first, so a
// stale version answers 400 whatever the email's state.
func (uc *Register) start(ctx context.Context, acceptance domain.Acceptance, emailHash, codeHash []byte, pending PendingSignUp) (registerOutcome, error) {
	current, err := uc.deps.Legal.CurrentVersions(ctx)
	if err != nil {
		return sendNothing, fmt.Errorf("auth: current legal documents: %w", err)
	}
	if err := acceptance.CheckCurrent(current); err != nil {
		return sendNothing, err
	}
	exists, err := uc.deps.SignUps.AccountExists(ctx, emailHash)
	if err != nil {
		return sendNothing, fmt.Errorf("auth: look up account: %w", err)
	}
	if exists {
		return sendNotice, nil
	}
	outcome, err := uc.deps.SignUps.Start(ctx, emailHash, codeHash, pending, pending.AcceptedAt.Add(CodeTTL))
	if err != nil {
		return sendNothing, fmt.Errorf("auth: start sign-up: %w", err)
	}
	if outcome == CodeUnchanged {
		return sendNothing, nil
	}
	uc.deps.Logger.InfoContext(ctx, "auth audit", slog.String("event", "register_started"))
	return sendCode, nil
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
