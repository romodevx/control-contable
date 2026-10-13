package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// issuedCode is a fresh verification code with its stored form.
type issuedCode struct {
	code domain.VerificationCode
	hash []byte
}

// issueCode draws a code from entropy and computes its code_hash.
func issueCode(entropy io.Reader, codes CodeHasher, emailHash []byte) (issuedCode, error) {
	code, err := domain.NewVerificationCode(entropy)
	if err != nil {
		return issuedCode{}, fmt.Errorf("auth: draw verification code: %w", err)
	}
	return issuedCode{code: code, hash: codes.HashCode(emailHash, code)}, nil
}

// hashEmail computes the blind index of a normalized email.
func hashEmail(index EmailIndex, email domain.Email) ([]byte, error) {
	h, err := index.Hash(email.String())
	if err != nil {
		return nil, fmt.Errorf("auth: email index: %w", err)
	}
	return h, nil
}

// mailer sends the auth emails. A delivery failure never changes the HTTP
// answer (that would be an oracle, MODULE-auth §4): it is logged, without the
// address or the code, and the user can ask for a resend.
type mailer struct {
	sender EmailSender
	logger *slog.Logger
}

func (m mailer) sendCode(ctx context.Context, to domain.Email, code domain.VerificationCode) {
	if err := m.sender.SendVerificationCode(ctx, to, code); err != nil {
		m.logger.WarnContext(ctx, "verification email not sent", slog.Any("error", err))
	}
}

func (m mailer) sendNotice(ctx context.Context, to domain.Email) {
	if err := m.sender.SendSignUpAttemptNotice(ctx, to); err != nil {
		m.logger.WarnContext(ctx, "sign-up attempt notice not sent", slog.Any("error", err))
	}
}
