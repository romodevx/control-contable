package usertext_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

func TestHasInvisible(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "plain ASCII", text: "Empleo principal", want: false},
		{name: "Spanish letters and accents", text: "Hija menor, Peñasco — año 2026", want: false},
		{name: "combining accent (NFD)", text: "Jose\u0301", want: false},
		{name: "emoji with variation selector", text: "Casa \u2764\ufe0f", want: false},
		{name: "inner spaces, no-break space included", text: "Renta\u00a0local 2", want: false},
		{name: "empty", text: "", want: false},

		{name: "zero-width space U+200B", text: "Empleo\u200bprincipal", want: true},
		{name: "zero-width non-joiner U+200C", text: "a\u200cb", want: true},
		{name: "zero-width joiner U+200D", text: "a\u200db", want: true},
		{name: "left-to-right mark U+200E", text: "a\u200eb", want: true},
		{name: "right-to-left mark U+200F", text: "a\u200fb", want: true},
		{name: "LRE U+202A", text: "a\u202ab", want: true},
		{name: "RLE U+202B", text: "a\u202bb", want: true},
		{name: "PDF U+202C", text: "a\u202cb", want: true},
		{name: "LRO U+202D", text: "a\u202db", want: true},
		{name: "RLO U+202E", text: "factura\u202efdp.exe", want: true},
		{name: "LRI U+2066", text: "a\u2066b", want: true},
		{name: "RLI U+2067", text: "a\u2067b", want: true},
		{name: "FSI U+2068", text: "a\u2068b", want: true},
		{name: "PDI U+2069", text: "a\u2069b", want: true},
		{name: "word joiner U+2060", text: "a\u2060b", want: true},
		{name: "BOM U+FEFF", text: "\ufeffEmpleo", want: true},
		{name: "soft hyphen U+00AD", text: "Em\u00adpleo", want: true},
		{name: "Arabic letter mark U+061C", text: "a\u061cb", want: true},
		{name: "Mongolian vowel separator U+180E", text: "a\u180eb", want: true},
		{name: "interlinear annotation U+FFF9", text: "a\ufff9b", want: true},
		{name: "tag character U+E0041", text: "a\U000e0041b", want: true},
		{name: "line separator U+2028", text: "a\u2028b", want: true},
		{name: "paragraph separator U+2029", text: "a\u2029b", want: true},
		{name: "control NUL", text: "a\x00b", want: true},
		{name: "control newline", text: "a\nb", want: true},
		{name: "control tab", text: "a\tb", want: true},
		{name: "C1 control U+0085", text: "a\u0085b", want: true},
		{name: "DEL", text: "a\x7fb", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, usertext.HasInvisible(tt.text))
		})
	}
}

// The detail code is part of the error envelope every module returns, so
// clients can key their copy on it.
func TestDetailInvisibleCharacters(t *testing.T) {
	assert.Equal(t, "invisible_characters", usertext.DetailInvisibleCharacters)
	assert.NotEmpty(t, usertext.MessageInvisibleCharacters)
}
