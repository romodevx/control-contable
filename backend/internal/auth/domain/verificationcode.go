package domain

import (
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	verificationCodeDigits = 6
	verificationCodeSpace  = 1_000_000 // 10^verificationCodeDigits
	// unbiasedDrawLimit is the largest multiple of verificationCodeSpace that
	// fits in a uint32. Draws at or above it are rejected, so draw % space is
	// uniform (rejection sampling; a draw is rejected with p ≈ 2.3·10⁻⁴).
	unbiasedDrawLimit = (1 << 32) / verificationCodeSpace * verificationCodeSpace
)

// VerificationCode is the 6-digit code emailed at registration. It renders
// as REDACTED in fmt and slog; Reveal returns the digits for the email body.
type VerificationCode struct {
	redactor
	digits string
}

// NewVerificationCode draws a uniformly distributed code from entropy, which
// must be a CSPRNG (crypto/rand.Reader in production).
func NewVerificationCode(entropy io.Reader) (VerificationCode, error) {
	var draw [4]byte
	for {
		if _, err := io.ReadFull(entropy, draw[:]); err != nil {
			return VerificationCode{}, fmt.Errorf("%w: %w", ErrRandomSource, err)
		}
		value := binary.BigEndian.Uint32(draw[:])
		if value < unbiasedDrawLimit {
			return VerificationCode{digits: formatCode(value % verificationCodeSpace)}, nil
		}
	}
}

// ParseVerificationCode parses a code typed by the user: exactly six ASCII
// digits, nothing trimmed (the contract pattern is ^[0-9]{6}$).
func ParseVerificationCode(text string) (VerificationCode, error) {
	if len(text) != verificationCodeDigits || strings.IndexFunc(text, isNotASCIIDigit) >= 0 {
		return VerificationCode{}, ErrInvalidVerificationCode
	}
	return VerificationCode{digits: text}, nil
}

// Reveal returns the six digits, for the verification email only.
func (c VerificationCode) Reveal() string { return c.digits }

// Equal compares two codes in constant time.
func (c VerificationCode) Equal(other VerificationCode) bool {
	return subtle.ConstantTimeCompare([]byte(c.digits), []byte(other.digits)) == 1
}

func formatCode(value uint32) string {
	digits := strconv.FormatUint(uint64(value), 10)
	return strings.Repeat("0", verificationCodeDigits-len(digits)) + digits
}

func isNotASCIIDigit(r rune) bool { return r < '0' || r > '9' }
