package infra_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

func email(t *testing.T) domain.Email {
	t.Helper()
	e, err := domain.NewEmail("ana.gomez@example.mx")
	require.NoError(t, err)
	return e
}

// sendAll exercises both emails through a LogSender for env and returns the
// log output (the production redacting logger).
func sendAll(t *testing.T, env config.Environment) string {
	t.Helper()
	var buf bytes.Buffer
	sender := infra.NewLogSender(logging.New(&buf, nil), env)
	require.NoError(t, sender.SendVerificationCode(t.Context(), email(t), code(t, "482913")))
	require.NoError(t, sender.SendSignUpAttemptNotice(t.Context(), email(t)))
	return buf.String()
}

// The dev adapter is how a developer reads the code locally (MODULE-auth §6.7).
func TestLogSender_revealsTheCodeInDev(t *testing.T) {
	logs := sendAll(t, config.EnvDev)

	assert.Contains(t, logs, `"verification_code":"482913"`)
	assert.Contains(t, logs, "sign-up attempt notice")
	assert.NotContains(t, logs, "ana.gomez@example.mx", "the address is redacted even in dev")
}

// Outside dev the code never reaches a log line, whatever the logger.
func TestLogSender_neverLogsTheCodeOutsideDev(t *testing.T) {
	for _, env := range []config.Environment{config.EnvStaging, config.EnvProduction, "", "qa"} {
		t.Run(string(env), func(t *testing.T) {
			logs := sendAll(t, env)

			assert.NotContains(t, logs, "482913")
			assert.NotContains(t, logs, "verification_code")
			assert.Contains(t, logs, "no email provider", "the drop is visible to operators")
			assert.NotContains(t, logs, "ana.gomez@example.mx")
		})
	}
}
