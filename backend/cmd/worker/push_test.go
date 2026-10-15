package main

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/fcm"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/fcm/fcmtest"
	reminders "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const testFCMProject = "cc-worker-test"

// A worker without FCM_PROJECT_ID starts with push disabled and says so; the
// reminder job then claims nothing until push is configured.
func TestNewPushSender_disabledWithoutProject(t *testing.T) {
	var logs bytes.Buffer

	sender, err := newPushSender(t.Context(), config.PushConfig{}, logging.New(&logs, nil))

	require.NoError(t, err)
	assert.Equal(t, reminders.DisabledPushSender{}, sender)
	assert.Contains(t, logs.String(), `"msg":"push disabled"`)
}

func TestNewPushSender_withServiceAccount(t *testing.T) {
	var logs bytes.Buffer
	creds := fcmtest.ServiceAccountJSON(t, testFCMProject, "http://127.0.0.1:1/token")
	cfg := config.PushConfig{ProjectID: testFCMProject, Credentials: config.NewSecret(string(creds))}

	sender, err := newPushSender(t.Context(), cfg, logging.New(&logs, nil))

	require.NoError(t, err)
	assert.IsType(t, &fcm.Sender{}, sender)
	assert.Contains(t, logs.String(), `"msg":"push enabled"`)
	assert.Contains(t, logs.String(), testFCMProject)
	assert.NotContains(t, logs.String(), "PRIVATE")
}

// Credentials that are present but unusable stop the worker at startup
// (fail closed), never at the first reminder.
func TestRun_refusesToStartWithUnusablePushCredentials(t *testing.T) {
	marker := uuid.NewString()
	env := map[string]string{
		"APP_ENV":              "dev",
		"DATABASE_URL":         "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
		"FCM_PROJECT_ID":       testFCMProject,
		"FCM_CREDENTIALS_JSON": `{"type":"service_account","private_key":"` + marker + `"}`,
		"AUTH_OUTBOX_KEY":      randomHex(t),
	}

	err := run(t.Context(), lookupFrom(env), &bytes.Buffer{})

	require.ErrorIs(t, err, fcm.ErrConfig)
	assert.ErrorContains(t, err, "refusing to start: push sender")
	assert.NotContains(t, err.Error(), marker)
}
