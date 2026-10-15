package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// WorkerConfig is the validated configuration of cmd/worker. The worker serves
// no HTTP, so it reads APP_ENV, LOG_LEVEL, the DATABASE_URL secret (logged
// in as cc_worker) and the optional push settings (FCM_PROJECT_ID and the
// FCM_CREDENTIALS_JSON secret, M9-07).
type WorkerConfig struct {
	Env         Environment
	LogLevel    slog.Level
	DatabaseURL Secret
	Push        PushConfig
}

// LoadWorker reads and validates the worker configuration. Like Load, the
// error joins every problem found and never contains a secret value.
func LoadWorker(ctx context.Context, lookup Lookup, secrets SecretSource) (WorkerConfig, error) {
	var (
		cfg  WorkerConfig
		errs []error
		err  error
	)
	if cfg.Env, err = parseEnvironment(lookup); err != nil {
		errs = append(errs, err)
	}
	if cfg.LogLevel, err = parseLogLevel(lookup); err != nil {
		errs = append(errs, err)
	}
	if cfg.DatabaseURL, err = secrets.Secret(ctx, secretDatabase); err != nil {
		errs = append(errs, fmt.Errorf("secret %s: %w", secretDatabase, err))
	}
	if cfg.Push, err = loadPush(ctx, lookup, secrets); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return WorkerConfig{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}
