package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// ResendDeps are the ports ResendVerification needs.
type ResendDeps struct {
	System  tx.SystemManager
	Clock   clock.Clock
	Entropy io.Reader
	Index   EmailIndex
	Codes   CodeHasher
	SignUps CodeResender
	Email   EmailSender
	Logger  *slog.Logger
}

// ResendVerification emails a new code for a live pending sign-up (US-A1c,
// D2). It succeeds the same way whether a code was sent or not: no pending
// sign-up, an expired one, a verified email and the 60 s cooldown are silent.
type ResendVerification struct {
	deps   ResendDeps
	mailer mailer
}

// NewResendVerification builds the use case.
func NewResendVerification(deps ResendDeps) *ResendVerification {
	return &ResendVerification{deps: deps, mailer: mailer{sender: deps.Email, logger: deps.Logger}}
}

// Execute resends the code. Only a malformed email (domain.ErrInvalidEmail)
// is a client error.
func (uc *ResendVerification) Execute(ctx context.Context, rawEmail string) error {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return err
	}
	emailHash, err := hashEmail(uc.deps.Index, email)
	if err != nil {
		return err
	}
	issued, err := issueCode(uc.deps.Entropy, uc.deps.Codes, emailHash)
	if err != nil {
		return err
	}
	expiresAt := uc.deps.Clock.Now().Add(CodeTTL)
	var outcome CodeOutcome
	err = uc.deps.System.System(ctx, func(ctx context.Context) error {
		outcome, err = uc.deps.SignUps.Resend(ctx, emailHash, issued.hash, expiresAt)
		return err
	})
	if err != nil {
		return fmt.Errorf("auth: resend code: %w", err)
	}
	if outcome != CodeUnchanged {
		uc.mailer.sendCode(ctx, email, issued.code)
	}
	return nil
}
