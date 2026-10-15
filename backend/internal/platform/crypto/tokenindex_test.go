package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTokenPepper is an obviously-fake 32-byte pepper.
const testTokenPepper = "test-pepper-token-0123456789abcd"

// The token_hmac index (MODULE-reminders §2.3) is HMAC-SHA256 of the token
// exactly as sent: a push token is case-sensitive and opaque, so nothing is
// trimmed or folded.
func TestTokenIndex_isHMACOfTheTokenAsIs(t *testing.T) {
	idx, err := NewTokenIndex([]byte(testTokenPepper))
	require.NoError(t, err)
	token := rand.Text()
	mac := hmac.New(sha256.New, []byte(testTokenPepper))
	mac.Write([]byte(token))

	got, err := idx.Hash(token)

	require.NoError(t, err)
	assert.Equal(t, mac.Sum(nil), got)
	assert.Len(t, got, HashSize)
}

func TestTokenIndex_keepsCaseAndSpaces(t *testing.T) {
	idx, err := NewTokenIndex([]byte(testTokenPepper))
	require.NoError(t, err)
	token := rand.Text()

	upper, err := idx.Hash(token)
	require.NoError(t, err)
	lower, err := idx.Hash(strings.ToLower(token))
	require.NoError(t, err)
	padded, err := idx.Hash(" " + token)
	require.NoError(t, err)

	assert.NotEqual(t, upper, lower, "case is significant")
	assert.NotEqual(t, upper, padded, "nothing is trimmed")
}

func TestTokenIndex_refusesAnEmptyTokenAndAShortPepper(t *testing.T) {
	idx, err := NewTokenIndex([]byte(testTokenPepper))
	require.NoError(t, err)

	_, err = idx.Hash("")
	require.ErrorIs(t, err, ErrEmptyInput)

	_, err = NewTokenIndex([]byte(testTokenPepper[:MinPepperSize-1]))
	require.ErrorIs(t, err, ErrInvalidPepper)
}
