package golden_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

const scope = "internal/taxengine/"

func src(body string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(body)} }

func TestFindRewritersFlagsEveryWayToRegenerateGoldens(t *testing.T) {
	fsys := fstest.MapFS{
		"cmd/tool/main.go": src(`package main
import "flag"
var update = flag.Bool("update", false, "rewrite goldens")
var spec = flag.String("spec", "", "fine: not an update knob")
`),
		"internal/x/aliased_test.go": src(`package x
import f "flag"
var regen bool
func init() { f.BoolVar(&regen, "regen-golden", false, "") }
`),
		"internal/x/env_test.go": src(`package x
import "os"
var accept = os.Getenv("ACCEPT_GOLDEN") != ""
var home = os.Getenv("HOME")
`),
		"internal/taxengine/golden/write_test.go": src(`package golden
import "os"
func save() { _ = os.WriteFile("testdata/golden/x.yaml", nil, 0o600) }
func move() { _ = os.Rename("a", "b") }
`),
		"internal/other/write.go": src(`package other
import "os"
func save() { _ = os.WriteFile("out.txt", nil, 0o600) }
`),
		"internal/taxengine/testdata/ignored.go": src(`package bad
import "flag"
var u = flag.Bool("update", false, "")
`),
		"internal/taxengine/notes.md": src("flag.Bool(\"update\")"),
	}

	got, err := golden.FindRewriters(fsys, scope)

	require.NoError(t, err)
	var all []string
	for _, v := range got {
		all = append(all, v.String())
	}
	assert.Equal(t, []string{
		`cmd/tool/main.go:3: flag "update" can rewrite goldens`,
		`internal/taxengine/golden/write_test.go:3: os.WriteFile writes files under internal/taxengine/`,
		`internal/taxengine/golden/write_test.go:4: os.Rename writes files under internal/taxengine/`,
		`internal/x/aliased_test.go:4: flag "regen-golden" can rewrite goldens`,
		`internal/x/env_test.go:3: environment variable "ACCEPT_GOLDEN" can rewrite goldens`,
	}, all)
}

func TestFindRewritersReportsUnparsableSource(t *testing.T) {
	_, err := golden.FindRewriters(fstest.MapFS{"a.go": src("package")}, scope)

	require.Error(t, err)
}
