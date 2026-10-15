// Package usertext holds the one rule every free-text field a user types
// (labels, display names, record descriptions) shares: it may not carry an
// invisible character. Such text is shown back as a single line, so a
// character that renders as nothing or reorders what follows (zero-width
// characters, bidi overrides and isolates, the BOM) only serves to spoof
// what a reader sees.
//
// The package imports only the standard library, so any domain may use it.
// Each module applies it at its own validation step, after NFC and trimming,
// and answers the 400 validation_failed envelope with a FieldError whose
// code is DetailInvisibleCharacters (the message is the module's, in its
// language).
package usertext

import (
	"strings"
	"unicode"
)

// DetailInvisibleCharacters is the FieldError code of a text field holding
// an invisible character.
const DetailInvisibleCharacters = "invisible_characters"

// invisible are the categories refused in user text: control (Cc, line
// breaks and tabs included: the fields are single-line), format (Cf:
// zero-width characters, bidi marks, embeddings, overrides and isolates,
// the BOM, the soft hyphen, tags) and the line and paragraph separators
// (Zl, Zp). Variation selectors and combining marks (Mn) stay allowed:
// they are part of ordinary emoji and accents.
var invisible = []*unicode.RangeTable{unicode.Cc, unicode.Cf, unicode.Zl, unicode.Zp}

// HasInvisible reports whether s holds a control, format, line separator or
// paragraph separator character. It scans s once, O(len(s)).
func HasInvisible(s string) bool {
	return strings.ContainsFunc(s, isInvisible)
}

func isInvisible(r rune) bool {
	return unicode.In(r, invisible...)
}
