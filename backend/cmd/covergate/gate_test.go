package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const modPrefix = "github.com/romodevx/control-contable/backend/"

var defaultExclude = regexp.MustCompile(`/adapter/http/gen/|/cmd/`)

func TestMeasure(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		wantCovered int64
		wantTotal   int64
	}{
		{
			name: "counts statements of covered and uncovered blocks",
			profile: "mode: atomic\n" +
				modPrefix + "internal/a/a.go:1.1,2.2 3 1\n" +
				modPrefix + "internal/a/a.go:3.1,4.2 1 0\n",
			wantCovered: 3,
			wantTotal:   4,
		},
		{
			name: "excludes generated code and commands",
			profile: "mode: set\n" +
				modPrefix + "internal/a/a.go:1.1,2.2 2 1\n" +
				modPrefix + "adapter/http/gen/api.gen.go:1.1,2.2 100 0\n" +
				modPrefix + "cmd/tool/main.go:1.1,2.2 50 0\n",
			wantCovered: 2,
			wantTotal:   2,
		},
		{
			name: "a block reported twice counts once and is covered if any run hit it",
			profile: "mode: atomic\n" +
				modPrefix + "internal/a/a.go:1.1,2.2 4 0\n" +
				modPrefix + "internal/a/a.go:1.1,2.2 4 7\n",
			wantCovered: 4,
			wantTotal:   4,
		},
		{
			name:        "ignores blank lines",
			profile:     "mode: set\n\n" + modPrefix + "internal/a/a.go:1.1,2.2 1 1\n\n",
			wantCovered: 1,
			wantTotal:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Measure(strings.NewReader(tt.profile), defaultExclude)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCovered, got.Covered)
			assert.Equal(t, tt.wantTotal, got.Total)
		})
	}
}

func TestMeasureRejectsMalformedProfiles(t *testing.T) {
	tests := []struct {
		name    string
		profile string
	}{
		{name: "empty input", profile: ""},
		{name: "missing mode header", profile: modPrefix + "internal/a/a.go:1.1,2.2 1 1\n"},
		{name: "too few fields", profile: "mode: set\n" + modPrefix + "internal/a/a.go:1.1,2.2 1\n"},
		{name: "non-numeric statements", profile: "mode: set\n" + modPrefix + "internal/a/a.go:1.1,2.2 x 1\n"},
		{name: "non-numeric count", profile: "mode: set\n" + modPrefix + "internal/a/a.go:1.1,2.2 1 y\n"},
		{name: "every block excluded", profile: "mode: set\n" + modPrefix + "cmd/t/main.go:1.1,2.2 1 1\n"},
		{name: "negative statements", profile: "mode: set\n" + modPrefix + "internal/a/a.go:1.1,2.2 -100 0\n" + modPrefix + "internal/a/b.go:1.1,2.2 1 1\n"},
		{name: "negative count", profile: "mode: set\n" + modPrefix + "internal/a/a.go:1.1,2.2 1 -1\n"},
		{name: "no statements to measure", profile: "mode: set\n" + modPrefix + "internal/a/a.go:1.1,2.2 0 0\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Measure(strings.NewReader(tt.profile), defaultExclude)
			assert.Error(t, err)
		})
	}
}

func TestCoverageMeets(t *testing.T) {
	tests := []struct {
		name     string
		coverage Coverage
		minPct   int64
		want     bool
	}{
		{name: "exactly at the threshold passes", coverage: Coverage{Covered: 80, Total: 100}, minPct: 80, want: true},
		{name: "above the threshold passes", coverage: Coverage{Covered: 9, Total: 10}, minPct: 80, want: true},
		{name: "just below the threshold fails", coverage: Coverage{Covered: 799, Total: 1000}, minPct: 80, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.coverage.Meets(tt.minPct))
		})
	}
}

func TestCoverageString(t *testing.T) {
	assert.Equal(t, "79.9% (799/1000 statements)", Coverage{Covered: 799, Total: 1000}.String())
	assert.Equal(t, "100.0% (3/3 statements)", Coverage{Covered: 3, Total: 3}.String())
}
