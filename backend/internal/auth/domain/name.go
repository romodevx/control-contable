package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// MaxNameLength is the longest display name, in characters (contract
// User.name maxLength).
const MaxNameLength = 120

// Name is a person's optional display name: NFC-normalized, surrounding
// whitespace trimmed, 1 to MaxNameLength characters, no control characters.
// It is PII but not secret, so it prints normally.
type Name struct {
	text string
}

// NewName validates and normalizes raw.
func NewName(raw string) (Name, error) {
	if !utf8.ValidString(raw) {
		return Name{}, ErrInvalidName
	}
	text := strings.TrimSpace(norm.NFC.String(raw))
	length := utf8.RuneCountInString(text)
	if length == 0 || length > MaxNameLength || strings.IndexFunc(text, unicode.IsControl) >= 0 {
		return Name{}, ErrInvalidName
	}
	return Name{text: text}, nil
}

// String returns the normalized name.
func (n Name) String() string { return n.text }

// IsZero reports whether n is the zero Name (no name given).
func (n Name) IsZero() bool { return n.text == "" }
