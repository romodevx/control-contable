package pgtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The harness must run the exact image the local stack runs (tag and digest).
func TestImage_MatchesTheLocalComposePin(t *testing.T) {
	root, err := repoRoot()
	require.NoError(t, err)
	compose, err := os.ReadFile(filepath.Join(root, "infra", "local", "compose.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(compose), "image: "+Image+"\n")
}

func TestQuoteIdent(t *testing.T) {
	tests := []struct{ in, want string }{
		{"notes", `"notes"`},
		{"Mixed Case", `"Mixed Case"`},
		{`we"ird`, `"we""ird"`},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, QuoteIdent(tc.in))
		})
	}
}

func TestRepoRoot_FindsTheLocalStack(t *testing.T) {
	root, err := repoRoot()
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, bootstrapRolesSQL))
	assert.NoError(t, err)
}
