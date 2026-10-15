package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

// Push is opt-in per environment: FCM_PROJECT_ID turns it on, and the
// service account key comes from the secret manager when present (otherwise
// the runtime's own credentials). Without the project the worker runs with
// push disabled.
func TestLoadWorker_push(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		secrets stubSecrets
		want    config.PushConfig
		creds   bool
	}{
		{
			name: "no project: push disabled, credentials not read",
			env:  map[string]string{"APP_ENV": "staging"},
			want: config.PushConfig{},
		},
		{
			name:  "project and key from the secret manager",
			env:   map[string]string{"APP_ENV": "staging", "FCM_PROJECT_ID": "control-contable-stg"},
			want:  config.PushConfig{ProjectID: "control-contable-stg"},
			creds: true,
		},
		{
			name:    "project without key: runtime credentials",
			env:     map[string]string{"APP_ENV": "production", "FCM_PROJECT_ID": "control-contable"},
			secrets: stubSecrets{"FCM_CREDENTIALS_JSON": config.ErrSecretNotFound},
			want:    config.PushConfig{ProjectID: "control-contable"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.LoadWorker(context.Background(), lookupFrom(tt.env), tt.secrets)

			require.NoError(t, err)
			assert.Equal(t, tt.creds, !got.Push.Credentials.IsZero())
			assert.Equal(t, tt.want.ProjectID, got.Push.ProjectID)
			assert.Equal(t, tt.want.ProjectID != "", got.Push.Enabled())
		})
	}
}

func TestLoadWorker_pushErrors(t *testing.T) {
	errBackend := errors.New("secret manager: permission denied")
	tests := []struct {
		name    string
		project string
		secrets stubSecrets
		want    string
	}{
		{"project id shape", "Not_A_Project", stubSecrets{}, `FCM_PROJECT_ID "Not_A_Project"`},
		{"project id too short", "abc", stubSecrets{}, `FCM_PROJECT_ID "abc"`},
		{"secret manager failure", "control-contable", stubSecrets{"FCM_CREDENTIALS_JSON": errBackend}, "secret FCM_CREDENTIALS_JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.LoadWorker(context.Background(),
				lookupFrom(map[string]string{"APP_ENV": "dev", "FCM_PROJECT_ID": tt.project}), tt.secrets)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.want)
			assert.NotContains(t, err.Error(), dbURL)
		})
	}
}
