package config_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

const dbURL = "fixture-database-url-s3cr3t" // opaque: tests only check it is never printed

func lookupFrom(env map[string]string) config.Lookup {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

type stubSecrets map[string]error

func (s stubSecrets) Secret(_ context.Context, name string) (config.Secret, error) {
	if err, ok := s[name]; ok {
		return config.Secret{}, err
	}
	return config.NewSecret(dbURL), nil
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		secrets config.SecretSource
		want    config.Config
	}{
		{
			name:    "defaults",
			env:     map[string]string{"APP_ENV": "dev", "DATABASE_URL": dbURL},
			secrets: nil,
			want: config.Config{
				Env: config.EnvDev, HTTPAddr: ":8080", LogLevel: slog.LevelInfo,
				RateLimit: config.RateLimit{PerSecond: 10, Burst: 20}, TrustedProxyHops: 0,
			},
		},
		{
			name: "overrides",
			env: map[string]string{
				"APP_ENV": "production", "PORT": "9090", "LOG_LEVEL": "debug",
				"RATE_LIMIT_RPS": "5", "RATE_LIMIT_BURST": "7", "TRUSTED_PROXY_HOPS": "1",
			},
			secrets: stubSecrets{},
			want: config.Config{
				Env: config.EnvProduction, HTTPAddr: ":9090", LogLevel: slog.LevelDebug,
				RateLimit: config.RateLimit{PerSecond: 5, Burst: 7}, TrustedProxyHops: 1,
			},
		},
		{
			name:    "staging with warn level",
			env:     map[string]string{"APP_ENV": "staging", "LOG_LEVEL": "WARN", "TRUSTED_PROXY_HOPS": "0"},
			secrets: stubSecrets{},
			want: config.Config{
				Env: config.EnvStaging, HTTPAddr: ":8080", LogLevel: slog.LevelWarn,
				RateLimit: config.RateLimit{PerSecond: 10, Burst: 20},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := lookupFrom(tt.env)
			secrets := tt.secrets
			if secrets == nil {
				secrets = config.NewEnvSecrets(lookup)
			}

			got, err := config.Load(context.Background(), lookup, secrets)

			require.NoError(t, err)
			assert.Equal(t, dbURL, got.DatabaseURL.Reveal())
			got.DatabaseURL = config.Secret{}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoad_errors(t *testing.T) {
	errBackend := errors.New("secret manager: permission denied")
	tests := []struct {
		name     string
		env      map[string]string
		secrets  stubSecrets
		wantErrs []string
		wantIs   error
	}{
		{
			name:     "missing APP_ENV",
			env:      map[string]string{},
			wantErrs: []string{"APP_ENV is required"},
		},
		{
			name:     "unknown APP_ENV",
			env:      map[string]string{"APP_ENV": "qa"},
			wantErrs: []string{`APP_ENV "qa" is not one of dev, staging, production`},
		},
		{
			name:     "bad LOG_LEVEL",
			env:      map[string]string{"APP_ENV": "dev", "LOG_LEVEL": "loud"},
			wantErrs: []string{`LOG_LEVEL "loud"`},
		},
		{
			name:     "non-numeric PORT",
			env:      map[string]string{"APP_ENV": "dev", "PORT": "http"},
			wantErrs: []string{`PORT "http" is not a TCP port`},
		},
		{
			name:     "PORT out of range",
			env:      map[string]string{"APP_ENV": "dev", "PORT": "70000"},
			wantErrs: []string{`PORT "70000" is not a TCP port`},
		},
		{
			name:     "non-numeric RATE_LIMIT_RPS",
			env:      map[string]string{"APP_ENV": "dev", "RATE_LIMIT_RPS": "fast"},
			wantErrs: []string{`RATE_LIMIT_RPS "fast" is not an integer in [1, 10000]`},
		},
		{
			name:     "zero RATE_LIMIT_RPS",
			env:      map[string]string{"APP_ENV": "dev", "RATE_LIMIT_RPS": "0"},
			wantErrs: []string{`RATE_LIMIT_RPS "0" is not an integer in [1, 10000]`},
		},
		{
			name:     "RATE_LIMIT_BURST out of range",
			env:      map[string]string{"APP_ENV": "dev", "RATE_LIMIT_BURST": "10001"},
			wantErrs: []string{`RATE_LIMIT_BURST "10001" is not an integer in [1, 10000]`},
		},
		{
			name:     "negative TRUSTED_PROXY_HOPS",
			env:      map[string]string{"APP_ENV": "dev", "TRUSTED_PROXY_HOPS": "-1"},
			wantErrs: []string{`TRUSTED_PROXY_HOPS "-1" is not an integer in [0, 10]`},
		},
		{
			// Behind a load balancer every request shares the proxy's peer
			// address, so a silent default of 0 would put all clients in one
			// rate-limit bucket. Outside dev the hop count must be explicit.
			name:     "TRUSTED_PROXY_HOPS unset in staging",
			env:      map[string]string{"APP_ENV": "staging"},
			wantErrs: []string{"TRUSTED_PROXY_HOPS is required when APP_ENV is staging"},
		},
		{
			name:     "TRUSTED_PROXY_HOPS empty in production",
			env:      map[string]string{"APP_ENV": "production", "TRUSTED_PROXY_HOPS": ""},
			wantErrs: []string{"TRUSTED_PROXY_HOPS is required when APP_ENV is production"},
		},
		{
			name:     "missing secret",
			env:      map[string]string{"APP_ENV": "dev"},
			secrets:  stubSecrets{"DATABASE_URL": config.ErrSecretNotFound},
			wantErrs: []string{"secret DATABASE_URL"},
			wantIs:   config.ErrSecretNotFound,
		},
		{
			name:     "secret backend failure is wrapped",
			env:      map[string]string{"APP_ENV": "dev"},
			secrets:  stubSecrets{"DATABASE_URL": errBackend},
			wantErrs: []string{"secret DATABASE_URL"},
			wantIs:   errBackend,
		},
		{
			name:     "every problem reported at once",
			env:      map[string]string{"PORT": "0", "LOG_LEVEL": "x"},
			secrets:  stubSecrets{"DATABASE_URL": config.ErrSecretNotFound},
			wantErrs: []string{"APP_ENV is required", "PORT", "LOG_LEVEL", "secret DATABASE_URL"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secrets := tt.secrets
			if secrets == nil {
				secrets = stubSecrets{}
			}

			_, err := config.Load(context.Background(), lookupFrom(tt.env), secrets)

			require.Error(t, err)
			for _, want := range tt.wantErrs {
				assert.ErrorContains(t, err, want)
			}
			if tt.wantIs != nil {
				assert.ErrorIs(t, err, tt.wantIs)
			}
			assert.NotContains(t, err.Error(), "s3cr3t")
		})
	}
}

func TestEnvSecrets(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    string
		wantErr error
	}{
		{name: "present", env: map[string]string{"DATABASE_URL": dbURL}, want: dbURL},
		{name: "absent", env: map[string]string{}, wantErr: config.ErrSecretNotFound},
		{name: "empty counts as absent", env: map[string]string{"DATABASE_URL": ""}, wantErr: config.ErrSecretNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.NewEnvSecrets(lookupFrom(tt.env)).Secret(context.Background(), "DATABASE_URL")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Reveal())
		})
	}
}

// TestSecret_neverPrinted covers every way a value usually escapes into logs.
func TestSecret_neverPrinted(t *testing.T) {
	s := config.NewSecret(dbURL)
	jsonBytes, err := json.Marshal(struct{ URL config.Secret }{s})
	require.NoError(t, err)
	var logBuf bytes.Buffer
	slog.New(slog.NewJSONHandler(&logBuf, nil)).Info("loaded", "db", s)

	rendered := map[string]string{
		"%v":   fmt.Sprintf("%v", s),
		"%s":   fmt.Sprintf("%s", s),
		"%+v":  fmt.Sprintf("%+v", s),
		"%#v":  fmt.Sprintf("%#v", s),
		"%x":   fmt.Sprintf("%x", s),
		"str":  s.String(),
		"json": string(jsonBytes),
		"slog": logBuf.String(),
	}
	for name, out := range rendered {
		t.Run(name, func(t *testing.T) {
			assert.NotContains(t, out, "s3cr3t")
			assert.Contains(t, out, config.Redacted)
		})
	}
}

func TestSecret_IsZero(t *testing.T) {
	assert.True(t, config.Secret{}.IsZero())
	assert.False(t, config.NewSecret("x").IsZero())
}
