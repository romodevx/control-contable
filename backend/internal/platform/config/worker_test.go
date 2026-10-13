package config_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

// The worker serves no HTTP, so it reads neither PORT, the rate limit nor
// TRUSTED_PROXY_HOPS: a staging worker without TRUSTED_PROXY_HOPS is valid.
func TestLoadWorker(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want config.WorkerConfig
	}{
		{
			name: "defaults",
			env:  map[string]string{"APP_ENV": "dev"},
			want: config.WorkerConfig{Env: config.EnvDev, LogLevel: slog.LevelInfo},
		},
		{
			name: "staging needs no HTTP settings",
			env:  map[string]string{"APP_ENV": "staging", "LOG_LEVEL": "debug", "PORT": "not-read"},
			want: config.WorkerConfig{Env: config.EnvStaging, LogLevel: slog.LevelDebug},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.LoadWorker(context.Background(), lookupFrom(tt.env), stubSecrets{})

			require.NoError(t, err)
			assert.Equal(t, dbURL, got.DatabaseURL.Reveal())
			got.DatabaseURL = config.Secret{}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadWorker_errors(t *testing.T) {
	errBackend := errors.New("secret manager: permission denied")

	_, err := config.LoadWorker(context.Background(),
		lookupFrom(map[string]string{"LOG_LEVEL": "loud"}),
		stubSecrets{"DATABASE_URL": errBackend})

	require.Error(t, err)
	for _, want := range []string{"APP_ENV is required", `LOG_LEVEL "loud"`, "secret DATABASE_URL"} {
		assert.ErrorContains(t, err, want)
	}
	assert.ErrorIs(t, err, errBackend)
	assert.NotContains(t, err.Error(), "s3cr3t")
}
