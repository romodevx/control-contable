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

const (
	// twelveCharSample is a made-up value that meets the policy; not a credential.
	twelveCharSample   = "tlacuache-87"
	combiningRingAbove = "\u030A"
	expandsUnderNFKC   = "\uFDFA"       // one code point, 33 bytes under NFKC
	fullwidthPassword  = "ｐａｓｓｗｏｒｄ１２３４" //nolint:gosec // G101: blocklisted test input; NFKC → "password1234"
)

func TestNewPasswordAccepts(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "exactly 12 characters", raw: twelveCharSample, want: twelveCharSample},
		{name: "no composition rules: lowercase only", raw: "mipatoazulvuela", want: "mipatoazulvuela"},
		{name: "spaces are kept", raw: "  mi pato azul vuela  ", want: "  mi pato azul vuela  "},
		{name: "unicode and emoji", raw: "contraseña🦆segura", want: "contraseña🦆segura"},
		{name: "length counted in characters after NFKC", raw: strings.Repeat("x", 11) + "A" + combiningRingAbove, want: strings.Repeat("x", 11) + "Å"},
		{name: "ligature decomposed by NFKC", raw: "ﬁesta-en-la-playa", want: "fiesta-en-la-playa"},
		{name: "1024 bytes", raw: strings.Repeat("ab", domain.MaxPasswordBytes/2), want: strings.Repeat("ab", domain.MaxPasswordBytes/2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := domain.NewPassword(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.want, password.Reveal())
		})
	}
}

func TestNewPasswordRejects(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want error
	}{
		{name: "empty", raw: "", want: domain.ErrPasswordTooShort},
		{name: "11 characters", raw: "tlacuache-8", want: domain.ErrPasswordTooShort},
		{name: "12 code points but 11 characters after NFKC", raw: strings.Repeat("x", 10) + "A" + combiningRingAbove, want: domain.ErrPasswordTooShort},
		{name: "over 1024 bytes before normalization", raw: strings.Repeat("a", domain.MaxPasswordBytes+1), want: domain.ErrPasswordTooLong},
		{name: "over 1024 bytes after NFKC expansion", raw: strings.Repeat(expandsUnderNFKC, 40), want: domain.ErrPasswordTooLong},
		{name: "invalid UTF-8", raw: "tlacuache-87\xff", want: domain.ErrPasswordMalformed},
		{name: "common password", raw: "password1234", want: domain.ErrPasswordCommon},
		{name: "common password in another case", raw: "PassWord1234", want: domain.ErrPasswordCommon},
		{name: "common password in fullwidth forms", raw: fullwidthPassword, want: domain.ErrPasswordCommon},
		{name: "contract example", raw: "correct-horse-battery-staple", want: domain.ErrPasswordCommon},
		{name: "service name", raw: "controlcontable", want: domain.ErrPasswordCommon},
		{name: "one repeated character", raw: strings.Repeat("z", 20), want: domain.ErrPasswordCommon},
		{name: "one repeated non-ASCII character", raw: strings.Repeat("ñ", 12), want: domain.ErrPasswordCommon},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewPassword(tt.raw)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

// TestPasswordForVerification covers login: the stored hash was made from
// the NFKC form, so login normalizes the same way but does not re-apply the
// creation policy (a policy change must never lock existing users out).
func TestPasswordForVerification(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "short password still verifiable", raw: "abc", want: "abc"},
		{name: "common password still verifiable", raw: "password1234", want: "password1234"},
		{name: "normalized like NewPassword", raw: fullwidthPassword, want: "password1234"},
		{name: "empty", raw: "", wantErr: domain.ErrPasswordTooShort},
		{name: "over 1024 bytes", raw: strings.Repeat("a", domain.MaxPasswordBytes+1), wantErr: domain.ErrPasswordTooLong},
		{name: "over 1024 bytes after NFKC", raw: strings.Repeat(expandsUnderNFKC, 40), wantErr: domain.ErrPasswordTooLong},
		{name: "invalid UTF-8", raw: "\xff", wantErr: domain.ErrPasswordMalformed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := domain.PasswordForVerification(tt.raw)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, password.Reveal())
		})
	}
}

func TestPasswordErrorsNeverCarryThePassword(t *testing.T) {
	for _, raw := range []string{"zqx", "password1234", strings.Repeat("s", domain.MaxPasswordBytes+1)} {
		_, err := domain.NewPassword(raw)
		require.Error(t, err)
		assert.NotContains(t, err.Error(), raw)
	}
}

func TestPasswordNeverPrints(t *testing.T) {
	accepted, err := domain.NewPassword(twelveCharSample)
	require.NoError(t, err)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		assert.NotContains(t, fmt.Sprintf(verb, accepted), "tlacuache", verb)
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("x", "password", accepted)
	assert.NotContains(t, buf.String(), "tlacuache")
}

// TestMaxPasswordBytesMatchesTheHasher keeps the domain bound and the argon2id
// hasher's defense-in-depth bound identical: a password the policy accepts
// must always be hashable.
func TestMaxPasswordBytesMatchesTheHasher(t *testing.T) {
	assert.Equal(t, crypto.MaxPasswordBytes, domain.MaxPasswordBytes)
}
