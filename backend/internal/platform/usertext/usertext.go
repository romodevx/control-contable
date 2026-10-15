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
	"unicode"
	"unicode/utf8"
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

const (
	// zeroWidthJoiner (Cf) glues emoji into one glyph (a family, a
	// profession, the rainbow flag); it is allowed only there.
	zeroWidthJoiner = '\u200d'
	// emojiPresentation (VS16) may follow the emoji a joiner continues.
	emojiPresentation = '\ufe0f'
	// skinToneFirst and skinToneLast bound the emoji modifiers (Fitzpatrick
	// skin tones) that may follow the emoji a joiner continues.
	skinToneFirst = '\U0001f3fb'
	skinToneLast  = '\U0001f3ff'
)

// HasInvisible reports whether s holds a control, format, line separator or
// paragraph separator character. The one exception is a zero-width joiner
// inside an emoji ZWJ sequence: between an emoji (optionally followed by
// VS16 or a skin tone) and another emoji, where it renders as one glyph and
// cannot hide or reorder letters. It scans s once, O(len(s)).
func HasInvisible(s string) bool {
	afterEmoji := false
	for i, r := range s {
		switch {
		case r == zeroWidthJoiner:
			if !afterEmoji || !emojiStarts(s[i+utf8.RuneLen(r):]) {
				return true
			}
		case isInvisible(r):
			return true
		case isEmoji(r):
			afterEmoji = true
		case !extendsEmoji(r):
			afterEmoji = false
		}
	}
	return false
}

func isInvisible(r rune) bool {
	return unicode.In(r, invisible...)
}

// isEmoji approximates Extended_Pictographic, which the standard library
// lacks, with Symbol, other (So): every emoji a ZWJ sequence joins is So.
func isEmoji(r rune) bool {
	return unicode.Is(unicode.So, r)
}

// emojiStarts reports whether rest begins with an emoji. An empty rest
// does not (DecodeRuneInString answers RuneError, itself So).
func emojiStarts(rest string) bool {
	next, size := utf8.DecodeRuneInString(rest)
	return size > 0 && next != utf8.RuneError && isEmoji(next)
}

// extendsEmoji reports whether r may sit between an emoji and the joiner
// that continues it.
func extendsEmoji(r rune) bool {
	return r == emojiPresentation || (r >= skinToneFirst && r <= skinToneLast)
}
