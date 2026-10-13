package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// OWASP Password Storage Cheat Sheet (2026), argon2id first configuration:
// m=19 MiB, t=2, p=1. Salt and output sizes follow RFC 9106 section 4.
const (
	owaspMemoryKiB = 19 * 1024
	owaspTime      = 2
	owaspThreads   = 1
	defaultSaltLen = 16
	defaultKeyLen  = 32
)

// Bounds accepted for the policy and for stored hashes. The upper bounds stop
// a tampered row from turning a login into a memory or CPU exhaustion attack:
// one Verify costs at most 256 MiB and 256 MiB·passes (memory × time), which
// still admits every OWASP option and the RFC 9106 m=64MiB,t=3 setting.
const (
	minMemoryPerThreadKiB = 8 // argon2 needs 8 KiB per lane
	maxMemoryKiB          = 256 * 1024
	maxCostKiB            = 256 * 1024 // memory (KiB) × time
	maxTime               = 16
	maxThreads            = 16
	minSaltLen            = 8
	maxSaltLen            = 64
	minKeyLen             = 16
	maxKeyLen             = 64
)

const (
	phcAlgorithm  = "argon2id"
	phcFieldCount = 6 // "", algorithm, version, params, salt, hash
)

var phcEncoding = base64.RawStdEncoding

// Argon2idParams are the argon2id cost parameters. Memory is in KiB.
type Argon2idParams struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	SaltLen uint32
	KeyLen  uint32
}

// DefaultArgon2idParams returns the OWASP-recommended production parameters.
func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{
		Memory:  owaspMemoryKiB,
		Time:    owaspTime,
		Threads: owaspThreads,
		SaltLen: defaultSaltLen,
		KeyLen:  defaultKeyLen,
	}
}

func (p Argon2idParams) validate() error {
	switch {
	case p.Time < 1 || p.Time > maxTime:
		return fmt.Errorf("%w: time %d not in [1, %d]", ErrInvalidParams, p.Time, maxTime)
	case p.Threads < 1 || p.Threads > maxThreads:
		return fmt.Errorf("%w: threads %d not in [1, %d]", ErrInvalidParams, p.Threads, maxThreads)
	case p.Memory < minMemoryPerThreadKiB*uint32(p.Threads) || p.Memory > maxMemoryKiB:
		return fmt.Errorf("%w: memory %d KiB out of bounds", ErrInvalidParams, p.Memory)
	case uint64(p.Memory)*uint64(p.Time) > maxCostKiB:
		return fmt.Errorf("%w: memory × time above %d KiB", ErrInvalidParams, maxCostKiB)
	case p.SaltLen < minSaltLen || p.SaltLen > maxSaltLen:
		return fmt.Errorf("%w: salt length %d not in [%d, %d]", ErrInvalidParams, p.SaltLen, minSaltLen, maxSaltLen)
	case p.KeyLen < minKeyLen || p.KeyLen > maxKeyLen:
		return fmt.Errorf("%w: key length %d not in [%d, %d]", ErrInvalidParams, p.KeyLen, minKeyLen, maxKeyLen)
	}
	return nil
}

// PasswordHasher hashes and verifies passwords with argon2id. It is safe for
// concurrent use.
type PasswordHasher struct {
	params Argon2idParams
	random io.Reader
}

// NewPasswordHasher returns a hasher that creates hashes with params.
func NewPasswordHasher(params Argon2idParams) (*PasswordHasher, error) {
	if err := params.validate(); err != nil {
		return nil, err
	}
	return &PasswordHasher{params: params, random: rand.Reader}, nil
}

// Hash returns the PHC string of password under a fresh random salt:
// $argon2id$v=19$m=<KiB>,t=<time>,p=<threads>$<salt>$<hash>.
func (h *PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.params.SaltLen)
	if _, err := io.ReadFull(h.random, salt); err != nil {
		return "", fmt.Errorf("crypto: read salt: %w", err)
	}
	p := h.params
	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		phcAlgorithm, argon2.Version, p.Memory, p.Time, p.Threads,
		phcEncoding.EncodeToString(salt), phcEncoding.EncodeToString(key)), nil
}

// Verify reports whether password matches encoded, using the parameters
// stored in encoded. A mismatch is (false, nil); an unparsable hash is
// ErrMalformedHash. The comparison runs in constant time.
func (h *PasswordHasher) Verify(password, encoded string) (bool, error) {
	stored, err := parsePHC(encoded)
	if err != nil {
		return false, err
	}
	p := stored.params
	key := argon2.IDKey([]byte(password), stored.salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return subtle.ConstantTimeCompare(key, stored.key) == 1, nil
}

// NeedsRehash reports whether encoded was made with parameters other than the
// hasher's current policy. Call it after a successful Verify and store a new
// Hash when it returns true.
func (h *PasswordHasher) NeedsRehash(encoded string) (bool, error) {
	stored, err := parsePHC(encoded)
	if err != nil {
		return false, err
	}
	return stored.params != h.params, nil
}

type phcHash struct {
	params Argon2idParams
	salt   []byte
	key    []byte
}

func parsePHC(encoded string) (phcHash, error) {
	fields := strings.Split(encoded, "$")
	if len(fields) != phcFieldCount || fields[0] != "" || fields[1] != phcAlgorithm {
		return phcHash{}, fmt.Errorf("%w: not an argon2id PHC string", ErrMalformedHash)
	}
	if fields[2] != "v="+strconv.Itoa(argon2.Version) {
		return phcHash{}, fmt.Errorf("%w: unsupported version", ErrMalformedHash)
	}
	params, err := parsePHCParams(fields[3])
	if err != nil {
		return phcHash{}, err
	}
	salt, err := phcEncoding.Strict().DecodeString(fields[4])
	if err != nil {
		return phcHash{}, fmt.Errorf("%w: salt encoding", ErrMalformedHash)
	}
	key, err := phcEncoding.Strict().DecodeString(fields[5])
	if err != nil {
		return phcHash{}, fmt.Errorf("%w: hash encoding", ErrMalformedHash)
	}
	params.SaltLen, params.KeyLen = uint32(len(salt)), uint32(len(key)) //nolint:gosec // bounded by validate below
	if err := params.validate(); err != nil {
		return phcHash{}, fmt.Errorf("%w: %w", ErrMalformedHash, err)
	}
	return phcHash{params: params, salt: salt, key: key}, nil
}

// parsePHCParams reads "m=<KiB>,t=<time>,p=<threads>" in the PHC order.
func parsePHCParams(field string) (Argon2idParams, error) {
	var memory, time, threads uint64
	targets := []struct {
		name string
		dst  *uint64
		bits int
	}{{"m", &memory, 32}, {"t", &time, 32}, {"p", &threads, 8}}

	pairs := strings.Split(field, ",")
	if len(pairs) != len(targets) {
		return Argon2idParams{}, fmt.Errorf("%w: want m,t,p parameters", ErrMalformedHash)
	}
	for i, target := range targets {
		value, ok := strings.CutPrefix(pairs[i], target.name+"=")
		if !ok {
			return Argon2idParams{}, fmt.Errorf("%w: want parameter %q", ErrMalformedHash, target.name)
		}
		n, err := strconv.ParseUint(value, 10, target.bits)
		if err != nil {
			return Argon2idParams{}, fmt.Errorf("%w: parameter %q", ErrMalformedHash, target.name)
		}
		*target.dst = n
	}
	return Argon2idParams{
		Memory:  uint32(memory),
		Time:    uint32(time),
		Threads: uint8(threads),
	}, nil
}
