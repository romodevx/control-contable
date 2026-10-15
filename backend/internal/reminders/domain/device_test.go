package domain_test

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

func TestParsePlatform(t *testing.T) {
	for raw, want := range map[string]domain.Platform{"ios": domain.PlatformIOS, "android": domain.PlatformAndroid} {
		got, err := domain.ParsePlatform(raw)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
	for _, raw := range []string{"", "iOS", "web", "windows"} {
		_, err := domain.ParsePlatform(raw)

		require.ErrorIs(t, err, domain.ErrInvalidPlatform, raw)
	}
}

// Contract DeviceCreate.push_token: 1-4096 printable ASCII, no spaces.
func TestNewPushToken(t *testing.T) {
	token := rand.Text() + ":" + rand.Text() + "_-~!"

	got, err := domain.NewPushToken(token)

	require.NoError(t, err)
	assert.Equal(t, token, got.Reveal())
}

func TestNewPushToken_refusesWhatTheContractForbids(t *testing.T) {
	for name, raw := range map[string]string{
		"empty":     "",
		"too long":  strings.Repeat("a", domain.MaxPushTokenLength+1),
		"a space":   rand.Text() + " " + rand.Text(),
		"a tab":     rand.Text() + "\t",
		"a newline": rand.Text() + "\n",
		"non-ASCII": rand.Text() + "ñ",
		"DEL":       rand.Text() + "\x7f",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := domain.NewPushToken(raw)

			require.ErrorIs(t, err, domain.ErrInvalidPushToken)
		})
	}
	_, err := domain.NewPushToken(strings.Repeat("a", domain.MaxPushTokenLength))
	require.NoError(t, err, "the maximum length is accepted")
}

// MODULE-reminders §2.3, §8: the token is never logged. Every fmt verb and
// slog render it redacted; only Reveal hands it to the sealer.
func TestPushToken_neverPrints(t *testing.T) {
	token := rand.Text()
	got, err := domain.NewPushToken(token)
	require.NoError(t, err)
	var logs strings.Builder
	logger := slog.New(slog.NewJSONHandler(&logs, nil))

	logger.Info("x", slog.Any("token", got), slog.Any("struct", struct{ T domain.PushToken }{got}))
	printed := fmt.Sprintf("%v %+v %#v %s %q", got, got, got, got, got)

	assert.NotContains(t, logs.String(), token)
	assert.NotContains(t, printed, token)
	assert.Contains(t, printed, "REDACTED")
}

func TestParseAppVersion(t *testing.T) {
	for _, raw := range []string{"1.2.0", "1.2.0+45", "2.0.0-beta.1", strings.Repeat("9", domain.MaxAppVersionLength)} {
		got, err := domain.ParseAppVersion(raw)

		require.NoError(t, err, raw)
		assert.Equal(t, raw, got)
	}
	for _, raw := range []string{"", "1.2 beta", "v1/2", "1.0_0", strings.Repeat("9", domain.MaxAppVersionLength+1)} {
		_, err := domain.ParseAppVersion(raw)

		require.ErrorIs(t, err, domain.ErrInvalidAppVersion, raw)
	}
}
