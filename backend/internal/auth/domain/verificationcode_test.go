package domain_test

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// unbiasedLimit is the largest multiple of 10^6 that fits in a uint32
// (2^32 − 2^32 mod 10^6); draws at or above it are rejected.
const unbiasedLimit = 4_294_000_000

func draws(values ...uint32) io.Reader {
	var buf bytes.Buffer
	for _, v := range values {
		_ = binary.Write(&buf, binary.BigEndian, v)
	}
	return &buf
}

func TestNewVerificationCode(t *testing.T) {
	tests := []struct {
		name    string
		entropy io.Reader
		want    string
	}{
		{name: "zero is zero-padded", entropy: draws(0), want: "000000"},
		{name: "plain value", entropy: draws(123_456), want: "123456"},
		{name: "reduced modulo 10^6", entropy: draws(7_000_042), want: "000042"},
		{name: "largest accepted draw", entropy: draws(unbiasedLimit - 1), want: "999999"},
		{name: "biased draw rejected, next one used", entropy: draws(unbiasedLimit, 654_321), want: "654321"},
		{name: "max uint32 rejected", entropy: draws(^uint32(0), 1), want: "000001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := domain.NewVerificationCode(tt.entropy)
			require.NoError(t, err)
			assert.Equal(t, tt.want, code.Reveal())
		})
	}
}

func TestNewVerificationCodeFromCryptoRand(t *testing.T) {
	sixDigits := regexp.MustCompile(`^[0-9]{6}$`)
	for range 100 {
		code, err := domain.NewVerificationCode(rand.Reader)
		require.NoError(t, err)
		assert.Regexp(t, sixDigits, code.Reveal())
	}
}

func TestNewVerificationCodeEntropyErrors(t *testing.T) {
	t.Run("exhausted after a rejected draw", func(t *testing.T) {
		_, err := domain.NewVerificationCode(draws(unbiasedLimit))
		require.ErrorIs(t, err, domain.ErrRandomSource)
	})
	t.Run("short read", func(t *testing.T) {
		_, err := domain.NewVerificationCode(bytes.NewReader([]byte{1, 2}))
		require.ErrorIs(t, err, domain.ErrRandomSource)
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	})
	t.Run("failing reader", func(t *testing.T) {
		boom := errors.New("boom")
		_, err := domain.NewVerificationCode(failingReader{err: boom})
		require.ErrorIs(t, err, boom)
	})
}

func TestParseVerificationCode(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "six digits", raw: "482913"},
		{name: "leading zeros", raw: "000042"},
		{name: "empty", raw: "", wantErr: true},
		{name: "five digits", raw: "48291", wantErr: true},
		{name: "seven digits", raw: "4829130", wantErr: true},
		{name: "letter", raw: "48291a", wantErr: true},
		{name: "surrounding space", raw: " 482913", wantErr: true},
		{name: "arabic-indic digits", raw: "٤٨٢٩١٣", wantErr: true},
		{name: "fullwidth digits", raw: "４８２９１３", wantErr: true},
		{name: "sign", raw: "-48291", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := domain.ParseVerificationCode(tt.raw)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidVerificationCode)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.raw, code.Reveal())
		})
	}
}

func TestVerificationCodeEqual(t *testing.T) {
	a, err := domain.ParseVerificationCode("482913")
	require.NoError(t, err)
	same, err := domain.NewVerificationCode(draws(482_913))
	require.NoError(t, err)
	other, err := domain.ParseVerificationCode("482914")
	require.NoError(t, err)

	assert.True(t, a.Equal(same))
	assert.False(t, a.Equal(other))
}

func TestVerificationCodeNeverPrints(t *testing.T) {
	code, err := domain.ParseVerificationCode("482913")
	require.NoError(t, err)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%d"} {
		assert.NotContains(t, fmt.Sprintf(verb, code), "482913", verb)
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("x", "code", code)
	assert.NotContains(t, buf.String(), "482913")
}
