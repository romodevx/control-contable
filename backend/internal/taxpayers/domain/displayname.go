package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// MaxDisplayNameLength is the contract's display_name maxLength, in
// characters (runes), as the database CHECK counts them.
const MaxDisplayNameLength = 80

// DisplayName is a taxpayer's label in the switcher (contract display_name):
// 1-80 characters, trimmed, NFC. The zero value is not valid.
type DisplayName struct {
	value string
}

// NewDisplayName trims raw, composes it to NFC and checks it holds 1 to
// MaxDisplayNameLength characters, valid UTF-8 and no control characters,
// else ErrInvalidDisplayName.
func NewDisplayName(raw string) (DisplayName, error) {
	if !utf8.ValidString(raw) {
		return DisplayName{}, ErrInvalidDisplayName
	}
	value := strings.TrimSpace(norm.NFC.String(strings.TrimSpace(raw)))
	length := utf8.RuneCountInString(value)
	if length == 0 || length > MaxDisplayNameLength || strings.ContainsFunc(value, unicode.IsControl) {
		return DisplayName{}, ErrInvalidDisplayName
	}
	return DisplayName{value: value}, nil
}

// String returns the name.
func (n DisplayName) String() string { return n.value }

// IsZero reports whether n is the zero DisplayName.
func (n DisplayName) IsZero() bool { return n.value == "" }
