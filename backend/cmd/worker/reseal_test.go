package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

// testSealingKey is a sealing key on id 2 with a retired key on id 1.
func testSealingKey(t *testing.T) config.SealingKey {
	t.Helper()
	return config.SealingKey{
		Key: config.NewSecret(randomBytes(t)), KeyID: 2,
		Previous: config.PreviousKey{Key: config.NewSecret(randomBytes(t)), KeyID: 1},
	}
}

// heldColumns builds the jobs and returns the columns the key_reseal job
// said it holds.
func heldColumns(t *testing.T, w jobWiring) []string {
	t.Helper()
	var logs bytes.Buffer
	w.logger = logging.New(&logs, nil)
	_, err := newJobs(w)
	require.NoError(t, err)
	for line := range strings.Lines(logs.String()) {
		var entry struct {
			Msg     string `json:"msg"`
			Columns string `json:"columns"`
		}
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		if entry.Msg == "key reseal columns" {
			if entry.Columns == "" {
				return nil
			}
			return strings.Split(entry.Columns, ",")
		}
	}
	t.Fatalf("no key reseal columns line in %s", logs.String())
	return nil
}

// ADR-0012: the worker re-seals a column only when it holds that column's
// keys; it always reports readiness (key_usage) for every column.
func TestNewJobs_keyResealHoldsTheConfiguredColumns(t *testing.T) {
	t.Run("every key", func(t *testing.T) {
		w := testWiring(t, testReminderKeys(t))
		w.resealKeys = config.ResealKeys{RFC: testSealingKey(t), Email: testSealingKey(t)}

		assert.Equal(t, []string{"users.email_encrypted", "taxpayers.rfc_encrypted", "devices.token_ciphertext"},
			heldColumns(t, w))
	})
	t.Run("push token keys only", func(t *testing.T) {
		assert.Equal(t, []string{"devices.token_ciphertext"}, heldColumns(t, testWiring(t, testReminderKeys(t))))
	})
	t.Run("no key: readiness only", func(t *testing.T) {
		assert.Empty(t, heldColumns(t, testWiring(t, config.Reminders{})))
	})
}

func TestNewJobs_refusesABrokenResealKey(t *testing.T) {
	w := testWiring(t, config.Reminders{})
	w.resealKeys = config.ResealKeys{RFC: config.SealingKey{Key: config.NewSecret("short"), KeyID: 1}}

	_, err := newJobs(w)

	require.Error(t, err)
	assert.ErrorContains(t, err, "rfc keyring")
}
