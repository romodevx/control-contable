package taxengine_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// annualMarker names the annual sub-context's packages (ADR-0009):
// domain/annual, usecase/annual, infra/annualinputs, infra/annuallimits,
// adapter/http/annual and the annualgolden harness.
const annualMarker = "annual"

// TestMonthlyEngineNeverImportsTheAnnual is ADR-0009 rule 1: the
// dependency is one-way (annual → monthly), so deleting the annual
// directories, their cmd/ wiring and taxengine.AnnualSections removes the
// module without touching M5. Only this module root package, the
// composition list, may name both.
func TestMonthlyEngineNeverImportsTheAnnual(t *testing.T) {
	checked := 0
	err := filepath.WalkDir(".", func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			return err
		}
		dir := filepath.ToSlash(filepath.Dir(file))
		if dir == "." || isAnnual(dir) {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		checked++
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			assert.Falsef(t, strings.Contains(p, "/taxengine/") && isAnnual(p), "%s imports %s: the monthly engine must not depend on the annual (ADR-0009)", file, p)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Positive(t, checked, "the scan is not vacuous")
}

func isAnnual(p string) bool {
	for _, part := range strings.Split(path.Clean(p), "/") {
		if strings.HasPrefix(part, annualMarker) {
			return true
		}
	}
	return false
}
