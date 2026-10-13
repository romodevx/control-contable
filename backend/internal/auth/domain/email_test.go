package domain_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
)

func TestNewEmailNormalizes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "already canonical", raw: "ana@example.mx", want: "ana@example.mx"},
		{name: "upper case folded to ASCII lower", raw: "Ana.Gomez@Example.MX", want: "ana.gomez@example.mx"},
		{name: "surrounding whitespace trimmed", raw: " \t ana@example.mx\n", want: "ana@example.mx"},
		{name: "atext specials in local part", raw: "a!#$%&'*+-/=?^_`{|}~b@example.mx", want: "a!#$%&'*+-/=?^_`{|}~b@example.mx"},
		{name: "dotted local part", raw: "a.b.c@example.mx", want: "a.b.c@example.mx"},
		{name: "subdomains and digits", raw: "x@mail-1.sub.example.com.mx", want: "x@mail-1.sub.example.com.mx"},
		{name: "64-octet local part", raw: strings.Repeat("a", 64) + "@example.mx", want: strings.Repeat("a", 64) + "@example.mx"},
		{name: "63-octet label", raw: "a@" + strings.Repeat("b", 63) + ".mx", want: "a@" + strings.Repeat("b", 63) + ".mx"},
		{name: "254-octet address", raw: maxLengthEmail(), want: maxLengthEmail()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := domain.NewEmail(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, email.String())
		})
	}
}

func TestNewEmailRejects(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "only whitespace", raw: "  \t "},
		{name: "no at sign", raw: "ana.example.mx"},
		{name: "two at signs", raw: "ana@b@example.mx"},
		{name: "empty local part", raw: "@example.mx"},
		{name: "empty domain", raw: "ana@"},
		{name: "local part leading dot", raw: ".ana@example.mx"},
		{name: "local part trailing dot", raw: "ana.@example.mx"},
		{name: "local part double dot", raw: "a..na@example.mx"},
		{name: "space inside local part", raw: "an a@example.mx"},
		{name: "quoted local part", raw: `"ana"@example.mx`},
		{name: "local part over 64 octets", raw: strings.Repeat("a", 65) + "@example.mx"},
		{name: "single-label domain", raw: "ana@localhost"},
		{name: "domain leading dot", raw: "ana@.example.mx"},
		{name: "domain trailing dot", raw: "ana@example.mx."},
		{name: "domain empty label", raw: "ana@example..mx"},
		{name: "label leading hyphen", raw: "ana@-example.mx"},
		{name: "label trailing hyphen", raw: "ana@example-.mx"},
		{name: "label underscore", raw: "ana@exa_mple.mx"},
		{name: "label over 63 octets", raw: "a@" + strings.Repeat("b", 64) + ".mx"},
		{name: "ip literal", raw: "ana@[127.0.0.1]"},
		{name: "address over 254 octets", raw: "a" + maxLengthEmail()},
		{name: "non-ASCII local part", raw: "añа@example.mx"},
		{name: "non-ASCII domain", raw: "ana@exámple.mx"},
		{name: "KELVIN SIGN would fold to k", raw: "\u212Aana@example.mx"},
		{name: "invalid UTF-8", raw: "an\xffa@example.mx"},
		{name: "control character", raw: "an\x00a@example.mx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewEmail(tt.raw)
			require.ErrorIs(t, err, domain.ErrInvalidEmail)
		})
	}
}

// TestEmailMatchesBlindIndexNormalization pins the contract with the
// email_hash blind index: hashing the raw input and hashing the domain's
// normalized form give the same index, so a lookup by Email.String() always
// finds the row written from the user's raw input.
func TestEmailMatchesBlindIndexNormalization(t *testing.T) {
	index, err := crypto.NewEmailIndex(bytes.Repeat([]byte{7}, crypto.MinPepperSize))
	require.NoError(t, err)
	for _, raw := range []string{"ana@example.mx", " Ana.Gomez@Example.MX\t", "X+tag@Sub.Example.COM"} {
		email, err := domain.NewEmail(raw)
		require.NoError(t, err)
		fromRaw, err := index.Hash(raw)
		require.NoError(t, err)
		fromEmail, err := index.Hash(email.String())
		require.NoError(t, err)
		assert.Equal(t, fromRaw, fromEmail, raw)
	}
}

func TestEmailNeverPrintsTheAddress(t *testing.T) {
	email, err := domain.NewEmail("ana@example.mx")
	require.NoError(t, err)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		assert.NotContains(t, fmt.Sprintf(verb, email), "ana", verb)
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("x", "email", email)
	assert.NotContains(t, buf.String(), "ana")
}

func TestEmailZeroValue(t *testing.T) {
	var email domain.Email
	assert.True(t, email.IsZero())
	parsed, err := domain.NewEmail("ana@example.mx")
	require.NoError(t, err)
	assert.False(t, parsed.IsZero())
}

// maxLengthEmail is a valid 254-octet address (the RFC 5321 path limit
// minus the angle brackets): 64-octet local part, 189-octet domain.
func maxLengthEmail() string {
	domainPart := strings.Join([]string{
		strings.Repeat("c", 60), strings.Repeat("d", 61), strings.Repeat("e", 61), "abcd",
	}, ".") // 60 + 61 + 61 + 4 + 3 dots = 189 octets
	return strings.Repeat("l", 64) + "@" + domainPart
}
