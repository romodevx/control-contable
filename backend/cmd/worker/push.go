package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/fcm"
	reminders "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// newPushSender builds the reminders' push sender (M9-07) from cfg. Without a
// Firebase project it is the disabled sender: the worker still runs (the
// records purge does not depend on push) and the reminder job skips every
// run without claiming, so no reminder is consumed. With a project, credentials that cannot be used stop the
// worker here (fail closed). Whether module_reminders is ON is a runtime
// flag the job reads on every run, not a startup decision.
func newPushSender(ctx context.Context, cfg config.PushConfig, logger *slog.Logger) (reminders.PushSender, error) {
	if !cfg.Enabled() {
		logger.Warn("push disabled", slog.String("reason", "FCM_PROJECT_ID is not set; reminders are not sent"))
		return reminders.DisabledPushSender{}, nil
	}
	sender, err := fcm.New(ctx, fcm.Config{
		ProjectID:       cfg.ProjectID,
		CredentialsJSON: cfg.Credentials.RevealBytes(),
	})
	if err != nil {
		return nil, fmt.Errorf("refusing to start: push sender: %w", err)
	}
	logger.Info("push enabled",
		slog.String("project", cfg.ProjectID),
		slog.Bool("service_account_key", !cfg.Credentials.IsZero()))
	return sender, nil
}
