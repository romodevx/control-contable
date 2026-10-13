package infra

import (
	"context"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

// LogSender is the EmailSender until a provider is chosen (M1-06, OB2). In
// dev it writes the verification code to the log, so a developer can finish
// a sign-up locally. In every other environment it never logs the code: it
// only records that the email was dropped. The address is passed to the
// logger as a value the redacting handler masks, never in clear.
type LogSender struct {
	logger      *slog.Logger
	revealCodes bool
}

// NewLogSender returns the log sender for env. Only config.EnvDev reveals
// codes; any other value, unknown ones included, keeps them out of the logs.
func NewLogSender(logger *slog.Logger, env config.Environment) *LogSender {
	return &LogSender{logger: logger, revealCodes: env == config.EnvDev}
}

// SendVerificationCode implements usecase.EmailSender.
func (s *LogSender) SendVerificationCode(ctx context.Context, to domain.Email, code domain.VerificationCode) error {
	if !s.revealCodes {
		s.logger.WarnContext(ctx, "verification email dropped: no email provider configured")
		return nil
	}
	s.logger.InfoContext(ctx, "dev email: verification code",
		slog.String("to_email", to.String()), slog.String("verification_code", code.Reveal()))
	return nil
}

// SendSignUpAttemptNotice implements usecase.EmailSender.
func (s *LogSender) SendSignUpAttemptNotice(ctx context.Context, to domain.Email) error {
	if !s.revealCodes {
		s.logger.WarnContext(ctx, "sign-up attempt notice dropped: no email provider configured")
		return nil
	}
	s.logger.InfoContext(ctx, "dev email: sign-up attempt notice", slog.String("to_email", to.String()))
	return nil
}
