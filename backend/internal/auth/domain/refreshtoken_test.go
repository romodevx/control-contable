package domain_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

const (
	userIDText = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"
	// secretOf0x01 is base64url(32 × 0x01) without padding.
	secretOf0x01 = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE" //nolint:gosec // G101: fixed test fixture, not a live token
)

func mustUserID(t *testing.T, text string) domain.UserID {
	t.Helper()
	id, err := domain.ParseUserID(text)
	require.NoError(t, err)
	return id
}

func TestParseUserID(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "canonical", raw: userIDText, want: userIDText},
		{name: "upper case canonicalized", raw: strings.ToUpper(userIDText), want: userIDText},
		{name: "nil uuid", raw: "00000000-0000-0000-0000-000000000000", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "no hyphens", raw: strings.ReplaceAll(userIDText, "-", ""), wantErr: true},
		{name: "braced", raw: "{" + userIDText + "}", wantErr: true},
		{name: "urn form", raw: "urn:uuid:" + userIDText, wantErr: true},
		{name: "hyphen misplaced", raw: "0199d4a27-c1e-7b3a-9f10-2c5e8a1b4d01", wantErr: true},
		{name: "non-hex digit", raw: "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d0g", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := domain.ParseUserID(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidUserID)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, id.String())
			assert.Equal(t, uuid.MustParse(tt.want), uuid.UUID(id), "UserID converts to uuid.UUID losslessly")
		})
	}
}

func TestNewRefreshTokenFormatsUserIDDotSecret(t *testing.T) {
	entropy := bytes.NewReader(bytes.Repeat([]byte{0x01}, domain.RefreshSecretBytes))
	token, err := domain.NewRefreshToken(mustUserID(t, userIDText), entropy)
	require.NoError(t, err)
	assert.Equal(t, userIDText+"."+secretOf0x01, token.Reveal())
	assert.Equal(t, userIDText, token.UserID().String())
}

func TestNewRefreshTokenFromCryptoRandRoundTrips(t *testing.T) {
	userID := mustUserID(t, userIDText)
	first, err := domain.NewRefreshToken(userID, rand.Reader)
	require.NoError(t, err)
	second, err := domain.NewRefreshToken(userID, rand.Reader)
	require.NoError(t, err)
	assert.NotEqual(t, first.Reveal(), second.Reveal())

	parsed, err := domain.ParseRefreshToken(first.Reveal())
	require.NoError(t, err)
	assert.Equal(t, first.Reveal(), parsed.Reveal())
	assert.True(t, first.Hash().Equal(parsed.Hash()))
	assert.Len(t, first.Reveal(), 80, "fits the contract's refresh_token length 40..200")
}

func TestNewRefreshTokenErrors(t *testing.T) {
	t.Run("zero user id", func(t *testing.T) {
		_, err := domain.NewRefreshToken(domain.UserID{}, rand.Reader)
		require.ErrorIs(t, err, domain.ErrInvalidUserID)
	})
	t.Run("short entropy", func(t *testing.T) {
		_, err := domain.NewRefreshToken(mustUserID(t, userIDText), bytes.NewReader(make([]byte, domain.RefreshSecretBytes-1)))
		require.ErrorIs(t, err, domain.ErrRandomSource)
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})
	t.Run("failing entropy", func(t *testing.T) {
		boom := errors.New("boom")
		_, err := domain.NewRefreshToken(mustUserID(t, userIDText), failingReader{err: boom})
		require.ErrorIs(t, err, domain.ErrRandomSource)
		require.ErrorIs(t, err, boom)
	})
}

func TestParseRefreshTokenCanonicalizesTheUserID(t *testing.T) {
	token, err := domain.ParseRefreshToken(strings.ToUpper(userIDText) + "." + secretOf0x01)
	require.NoError(t, err)
	assert.Equal(t, userIDText+"."+secretOf0x01, token.Reveal())
}

func TestParseRefreshTokenRejects(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "no separator", raw: userIDText + secretOf0x01},
		{name: "empty secret", raw: userIDText + "."},
		{name: "empty user id", raw: "." + secretOf0x01},
		{name: "two separators", raw: userIDText + "." + secretOf0x01 + ".x"},
		{name: "bad user id", raw: "not-a-uuid." + secretOf0x01},
		{name: "nil user id", raw: "00000000-0000-0000-0000-000000000000." + secretOf0x01},
		{name: "secret one byte short", raw: userIDText + "." + secretOf0x01[:42]},
		{name: "secret too long", raw: userIDText + "." + secretOf0x01 + "AA"},
		{name: "secret with padding", raw: userIDText + "." + secretOf0x01 + "="},
		{name: "secret std alphabet", raw: userIDText + "." + strings.Repeat("+", 43)},
		{name: "secret non-canonical trailing bits", raw: userIDText + "." + secretOf0x01[:42] + "F"},
		{name: "surrounding whitespace", raw: " " + userIDText + "." + secretOf0x01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.ParseRefreshToken(tt.raw)
			require.ErrorIs(t, err, domain.ErrMalformedRefreshToken)
		})
	}
}

// TestRefreshTokenHashIsSHA256OfTheCanonicalToken pins what refresh_tokens.token_hash
// stores: SHA-256 of the whole canonical "<uuid>.<secret>" string, so the
// hash also binds the user id.
func TestRefreshTokenHashIsSHA256OfTheCanonicalToken(t *testing.T) {
	token, err := domain.ParseRefreshToken(userIDText + "." + secretOf0x01)
	require.NoError(t, err)
	want := sha256.Sum256([]byte(userIDText + "." + secretOf0x01))
	assert.Equal(t, want[:], token.Hash().Bytes())
}

func TestTokenHashEqual(t *testing.T) {
	a, err := domain.ParseRefreshToken(userIDText + "." + secretOf0x01)
	require.NoError(t, err)
	b, err := domain.NewRefreshToken(mustUserID(t, userIDText), bytes.NewReader(bytes.Repeat([]byte{0x02}, domain.RefreshSecretBytes)))
	require.NoError(t, err)

	assert.True(t, a.Hash().Equal(a.Hash()))
	assert.False(t, a.Hash().Equal(b.Hash()))

	stored, err := domain.TokenHashFromBytes(a.Hash().Bytes())
	require.NoError(t, err)
	assert.True(t, stored.Equal(a.Hash()))
}

func TestTokenHashFromBytesRejectsWrongLength(t *testing.T) {
	for _, n := range []int{0, sha256.Size - 1, sha256.Size + 1} {
		_, err := domain.TokenHashFromBytes(make([]byte, n))
		require.ErrorIs(t, err, domain.ErrMalformedTokenHash, n)
	}
}

func TestTokenHashBytesIsACopy(t *testing.T) {
	token, err := domain.ParseRefreshToken(userIDText + "." + secretOf0x01)
	require.NoError(t, err)
	hash := token.Hash()
	raw := hash.Bytes()
	raw[0] ^= 0xFF
	assert.True(t, hash.Equal(token.Hash()))
}

func TestRefreshTokenNeverPrintsTheSecret(t *testing.T) {
	token, err := domain.ParseRefreshToken(userIDText + "." + secretOf0x01)
	require.NoError(t, err)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		assert.NotContains(t, fmt.Sprintf(verb, token), secretOf0x01[:8], verb)
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("x", "token", token)
	assert.NotContains(t, buf.String(), secretOf0x01[:8])
}

func TestParseRefreshTokenErrorNeverCarriesTheInput(t *testing.T) {
	raw := userIDText + "." + secretOf0x01[:42]
	_, err := domain.ParseRefreshToken(raw)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), secretOf0x01[:8])
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }
