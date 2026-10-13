package domain

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/unicode/norm"
)

// TestCommonPasswordListIsCanonical keeps every embedded entry reachable:
// candidates are compared in NFKC lower case, and anything shorter than the
// minimum is already rejected by length, so such an entry would be dead.
func TestCommonPasswordListIsCanonical(t *testing.T) {
	require.NotEmpty(t, commonPasswords)
	for entry := range commonPasswords {
		assert.Equal(t, strings.ToLower(norm.NFKC.String(entry)), entry, "entry must be NFKC lower case")
		assert.GreaterOrEqual(t, utf8.RuneCountInString(entry), MinPasswordLength, entry)
		assert.Equal(t, strings.TrimSpace(entry), entry, "entry must have no surrounding space")
	}
}

func TestParseCommonPasswordsSkipsBlankAndCommentLines(t *testing.T) {
	got := parseCommonPasswords("# header\n\nalpha-bravo-charlie\n  \nDelta-Echo-Foxtrot\r\n")
	assert.Equal(t, map[string]struct{}{
		"alpha-bravo-charlie": {},
		"delta-echo-foxtrot":  {},
	}, got)
}
