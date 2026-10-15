package annual_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// strategyFile holds the section strategies and the registry's fixed order:
// the only place in this package that may name a computed section kind.
const strategyFile = "sections.go"

// computedKindNames are the computed section kinds of the domain
// (adomain.SectionPF*, adomain.SectionPM*).
var computedKindNames = map[string]bool{
	"SectionPFArt152": true, "SectionPFRESICOCheck": true, "SectionPM601": true, "SectionPMRESICO": true,
}

// ANN-24 gate (OCP per section): a new section is a strategy plus one
// registration line. The assembler must read what a section needs from the
// ComputedSection its strategy returned, never branch on the section kind;
// otherwise a new section compiles, runs, and silently gets another
// section's reasons, notes or line regimes.
func TestOnlyTheStrategiesNameAComputedSectionKind(t *testing.T) {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)
	fset := token.NewFileSet()
	var offenders []string
	for _, name := range files {
		if name == strategyFile || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		require.NoError(t, err)
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if ok && computedKindNames[sel.Sel.Name] {
				offenders = append(offenders, fset.Position(sel.Pos()).String()+" "+sel.Sel.Name)
			}
			return true
		})
	}
	require.Empty(t, offenders, "move the per-kind fact onto ComputedSection, set by the strategy")
}
