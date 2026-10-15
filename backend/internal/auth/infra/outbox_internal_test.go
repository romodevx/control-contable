package infra

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
)

func outboxKeyring(t *testing.T, fill byte) *crypto.Keyring {
	t.Helper()
	k, err := crypto.NewKeyring(1, map[crypto.KeyID][]byte{1: bytes.Repeat([]byte{fill}, crypto.KeySize)})
	require.NoError(t, err)
	return k
}

func sampleEmails(t *testing.T) []usecase.OutboundEmail {
	t.Helper()
	to, err := domain.NewEmail("ana.gomez@example.mx")
	require.NoError(t, err)
	code, err := domain.ParseVerificationCode("482913")
	require.NoError(t, err)
	return []usecase.OutboundEmail{
		{Kind: usecase.EmailVerificationCode, To: to, Code: code},
		{Kind: usecase.EmailSignUpAttemptNotice, To: to},
	}
}

func TestOutboxPayload_roundTripsSealed(t *testing.T) {
	box := NewOutbox(outboxKeyring(t, 0x31))
	for _, email := range sampleEmails(t) {
		id := uuid.New()

		sealed, err := box.seal(id, email)
		require.NoError(t, err)
		got, ok := box.open(id, sealed)

		require.True(t, ok)
		assert.Equal(t, email, got)
		assert.NotContains(t, string(sealed), "example.mx", "the address is sealed")
		assert.NotContains(t, string(sealed), "482913", "the code is sealed")
	}
}

// The database holds nothing that tells a code from a notice: both seal to the
// same length for the same address (the kind alone would reveal an account).
func TestOutboxPayload_kindDoesNotShowInTheLength(t *testing.T) {
	box := NewOutbox(outboxKeyring(t, 0x31))
	emails := sampleEmails(t)

	code, err := box.seal(uuid.New(), emails[0])
	require.NoError(t, err)
	notice, err := box.seal(uuid.New(), emails[1])
	require.NoError(t, err)

	assert.Len(t, notice, len(code))
}

// Every address domain.NewEmail accepts must seal, at the same padded length.
// RFC 5322 atext allows '&', which an HTML-escaping JSON encoder turns into six
// bytes: a 64-byte local part of '&' would overflow the pad and fail every
// sign-up for that address with a 500.
func TestOutboxPayload_sealsTheLongestValidAddress(t *testing.T) {
	box := NewOutbox(outboxKeyring(t, 0x31))
	host := strings.Repeat(strings.Repeat("a", 63)+".", 2) + strings.Repeat("b", 58) + ".mx"
	tests := map[string]string{
		"html-escapable local part": strings.Repeat("&", 64) + "@example.mx",
		"longest address":           strings.Repeat("&", 64) + "@" + host,
	}
	reference, err := box.seal(uuid.New(), sampleEmails(t)[0])
	require.NoError(t, err)
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			to, err := domain.NewEmail(raw)
			require.NoError(t, err, "the domain accepts it")
			email := usecase.OutboundEmail{Kind: usecase.EmailVerificationCode, To: to, Code: sampleEmails(t)[0].Code}
			id := uuid.New()

			sealed, err := box.seal(id, email)

			require.NoError(t, err)
			assert.Len(t, sealed, len(reference))
			got, ok := box.open(id, sealed)
			require.True(t, ok)
			assert.Equal(t, email, got)
		})
	}
}

// Anything this code did not seal for this row is unreadable: another row's
// payload (the id is the additional data), another key, tampered bytes.
func TestOutboxPayload_refusesWhatItDidNotSealForThisRow(t *testing.T) {
	box := NewOutbox(outboxKeyring(t, 0x31))
	id := uuid.New()
	sealed, err := box.seal(id, sampleEmails(t)[0])
	require.NoError(t, err)
	tampered := bytes.Clone(sealed)
	tampered[len(tampered)-1] ^= 1

	tests := map[string]struct {
		box     *Outbox
		id      uuid.UUID
		payload []byte
	}{
		"moved to another row": {box, uuid.New(), sealed},
		"another key":          {NewOutbox(outboxKeyring(t, 0x32)), id, sealed},
		"tampered":             {box, id, tampered},
		"garbage":              {box, id, []byte("not a ciphertext")},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, ok := tt.box.open(tt.id, tt.payload)
			assert.False(t, ok)
		})
	}
}

func TestOutboxPayload_refusesAPlaintextItDidNotWrite(t *testing.T) {
	box := NewOutbox(outboxKeyring(t, 0x31))
	id := uuid.New()
	for name, plain := range map[string]string{
		"unknown version": `{"v":9,"kind":"notice","to":"ana@example.mx","code":""}`,
		"unknown kind":    `{"v":1,"kind":"invoice","to":"ana@example.mx","code":""}`,
		"bad address":     `{"v":1,"kind":"notice","to":"ana@localhost","code":""}`,
		"bad code":        `{"v":1,"kind":"code","to":"ana@example.mx","code":"12ab56"}`,
		"code on notice":  `{"v":1,"kind":"notice","to":"ana@example.mx","code":"123456"}`,
		"unknown field":   `{"v":1,"kind":"notice","to":"ana@example.mx","code":"","x":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			sealed, err := box.keyring.Encrypt([]byte(plain), outboxAAD(id))
			require.NoError(t, err)
			_, ok := box.open(id, sealed)
			assert.False(t, ok)
		})
	}
}

func TestLeaseSeconds_roundsUpToAWholeSecond(t *testing.T) {
	assert.Equal(t, 120, wholeSeconds(usecase.DeliveryLease))
	assert.Equal(t, 1, wholeSeconds(1))
	assert.Equal(t, 31, wholeSeconds(30500e6))
}
