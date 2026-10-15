package main

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	reminders "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	taxdomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// randomHex is a 32-byte key, hex, generated per call: never a literal.
func randomHex(t *testing.T) string {
	t.Helper()
	b := make([]byte, crypto.KeySize)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return hex.EncodeToString(b)
}

func randomBytes(t *testing.T) string {
	t.Helper()
	b := make([]byte, crypto.KeySize)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return string(b)
}

func testReminderKeys(t *testing.T) config.Reminders {
	t.Helper()
	return config.Reminders{
		PushTokenKey: config.NewSecret(randomBytes(t)), PushTokenKeyID: 1,
		PushTokenIndexPepper: config.NewSecret(randomBytes(t)),
	}
}

// testOutboxKeys is a random fixture outbox key, never a real one.
func testOutboxKeys(t *testing.T) config.AuthOutbox {
	t.Helper()
	return config.AuthOutbox{Key: config.NewSecret(randomBytes(t)), KeyID: 1}
}

// The collapse-id key is derived from the index pepper under its own
// label: stable for a pepper, never the pepper itself.
func TestCollapseKey(t *testing.T) {
	pepper := []byte(randomBytes(t))

	key := collapseKey(pepper)

	assert.Len(t, key, reminders.MinCollapseKeyBytes)
	assert.Equal(t, key, collapseKey(pepper))
	assert.NotEqual(t, pepper, key)
	assert.NotEqual(t, key, collapseKey([]byte(randomBytes(t))))
}

func regimeResult(status taxdomain.Status) taxusecase.RegimeResult {
	return taxusecase.RegimeResult{Status: status}
}

// MODULE-reminders §2.5: the estimate's taxpayer-level status, as the
// dashboard derives it, decides what a reminder shows.
func TestEstimateSummary(t *testing.T) {
	tests := []struct {
		name     string
		statuses []taxdomain.Status
		want     reminders.EstimateStatus
	}{
		{"every regime complete", []taxdomain.Status{taxdomain.StatusComplete, taxdomain.StatusComplete}, reminders.EstimateComplete},
		{"complete next to informational", []taxdomain.Status{taxdomain.StatusComplete, taxdomain.StatusInformational}, reminders.EstimateComplete},
		{"any incomplete", []taxdomain.Status{taxdomain.StatusComplete, taxdomain.StatusIncomplete, taxdomain.StatusPartial}, reminders.EstimateIncomplete},
		{"partial", []taxdomain.Status{taxdomain.StatusComplete, taxdomain.StatusPartial}, reminders.EstimateOther},
		{"informational only", []taxdomain.Status{taxdomain.StatusInformational}, reminders.EstimateOther},
		{"no regime", nil, reminders.EstimateOther},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := taxusecase.TaxpayerEstimate{Totals: taxdomain.Totals{TotalPayable: 4_520_00}}
			for _, s := range tc.statuses {
				e.Regimes = append(e.Regimes, regimeResult(s))
			}

			got := estimateSummary(e)

			assert.Equal(t, reminders.EstimateSummary{Status: tc.want, TotalPayableCents: 4_520_00}, got)
		})
	}
}
