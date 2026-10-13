package domain

import (
	_ "embed" // the common-password blocklist ships inside the binary
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	// MinPasswordLength is the NIST SP 800-63B-4 minimum for a password used
	// as a single factor, counted in Unicode characters after NFKC.
	MinPasswordLength = 12
	// MaxPasswordBytes bounds a password before and after normalization. It
	// equals the argon2id hasher's crypto.MaxPasswordBytes, so every password
	// the policy accepts can be hashed (TestMaxPasswordBytesMatchesTheHasher).
	MaxPasswordBytes = 1024
)

//go:embed common_passwords.txt
var commonPasswordsFile string

// commonPasswords is the blocklist, keyed by NFKC lower-case form. It is
// built once at init and only read afterwards, so it is safe for concurrent use.
var commonPasswords = parseCommonPasswords(commonPasswordsFile)

// Password is a password in NFKC form. NIST SP 800-63B-4 asks verifiers to
// normalize before hashing so that the same characters typed on different
// keyboards or platforms produce the same hash. It renders as REDACTED in
// fmt and slog; Reveal returns the text to hash or verify.
type Password struct {
	redactor
	normalized string
}

// NewPassword applies the account-creation policy of NIST SP 800-63B-4
// section 3.1.1.2: NFKC normalization, at least MinPasswordLength
// characters, at most MaxPasswordBytes bytes, not a common password, and no
// composition rules (any character, spaces included, is allowed).
func NewPassword(raw string) (Password, error) {
	password, err := PasswordForVerification(raw)
	if err != nil {
		return Password{}, err
	}
	if utf8.RuneCountInString(password.normalized) < MinPasswordLength {
		return Password{}, ErrPasswordTooShort
	}
	if isCommonPassword(password.normalized) {
		return Password{}, ErrPasswordCommon
	}
	return password, nil
}

// PasswordForVerification normalizes a password typed at login so it
// matches the form NewPassword hashed. It applies only the encoding and size
// bounds, never the creation policy: tightening the policy must not lock
// existing accounts out.
func PasswordForVerification(raw string) (Password, error) {
	switch {
	case raw == "":
		return Password{}, ErrPasswordTooShort
	case len(raw) > MaxPasswordBytes:
		// Checked before normalizing, so oversized input costs no NFKC work.
		return Password{}, ErrPasswordTooLong
	case !utf8.ValidString(raw):
		return Password{}, ErrPasswordMalformed
	}
	normalized := norm.NFKC.String(raw)
	if len(normalized) > MaxPasswordBytes {
		return Password{}, ErrPasswordTooLong
	}
	return Password{normalized: normalized}, nil
}

// Reveal returns the NFKC-normalized password, for hashing or verification only.
func (p Password) Reveal() string { return p.normalized }

// isCommonPassword reports a blocklisted password or one that repeats a
// single character, the "repetitive characters" case of NIST SP 800-63B.
func isCommonPassword(normalized string) bool {
	if _, listed := commonPasswords[strings.ToLower(normalized)]; listed {
		return true
	}
	first, _ := utf8.DecodeRuneInString(normalized)
	return strings.Trim(normalized, string(first)) == ""
}

// parseCommonPasswords reads one entry per line, skipping blank lines and
// lines starting with '#'. Entries are stored in NFKC lower case.
func parseCommonPasswords(file string) map[string]struct{} {
	entries := make(map[string]struct{})
	for line := range strings.Lines(file) {
		entry := strings.TrimSpace(line)
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		entries[strings.ToLower(norm.NFKC.String(entry))] = struct{}{}
	}
	return entries
}
