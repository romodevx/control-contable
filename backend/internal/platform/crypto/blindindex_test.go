package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Obviously-fake 32-byte test peppers. The expected digests were computed
// independently with:
//
//	printf 'GOMA800101AB3' | openssl dgst -sha256 -hmac 'test-pepper-rfc-0123456789abcdef'
//	printf 'ana@example.com' | openssl dgst -sha256 -hmac 'test-pepper-email-0123456789abcd'
const (
	testRFCPepper   = "test-pepper-rfc-0123456789abcdef"
	testEmailPepper = "test-pepper-email-0123456789abcd"
	wantRFCHash     = "75df12795de51db1abb755397831e7d9c2ee7bb60434946fa7e3ba33483b0da4"
	wantEmailHash   = "27621bb2e43220cfa74130817ea014202faa2715438efbed6d533dbc60729bd7"
)

func TestNewBlindIndexRejectsShortPepper(t *testing.T) {
	short := []byte(testRFCPepper[:MinPepperSize-1])

	for name, build := range map[string]func([]byte) (*BlindIndex, error){
		"rfc": NewRFCIndex, "email": NewEmailIndex,
	} {
		t.Run(name, func(t *testing.T) {
			idx, err := build(short)

			require.ErrorIs(t, err, ErrInvalidPepper)
			assert.Nil(t, idx)
		})
	}
}

func TestRFCIndex(t *testing.T) {
	idx, err := NewRFCIndex([]byte(testRFCPepper))
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   string
		wantHex string
		wantErr error
	}{
		{"canonical", "GOMA800101AB3", wantRFCHash, nil},
		{"lowercase", "goma800101ab3", wantRFCHash, nil},
		{"mixed case with spaces", "  GoMa800101aB3\t\n", wantRFCHash, nil},
		{"different rfc", "GOMA800101AB4", "", nil},
		{"empty", "", "", ErrEmptyInput},
		{"only whitespace", " \t\n", "", ErrEmptyInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := idx.Hash(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, HashSize)
			if tt.wantHex == "" {
				assert.NotEqual(t, wantRFCHash, hex.EncodeToString(got))
				return
			}
			assert.Equal(t, tt.wantHex, hex.EncodeToString(got))
		})
	}
}

func TestRFCIndexUppercasesNonASCII(t *testing.T) {
	idx, err := NewRFCIndex([]byte(testRFCPepper))
	require.NoError(t, err)

	lower, err := idx.Hash("ñaño800101ab3")
	require.NoError(t, err)
	upper, err := idx.Hash("ÑAÑO800101AB3")
	require.NoError(t, err)

	assert.Equal(t, upper, lower)
}

func TestEmailIndex(t *testing.T) {
	idx, err := NewEmailIndex([]byte(testEmailPepper))
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   string
		wantHex string
		wantErr error
	}{
		{"canonical", "ana@example.com", wantEmailHash, nil},
		{"uppercase", "ANA@EXAMPLE.COM", wantEmailHash, nil},
		{"surrounding whitespace", " Ana@Example.com \n", wantEmailHash, nil},
		{"empty", "", "", ErrEmptyInput},
		{"only whitespace", "   ", "", ErrEmptyInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := idx.Hash(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantHex, hex.EncodeToString(got))
		})
	}
}

func TestBlindIndexesAreDomainSeparatedByPepper(t *testing.T) {
	rfc, err := NewRFCIndex([]byte(testRFCPepper))
	require.NoError(t, err)
	otherPepper, err := NewRFCIndex([]byte(testEmailPepper))
	require.NoError(t, err)

	a, err := rfc.Hash("GOMA800101AB3")
	require.NoError(t, err)
	b, err := otherPepper.Hash("GOMA800101AB3")
	require.NoError(t, err)

	assert.NotEqual(t, a, b)
}

func TestBlindIndexCopiesPepper(t *testing.T) {
	pepper := []byte(testRFCPepper)
	idx, err := NewRFCIndex(pepper)
	require.NoError(t, err)

	pepper[0] ^= 0xFF

	got, err := idx.Hash("GOMA800101AB3")
	require.NoError(t, err)
	assert.Equal(t, wantRFCHash, hex.EncodeToString(got))
}

// Regression (PR #7 follow-up 4): a dereferenced BlindIndex printed its pepper.
func TestBlindIndexValueNeverPrintsPepper(t *testing.T) {
	ptr, err := NewEmailIndex([]byte(testEmailPepper))
	require.NoError(t, err)
	idx := *ptr
	var logged bytes.Buffer
	slog.New(slog.NewTextHandler(&logged, nil)).Info("boot", "index", idx)

	for _, rendered := range []string{
		fmt.Sprintf("%v", idx), fmt.Sprintf("%+v", idx), fmt.Sprintf("%#v", idx), logged.String(),
	} {
		assert.NotContains(t, rendered, "test-pepper")
		assert.NotContains(t, rendered, "116 101 115 116") // "test" as %v bytes
		assert.Contains(t, rendered, "REDACTED")
	}
}

func TestBlindIndexNeverPrintsPepper(t *testing.T) {
	idx, err := NewEmailIndex([]byte(testEmailPepper))
	require.NoError(t, err)
	var logged bytes.Buffer
	slog.New(slog.NewTextHandler(&logged, nil)).Info("boot", "index", idx)

	for _, rendered := range []string{
		fmt.Sprintf("%v", idx), fmt.Sprintf("%+v", idx), fmt.Sprintf("%#v", idx), logged.String(),
	} {
		assert.NotContains(t, rendered, "test-pepper")
		assert.NotContains(t, rendered, "116 101 115 116") // "test" as %v bytes
		assert.Contains(t, rendered, "REDACTED")
	}
}

// Regression: strings.ToLower/ToUpper apply Unicode case mappings that fold
// distinct non-ASCII runes onto ASCII ones (KELVIN SIGN -> "k", LATIN SMALL
// LETTER LONG S -> "S", DOTLESS I -> "I"). A blind index built on them lets an
// attacker-chosen value collide with a victim's ASCII email or RFC.
func TestEmailIndexDoesNotFoldUnicodeOntoASCII(t *testing.T) {
	idx, err := NewEmailIndex([]byte(testEmailPepper))
	require.NoError(t, err)
	victim, err := idx.Hash("kevin@example.com")
	require.NoError(t, err)

	for name, lookalike := range map[string]string{
		"kelvin sign":  "Kevin@example.com",
		"dotted I":     "kevİn@example.com",
		"kelvin upper": "KEVIN@EXAMPLE.COM",
	} {
		t.Run(name, func(t *testing.T) {
			got, err := idx.Hash(lookalike)

			require.NoError(t, err)
			assert.NotEqual(t, victim, got)
		})
	}
}

func TestRFCIndexRejectsValuesThatAreNotAnRFC(t *testing.T) {
	idx, err := NewRFCIndex([]byte(testRFCPepper))
	require.NoError(t, err)

	for name, input := range map[string]string{
		"long s folds to S":    "GOMſ800101AB3",
		"dotless i folds to I": "GOMı800101AB3",
		"decomposed N tilde":   "ÑAN800101AB3",
		"internal space":       "GOMA 800101AB3",
		"too short":            "GOM80010AB3",
		"too long":             "GOMAX800101AB3",
		"letter in date":       "GOMA8001O1AB3",
		"digit in name":        "G0MA800101AB3",
		"symbol in homoclave":  "GOMA800101A-3",
	} {
		t.Run(name, func(t *testing.T) {
			got, err := idx.Hash(input)

			require.ErrorIs(t, err, ErrInvalidRFC)
			assert.Nil(t, got)
			assert.NotContains(t, err.Error(), input, "errors never carry the RFC")
		})
	}
}

func TestRFCIndexAcceptsBothRFCShapes(t *testing.T) {
	idx, err := NewRFCIndex([]byte(testRFCPepper))
	require.NoError(t, err)

	for name, input := range map[string]string{
		"persona fisica, 13 chars": "GOMA800101AB3",
		"persona moral, 12 chars":  "ABC800101AB3",
		"ampersand":                "A&C800101AB3",
		"enie":                     "ÑAÑO800101AB3",
		"lowercase enie":           "ñaño800101ab3",
		"generic publico general":  "XAXX010101000",
	} {
		t.Run(name, func(t *testing.T) {
			got, err := idx.Hash(input)

			require.NoError(t, err)
			assert.Len(t, got, HashSize)
		})
	}
}
