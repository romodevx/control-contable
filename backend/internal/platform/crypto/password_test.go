package crypto

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cheapParams keeps the unit suite fast under -race; production uses
// DefaultArgon2idParams, exercised once in TestDefaultParamsRoundTrip.
var cheapParams = Argon2idParams{Memory: 64, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}

func newCheapHasher(t *testing.T) *PasswordHasher {
	t.Helper()
	h, err := NewPasswordHasher(cheapParams)
	require.NoError(t, err)
	return h
}

func TestDefaultArgon2idParamsMatchOWASP(t *testing.T) {
	p := DefaultArgon2idParams()

	assert.Equal(t, uint32(19*1024), p.Memory, "OWASP: m=19 MiB")
	assert.Equal(t, uint32(2), p.Time, "OWASP: t=2")
	assert.Equal(t, uint8(1), p.Threads, "OWASP: p=1")
	assert.Equal(t, uint32(16), p.SaltLen)
	assert.Equal(t, uint32(32), p.KeyLen)
}

func TestDefaultParamsRoundTrip(t *testing.T) {
	h, err := NewPasswordHasher(DefaultArgon2idParams())
	require.NoError(t, err)

	encoded, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(encoded, "$argon2id$v=19$m=19456,t=2,p=1$"), encoded)

	ok, err := h.Verify("correct horse battery staple", encoded)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestNewPasswordHasherRejectsInvalidParams(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Argon2idParams)
	}{
		{"zero time", func(p *Argon2idParams) { p.Time = 0 }},
		{"time above max", func(p *Argon2idParams) { p.Time = maxTime + 1 }},
		{"zero threads", func(p *Argon2idParams) { p.Threads = 0 }},
		{"threads above max", func(p *Argon2idParams) { p.Threads = maxThreads + 1 }},
		{"memory below 8 KiB per thread", func(p *Argon2idParams) { p.Threads = 2; p.Memory = 15 }},
		{"memory above max", func(p *Argon2idParams) { p.Memory = maxMemoryKiB + 1 }},
		{"salt too short", func(p *Argon2idParams) { p.SaltLen = minSaltLen - 1 }},
		{"salt too long", func(p *Argon2idParams) { p.SaltLen = maxSaltLen + 1 }},
		{"key too short", func(p *Argon2idParams) { p.KeyLen = minKeyLen - 1 }},
		{"key too long", func(p *Argon2idParams) { p.KeyLen = maxKeyLen + 1 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := cheapParams
			tt.mutate(&p)

			h, err := NewPasswordHasher(p)

			require.ErrorIs(t, err, ErrInvalidParams)
			assert.Nil(t, h)
		})
	}
}

// The cost caps bound what one Verify of a stored (possibly tampered) hash can
// cost: at most 256 MiB of memory and 256 MiB·passes (memory × time). They
// still admit every OWASP option and the RFC 9106 second recommendation.
func TestArgon2idCostCaps(t *testing.T) {
	const mib = 1024
	tests := []struct {
		name    string
		memory  uint32
		time    uint32
		threads uint8
		wantErr error
	}{
		{"OWASP m=46MiB t=1", 46 * mib, 1, 1, nil},
		{"OWASP m=19MiB t=2", 19 * mib, 2, 1, nil},
		{"OWASP m=7MiB t=5", 7 * mib, 5, 1, nil},
		{"RFC 9106 m=64MiB t=3 p=4", 64 * mib, 3, 4, nil},
		{"memory at cap", 256 * mib, 1, 1, nil},
		{"memory above cap (1 GiB)", 1024 * mib, 1, 1, ErrInvalidParams},
		{"memory x time above cap", 64 * mib, 5, 1, ErrInvalidParams},
		{"max time with large memory", 128 * mib, 16, 1, ErrInvalidParams},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Argon2idParams{Memory: tt.memory, Time: tt.time, Threads: tt.threads, SaltLen: 16, KeyLen: 32}

			_, err := NewPasswordHasher(p)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// Regression: a stored hash whose cost exceeds the caps is rejected while
// parsing, before any argon2 work, so a tampered row cannot exhaust memory.
func TestParsePHCRejectsCostBombs(t *testing.T) {
	const salt, key = "c2FsdHNhbHRzYWx0c2FsdA", "a2V5a2V5a2V5a2V5a2V5a2V5a2V5a2V5a2V5a2U"
	for name, params := range map[string]string{
		"1 GiB memory":       "m=1048576,t=1,p=1",
		"1 GiB with t=16":    "m=1048576,t=16,p=16",
		"memory x time bomb": "m=131072,t=16,p=1",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parsePHC("$argon2id$v=19$" + params + "$" + salt + "$" + key)

			require.ErrorIs(t, err, ErrMalformedHash)
		})
	}
}

func TestHashProducesPHCStringWithFreshSalt(t *testing.T) {
	h := newCheapHasher(t)

	first, err := h.Hash("s3cret")
	require.NoError(t, err)
	second, err := h.Hash("s3cret")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(first, "$argon2id$v=19$m=64,t=1,p=1$"), first)
	assert.Len(t, strings.Split(first, "$"), phcFieldCount)
	assert.NotEqual(t, first, second, "each hash must use a fresh random salt")
}

func TestHashFailsWhenRandomSourceFails(t *testing.T) {
	h := newCheapHasher(t)
	h.random = failingReader{}

	encoded, err := h.Hash("s3cret")

	require.Error(t, err)
	assert.Empty(t, encoded)
	assert.NotContains(t, err.Error(), "s3cret")
}

func TestVerify(t *testing.T) {
	h := newCheapHasher(t)
	encoded, err := h.Hash("s3cret")
	require.NoError(t, err)
	parts := strings.Split(encoded, "$")
	salt, hash := parts[4], parts[5]

	tests := []struct {
		name     string
		password string
		encoded  string
		wantOK   bool
		wantErr  error
	}{
		{"correct password", "s3cret", encoded, true, nil},
		{"wrong password", "s3cret!", encoded, false, nil},
		{"empty password", "", encoded, false, nil},
		{"tampered hash", "s3cret", replaceField(encoded, 5, flipFirstChar(hash)), false, nil},
		{"tampered salt", "s3cret", replaceField(encoded, 4, flipFirstChar(salt)), false, nil},
		{"tampered cost", "s3cret", replaceField(encoded, 3, "m=64,t=2,p=1"), false, nil},
		{"empty string", "s3cret", "", false, ErrMalformedHash},
		{"bcrypt hash", "s3cret", "$2a$10$abcdefghijklmnopqrstuuMx2eTXdJp2kM9fLqGyQe5hB2ZpQ3a3e", false, ErrMalformedHash},
		{"argon2i variant", "s3cret", replaceField(encoded, 1, "argon2i"), false, ErrMalformedHash},
		{"unsupported version", "s3cret", replaceField(encoded, 2, "v=16"), false, ErrMalformedHash},
		{"version not a number", "s3cret", replaceField(encoded, 2, "v=x"), false, ErrMalformedHash},
		{"params reordered", "s3cret", replaceField(encoded, 3, "t=1,m=64,p=1"), false, ErrMalformedHash},
		{"params missing", "s3cret", replaceField(encoded, 3, "m=64,t=1"), false, ErrMalformedHash},
		{"params not numbers", "s3cret", replaceField(encoded, 3, "m=a,t=1,p=1"), false, ErrMalformedHash},
		{"threads overflow uint8", "s3cret", replaceField(encoded, 3, "m=64,t=1,p=256"), false, ErrMalformedHash},
		{"memory above max (DoS guard)", "s3cret", replaceField(encoded, 3, "m=99999999,t=1,p=1"), false, ErrMalformedHash},
		{"time zero", "s3cret", replaceField(encoded, 3, "m=64,t=0,p=1"), false, ErrMalformedHash},
		{"salt not base64", "s3cret", replaceField(encoded, 4, "!!!!"), false, ErrMalformedHash},
		{"salt padded base64", "s3cret", replaceField(encoded, 4, salt+"=="), false, ErrMalformedHash},
		{"salt too short", "s3cret", replaceField(encoded, 4, "YWJj"), false, ErrMalformedHash},
		{"hash not base64", "s3cret", replaceField(encoded, 5, "!!!!"), false, ErrMalformedHash},
		{"hash too short", "s3cret", replaceField(encoded, 5, "YWJj"), false, ErrMalformedHash},
		{"extra field", "s3cret", encoded + "$extra", false, ErrMalformedHash},
		{"missing leading dollar", "s3cret", strings.TrimPrefix(encoded, "$"), false, ErrMalformedHash},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := h.Verify(tt.password, tt.encoded)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.NotContains(t, err.Error(), tt.password, "errors never carry the password")
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestVerifyAcceptsHashesMadeWithOtherParams(t *testing.T) {
	old, err := NewPasswordHasher(Argon2idParams{Memory: 32, Time: 2, Threads: 1, SaltLen: 8, KeyLen: 16})
	require.NoError(t, err)
	encoded, err := old.Hash("s3cret")
	require.NoError(t, err)

	ok, err := newCheapHasher(t).Verify("s3cret", encoded)

	require.NoError(t, err)
	assert.True(t, ok, "parameters come from the PHC string, not the current policy")
}

func TestNeedsRehash(t *testing.T) {
	h := newCheapHasher(t)
	current, err := h.Hash("s3cret")
	require.NoError(t, err)

	hashWith := func(p Argon2idParams) string {
		other, err := NewPasswordHasher(p)
		require.NoError(t, err)
		encoded, err := other.Hash("s3cret")
		require.NoError(t, err)
		return encoded
	}
	with := func(mutate func(*Argon2idParams)) Argon2idParams {
		p := cheapParams
		mutate(&p)
		return p
	}

	tests := []struct {
		name    string
		encoded string
		want    bool
		wantErr error
	}{
		{"current params", current, false, nil},
		{"lower memory", hashWith(with(func(p *Argon2idParams) { p.Memory = 32 })), true, nil},
		{"higher memory", hashWith(with(func(p *Argon2idParams) { p.Memory = 128 })), true, nil},
		{"different time", hashWith(with(func(p *Argon2idParams) { p.Time = 2 })), true, nil},
		{"different threads", hashWith(with(func(p *Argon2idParams) { p.Threads = 2 })), true, nil},
		{"shorter salt", hashWith(with(func(p *Argon2idParams) { p.SaltLen = 8 })), true, nil},
		{"shorter key", hashWith(with(func(p *Argon2idParams) { p.KeyLen = 16 })), true, nil},
		{"malformed", "$argon2id$garbage", false, ErrMalformedHash},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.NeedsRehash(tt.encoded)

			assert.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// failingReader simulates an exhausted entropy source.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("entropy unavailable") }

func replaceField(encoded string, index int, value string) string {
	parts := strings.Split(encoded, "$")
	parts[index] = value
	return strings.Join(parts, "$")
}

func flipFirstChar(s string) string {
	replacement := "A"
	if s[0] == 'A' {
		replacement = "B"
	}
	return replacement + s[1:]
}
