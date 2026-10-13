package domain

import (
	"strings"
)

// RFC 5321 section 4.5.3.1 size limits, in octets. A path is at most 256
// octets including the angle brackets, so an address is at most 254.
const (
	maxEmailLength     = 254
	maxLocalPartLength = 64
	maxLabelLength     = 63
	minDomainLabels    = 2
)

// Email is a normalized email address: surrounding whitespace trimmed and
// A-Z lower-cased, exactly as the email_hash blind index normalizes it, so
// String() and the raw input always hash to the same index. The zero value
// is not a valid address (see IsZero).
type Email struct {
	redactor
	address string
}

// NewEmail validates and normalizes raw. It accepts only ASCII dot-atom
// addresses (RFC 5322 section 3.2.3) on a host name of at least two LDH
// labels: no quoted local parts, IP literals or internationalized
// addresses. Non-ASCII is rejected rather than case-mapped, so a look-alike
// such as U+212A KELVIN SIGN can never collide with an ASCII "k".
func NewEmail(raw string) (Email, error) {
	address := strings.TrimSpace(raw)
	if len(address) > maxEmailLength {
		return Email{}, ErrInvalidEmail
	}
	address = strings.Map(lowerASCII, address)
	local, host, found := strings.Cut(address, "@")
	if !found || !isDotAtom(local) || len(local) > maxLocalPartLength || !isHostName(host) {
		return Email{}, ErrInvalidEmail
	}
	return Email{address: address}, nil
}

// String returns the normalized address. Email is PII: it renders as
// REDACTED in fmt and slog, so only an explicit String() call reveals it.
func (e Email) String() string { return e.address }

// IsZero reports whether e is the zero Email.
func (e Email) IsZero() bool { return e.address == "" }

// lowerASCII lower-cases A-Z only; strings.ToLower would also fold Unicode.
func lowerASCII(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}

// isDotAtom reports whether s is one or more atext runs joined by single dots.
func isDotAtom(s string) bool {
	for atom := range strings.SplitSeq(s, ".") {
		if atom == "" || strings.IndexFunc(atom, isNotAtext) >= 0 {
			return false
		}
	}
	return true
}

// isNotAtext reports a rune outside RFC 5322 atext (letters, digits and
// !#$%&'*+-/=?^_`{|}~). Every non-ASCII rune is outside it.
func isNotAtext(r rune) bool {
	return !isAlphaNumeric(r) && !strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", r)
}

// isHostName reports whether s is at least minDomainLabels LDH labels
// (RFC 1035 section 2.3.1), each 1 to 63 octets, none starting or ending
// with a hyphen.
func isHostName(s string) bool {
	labels := strings.Split(s, ".")
	if len(labels) < minDomainLabels {
		return false
	}
	for _, label := range labels {
		if !isLabel(label) {
			return false
		}
	}
	return true
}

func isLabel(label string) bool {
	if label == "" || len(label) > maxLabelLength {
		return false
	}
	if label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	return strings.IndexFunc(label, isNotLDH) < 0
}

func isNotLDH(r rune) bool { return !isAlphaNumeric(r) && r != '-' }

// isAlphaNumeric reports an ASCII letter or digit; input is already lower-cased.
func isAlphaNumeric(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') }
