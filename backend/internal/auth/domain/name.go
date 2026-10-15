package domain

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

// MaxNameLength is the longest display name, in characters (contract
// User.name maxLength).
const MaxNameLength = 120

// Name is a person's optional display name: NFC-normalized, surrounding
// whitespace trimmed, 1 to MaxNameLength characters, no invisible characters.
// It is PII but not secret, so it prints normally.
type Name struct {
	text string
}

// NewName validates and normalizes raw: ErrInvalidName when it is not 1 to
// MaxNameLength characters of valid UTF-8, ErrNameInvisible when it holds
// an invisible character.
func NewName(raw string) (Name, error) {
	if !utf8.ValidString(raw) {
		return Name{}, ErrInvalidName
	}
	text := strings.TrimSpace(norm.NFC.String(raw))
	length := utf8.RuneCountInString(text)
	if length == 0 || length > MaxNameLength {
		return Name{}, ErrInvalidName
	}
	if usertext.HasInvisible(text) {
		return Name{}, ErrNameInvisible
	}
	return Name{text: text}, nil
}

// String returns the normalized name.
func (n Name) String() string { return n.text }

// IsZero reports whether n is the zero Name (no name given).
func (n Name) IsZero() bool { return n.text == "" }
