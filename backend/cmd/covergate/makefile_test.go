package main

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// coverExcludeLine finds the default COVER_EXCLUDE in backend/Makefile.
var coverExcludeLine = regexp.MustCompile(`(?m)^COVER_EXCLUDE\s*\?=\s*(.+)$`)

// PR #13 follow-up 3 (P0-30): the gate's default exclusion must leave out only
// the generated server and the top-level commands. An unanchored "/cmd/" also
// dropped any business package that merely had a cmd/ directory in its path.
func TestMakefileCoverExcludeIsAnchoredToTheModule(t *testing.T) {
	makefile, err := os.ReadFile("../../Makefile")
	require.NoError(t, err)
	match := coverExcludeLine.FindSubmatch(makefile)
	require.NotNil(t, match, "COVER_EXCLUDE ?= ... not found in backend/Makefile")
	exclude, err := regexp.Compile(string(match[1]))
	require.NoError(t, err)

	tests := []struct {
		file     string
		excluded bool
	}{
		{modPrefix + "cmd/api/main.go", true},
		{modPrefix + "cmd/covergate/gate.go", true},
		{modPrefix + "adapter/http/gen/api.gen.go", true},
		{modPrefix + "internal/ledger/cmd/post.go", false},
		{modPrefix + "internal/platform/worker/cmd/run.go", false},
		{modPrefix + "adapter/http/router/gen/x.go", false},
		{modPrefix + "internal/platform/crypto/aead.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			assert.Equal(t, tt.excluded, exclude.MatchString(tt.file))
		})
	}
}
